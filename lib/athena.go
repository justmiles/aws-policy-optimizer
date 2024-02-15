package optimizer

import (
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/athena/athenaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/gocarina/gocsv"
)

var (
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc        athenaiface.AthenaAPI
	downloader s3manageriface.DownloaderAPI
)

func init() {
	svc = athena.New(sess)
	downloader = s3manager.NewDownloader(sess)
}

// Execute a SQL query against Athena
func QueryAthena(sql, database, queryResultsBucket, queryResultsPrefix, workgroup string, out *[]UsageHistoryRecord) error {

	startQueryExecutionInput := athena.StartQueryExecutionInput{
		QueryString: &sql,
		QueryExecutionContext: &athena.QueryExecutionContext{
			Database: &database,
		},
		ResultConfiguration: &athena.ResultConfiguration{},
		ResultReuseConfiguration: &athena.ResultReuseConfiguration{
			ResultReuseByAgeConfiguration: &athena.ResultReuseByAgeConfiguration{
				Enabled:         aws.Bool(true),
				MaxAgeInMinutes: aws.Int64(10080),
			},
		},
	}

	if queryResultsBucket != "" {
		startQueryExecutionInput.ResultConfiguration.OutputLocation = aws.String("s3://" + path.Join(queryResultsBucket, queryResultsPrefix))
	}

	if workgroup != "" {
		startQueryExecutionInput.WorkGroup = &workgroup
	}

	result, err := svc.StartQueryExecution(&startQueryExecutionInput)
	if err != nil {
		return err
	}

	queryExecutionInput := athena.GetQueryExecutionInput{
		QueryExecutionId: result.QueryExecutionId,
	}

	var qrop *athena.GetQueryExecutionOutput

	// Wait until query finishes
	for {
		qrop, err = svc.GetQueryExecution(&queryExecutionInput)
		if err != nil {
			return err
		}

		if *qrop.QueryExecution.Status.State == athena.QueryExecutionStateSucceeded || *qrop.QueryExecution.Status.State == athena.QueryExecutionStateFailed || *qrop.QueryExecution.Status.State == athena.QueryExecutionStateCancelled {
			break
		}

	}

	if *qrop.QueryExecution.Status.State == "SUCCEEDED" {

		file, err := os.CreateTemp("", "athena-query-results-"+*result.QueryExecutionId)
		if err != nil {
			return fmt.Errorf("unable to create temp file %q, %v", *result.QueryExecutionId, err)
		}
		defer os.Remove(file.Name())

		u, _ := url.Parse(*qrop.QueryExecution.ResultConfiguration.OutputLocation)

		_, err = downloader.Download(file, &s3.GetObjectInput{
			Bucket: &u.Host,
			Key:    &u.Path,
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					return fmt.Errorf("unable to download query results for %q. Bucket %s does not exist", *result.QueryExecutionId, u.Host)
				case s3.ErrCodeNoSuchKey:
					return fmt.Errorf("unable to download query results for %q, %v: results do not exist!", *result.QueryExecutionId, err)
				default:
					return fmt.Errorf("unable to download query results for %q, %v", *result.QueryExecutionId, err)
				}
			}
		}

		if err := gocsv.UnmarshalFile(file, out); err != nil {
			return fmt.Errorf("unable to unmarshal csv: %v", err)
		}

		return nil
	}

	return fmt.Errorf("query state: %s\n\t%s", *qrop.QueryExecution.Status.State, *qrop.QueryExecution.Status.StateChangeReason)
}
