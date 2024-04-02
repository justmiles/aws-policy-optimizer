package optimizer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gigawattio/awsarn"

	"github.com/flosell/iam-policy-json-to-terraform/converter"
	"github.com/micahhausler/aws-iam-policy/policy"
)

type GenerateOptimizedPolicyOptions struct {
	Database           string
	Table              string
	QueryResultsBucket string
	QueryResultsPrefix string
	AthenaWorkgroup    string
	UserIdentityARN    string
	AccountID          string
	Region             string
	OutputFormat       string
	AnalysisPeriod     int
	CombinePrefixes    bool
}

func GenerateOptimizedPolicy(options GenerateOptimizedPolicyOptions) (string, error) {

	start := time.Now().AddDate(0, 0, options.AnalysisPeriod*-1)

	sql := fmt.Sprintf(`
	SELECT DISTINCT
		useridentity.arn as useridentity,
		CONCAT(SPLIT_PART(eventsource, '.', 1),':',eventname) as permission,
		resource.arn as resource
	FROM "%s"."%s"
	CROSS JOIN UNNEST(resources) AS t(resource)
	WHERE day > '%s'
	AND regexp_like(useridentity.arn, '%s')
	AND account_id = '%s'
	AND region = '%s'
	AND NULLIF(errorcode, '') IS NULL
	`, options.Database, options.Table, start.Format("2006/01/02"), options.UserIdentityARN, options.AccountID, options.Region)

	var usageHistory []UsageHistoryRecord
	err := QueryAthena(sql, options.Database, options.QueryResultsBucket, options.QueryResultsPrefix, options.AthenaWorkgroup, &usageHistory)
	if err != nil {
		return "", err
	}

	// generate the permissions map map[identity]map[permission]resource
	var permissionMap = make(map[string]map[string][]string)
	for _, record := range usageHistory {
		if _, ok := permissionMap[record.UserIdenityArn]; ok {
			permissionMap[record.UserIdenityArn][record.Permission] = append(permissionMap[record.UserIdenityArn][record.Permission], record.ResourceArn)

			if _, ok := permissionMap[record.UserIdenityArn][record.Permission]; ok {
				permissionMap[record.UserIdenityArn][record.Permission] = append(permissionMap[record.UserIdenityArn][record.Permission], record.ResourceArn)
			} else {
				permissionMap[record.UserIdenityArn][record.Permission] = []string{record.ResourceArn}
			}

		} else {
			permissionMap[record.UserIdenityArn] = make(map[string][]string)
			permissionMap[record.UserIdenityArn][record.Permission] = []string{record.ResourceArn}
		}
	}

	// Deduplicate the permissions -> Resource map
	// Build the final IAM Policy
	var statements = []policy.Statement{}
	for identity, permissionSet := range permissionMap {
		for action, resources := range permissionSet {
			var consolidatedResources []string
			if options.CombinePrefixes {
				consolidatedResources = consolidatePrefixes(resources)
			} else {
				var err error
				consolidatedResources, err = consolidateARNs(resources)
				if err != nil {
					return "", err
				}
			}

			actions := []string{action}

			// deduplicate policies
			for dupeAction, dupeResources := range permissionSet {
				var dupeConsolidatedResources []string
				var err error
				if options.CombinePrefixes {
					dupeConsolidatedResources = consolidatePrefixes(dupeResources)
				} else {
					dupeConsolidatedResources, err = consolidateARNs(dupeResources)
					if err != nil {
						return "", err
					}
				}
				if dupeAction == action {
					continue
				}
				if reflect.DeepEqual(consolidatedResources, dupeConsolidatedResources) {
					actions = append(actions, dupeAction)
					delete(permissionMap[identity], dupeAction)
				}
			}

			statements = append(statements, policy.Statement{
				Effect:   policy.EffectAllow,
				// Principal: policy.NewServicePrincipal("cloudtrail.amazonaws.com"), // TODO: consider getting the principal
				Action:   policy.NewStringOrSlice(false, actions...),
				Resource: policy.NewStringOrSlice(false, consolidatedResources...),
			})
		}
	}

	p := policy.Policy{
		Version:    policy.VersionLatest,
		Id:         "GenIAMPolicy", // TODO: better ID
		Statements: policy.NewStatementOrSlice(statements...),
	}

	out, _ := json.MarshalIndent(p, "", "\t")

	if options.OutputFormat == "hcl" {
		return converter.Convert("GenIAMPolicy", out)

	} else {
		return string(out), nil
	}
}

func consolidatePrefixes(resources []string) []string {
	prefixMap := make(map[string]struct{})
	for _, resource := range resources {
		parts := strings.Split(resource, "/")
		for i := range parts {
			prefix := strings.Join(parts[:i+1], "/")
			prefixMap[prefix] = struct{}{}
		}
	}

	consolidated := make([]string, 0, len(prefixMap))
	for prefix := range prefixMap {
		consolidated = append(consolidated, prefix+"/*")
	}

	return consolidated
}

func consolidateARNs(arns []string) ([]string, error) {
	var arnMap = make(map[string][]string)
	for _, arn := range arns {
		if arn == "" {
			continue
		}
		components, err := awsarn.Parse(arn)
		if err != nil {
			return nil, err
		}
		resource := components.Resource
		components.Resource = ""
		if val, ok := arnMap[components.String()]; ok {
			arnMap[components.String()] = append(val, resource)
		} else {
			arnMap[components.String()] = []string{resource}
		}
	}

	var ss []string
	for arn, resources := range arnMap {
		globbedArn, _ := awsarn.Parse(arn)
		globbedArn.Resource = generateGlobPattern(resources)
		ss = append(ss, globbedArn.String())
	}

	return ss, nil
}

func generateGlobPattern(ss []string) string {
	if len(ss) == 0 {
		return ""
	}

	parts := strings.Split(ss[0], "/")
	for i := 1; i < len(parts); i++ {
		for _, s := range ss {
			if !strings.HasPrefix(s, strings.Join(parts[:i+1], "/")) {
				return strings.Join(parts[:i], "/") + "/*"
			}
		}
	}

	return strings.Join(parts, "/")
}

type UsageHistoryRecord struct {
	UserIdenityArn string `csv:"useridentity"`
	Permission     string `csv:"permission"`
	ResourceArn    string `csv:"resource"`
}
