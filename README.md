# AWS Policy Optimizer

The AWS Policy Optimizer is a tool that analyzes AWS CloudTrail Access Logs and generates least-privilege IAM policies based on utilization. It aims to help optimize resource access by identifying the specific permissions needed for each resource.

## Usage

Once you have built the application, use the following command to generate an optimized policy:

```bash
aws-policy-optimizer [flags]
```

### Flags

The AWS Policy Optimizer supports the following flags:

- `--account-id`: limit analysis to events in this AWS account
- `--analysis-period`: how far back into the access records to look (default 90)
- `--athena-workgroup`: run analysis in this Athena workgroup (default "primary")
- `--database`: database name for Athena CloudTrail Table (default "default")
- `--query-results-bucket`: (optional) S3 bucket for Athena query results
- `--query-results-prefix`: (optional) S3 bucket for Athena query prefix
- `--region`: limit analysis to events in this region
- `--table`: table name for Athena CloudTrail Table (default "cloudtrail")
- `--user-identity-arn`: the ARN of the target resource

## Example

Here's an example command that generates an optimized policy:

```bash
aws-policy-optimizer --user-identity-arn arn:aws:iam::123456789012:user/my-user --account-id 123456789012 --region us-west-2
```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please feel free to open an issue or submit a pull request in the [GitHub repository](https://github.com/justmiles/aws-policy-optimizer).

## License

This project is licensed under the Mozilla Public License. For more information, please refer to the [LICENSE](LICENSE) file.
