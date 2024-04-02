package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	optimizer "github.com/justmiles/aws-policy-optimizer/lib"
)

var (
	options optimizer.GenerateOptimizedPolicyOptions
)

func init() {

	rootCmd.PersistentFlags().StringVar(&options.Database, "database", "default", "database name for Athena CloudTrail Table")
	rootCmd.PersistentFlags().StringVar(&options.Table, "table", "cloudtrail", "table name for Athena CloudTrail Table")
	rootCmd.PersistentFlags().StringVar(&options.UserIdentityARN, "user-identity-arn", "", "(required) the whole or partial ARN of the target resource")
	rootCmd.PersistentFlags().StringVar(&options.AccountID, "account-id", "", "(required) limit analysis to events in this AWS account")
	rootCmd.PersistentFlags().StringVar(&options.Region, "region", "", "(required) limit analysis to events in this region")
	rootCmd.PersistentFlags().StringVar(&options.AthenaWorkgroup, "athena-workgroup", "primary", "run analysis in this Athena workgroup")
	rootCmd.PersistentFlags().StringVar(&options.QueryResultsBucket, "query-results-bucket", "", "(optional) S3 bucket for Athena query results")
	rootCmd.PersistentFlags().StringVar(&options.QueryResultsPrefix, "query-results-prefix", "", "(optional) S3 bucket for Athena query prefix")
	rootCmd.PersistentFlags().StringVar(&options.OutputFormat, "output-format", "json", "json or hcl")
	rootCmd.PersistentFlags().IntVar(&options.AnalysisPeriod, "analysis-period", 90, "how far back into the access records to look")
	rootCmd.PersistentFlags().BoolVar(&options.CombinePrefixes, "shrink-policy", false, "whether or not to combine statements by api call prefix")

	rootCmd.MarkPersistentFlagRequired("user-identity-arn")
	rootCmd.MarkPersistentFlagRequired("account-id")
	rootCmd.MarkPersistentFlagRequired("region")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aws-policy-optimizer",
	Short: "analyze AWS CloudTrail Access Logs and generate least-privilege IAM policies based on utilization",
	Run: func(rootCmd *cobra.Command, args []string) {

		if options.OutputFormat != "json" && options.OutputFormat != "hcl" {
			log.Fatalf("invalid output format '%s'", options.OutputFormat)
		}
		policy, err := optimizer.GenerateOptimizedPolicy(options)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(policy)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s" .Version}}
`)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
