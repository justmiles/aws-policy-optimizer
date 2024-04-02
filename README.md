# AWS Policy Optimizer

The AWS Policy Optimizer is a tool that analyzes AWS CloudTrail Access Logs and generates least-privilege IAM policies based on utilization. It aims to help optimize resource access by identifying the specific permissions needed for each resource.

## Usage

Once you have built the application, use the following command to generate an optimized policy:

```bash
aws-policy-optimizer [flags]
```

### Flags

The AWS Policy Optimizer supports the following flags:

- `--account-id`: (required) limit analysis to events in this AWS account
- `--analysis-period`: how far back into the access records to look (default 90)
- `--athena-workgroup`: run analysis in this Athena workgroup (default "primary")
- `--database`: database name for Athena CloudTrail Table (default "default")
- `--query-results-bucket`: (optional) S3 bucket for Athena query results
- `--query-results-prefix`: (optional) S3 bucket for Athena query prefix
- `--region`: (required) limit analysis to events in this region
- `--table`: table name for Athena CloudTrail Table (default "cloudtrail")
- `--iam-role`: (required) the IAM name to lookup, or the entire sessionIssuer arn
- `--output-format`: (defaults json) json or hcl

## Example

Here's a couple example commands that generate an optimized policy:

```bash
aws-policy-optimizer --iam-role arn:aws:iam::123456789012:user/my-user --account-id 123456789012 --region us-west-2
```

```bash
# useful in cases of ECS where task arns change for assumed roles, but takes longer
aws-policy-optimizer --iam-role my-role-name --account-id 123456789012 --region us-east-1 --output-format hcl > my-policy.tf

```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please feel free to open an issue or submit a pull request in the [GitHub repository](https://github.com/justmiles/aws-policy-optimizer).

## License

This project is licensed under the Mozilla Public License. For more information, please refer to the [LICENSE](LICENSE) file.
