# Infrastructure

`cloudformation.yaml` defines the AWS infrastructure for `card-onboarding-services`:

- `onboard-service`
- `customer-management-service`
- `account-management-service`
- `onboard-service-request-status-${EnvironmentName}`
- `onboard-service-account-details-${EnvironmentName}`

The template runs the services on **Amazon ECS/Fargate**, exposes them through a shared public Application Load Balancer, and uses DynamoDB on-demand tables for persistence. Service images are passed in as private ECR image URI parameters.

## Required Parameters

- `EnvironmentName`
- `VpcId`
- `PublicSubnetIds`
- `OnboardServiceImageUri`
- `CustomerManagementServiceImageUri`
- `AccountManagementServiceImageUri`

Optional parameters include service ports, `DesiredCount`, `TaskCpu`, `TaskMemory`, and `DownstreamTimeout`.

## ECS Runtime

The template creates:

- an ECS cluster
- one Fargate task definition and ECS service per application service
- one shared internet-facing Application Load Balancer
- one listener and target group per service port
- one CloudWatch log group for service logs
- security groups for the load balancer and Fargate tasks
- an ECS task execution role for pulling ECR images and writing logs
- an onboard-service task role for DynamoDB access

The default public URLs are:

- `http://<alb-dns>:8080` for `onboard-service`
- `http://<alb-dns>:8081` for `customer-management-service`
- `http://<alb-dns>:8082` for `account-management-service`

## Onboard Service Configuration

The deployed `onboard-service` receives:

- `CUSTOMER_MANAGEMENT_BASE_URL`
- `ACCOUNT_MANAGEMENT_BASE_URL`
- `REQUEST_STATUS_TABLE_NAME`
- `ACCOUNT_DETAILS_TABLE_NAME`
- `DOWNSTREAM_TIMEOUT`

Its ECS task role can read and write both DynamoDB tables.

## Deployment

The pipeline deploys production from the `main` branch with:

```sh
make deploy-production
```

That target:

- builds all three Docker images
- creates the ECR repositories if they do not exist
- logs in to ECR
- tags and pushes all three images with `IMAGE_TAG`
- deploys `infra/cloudformation.yaml` with the pushed image URIs

The CI environment must provide AWS credentials with permission to manage ECR, CloudFormation, IAM, ECS, ELB, CloudWatch Logs, EC2 security groups, and DynamoDB.

Required CI variables:

- `AWS_ACCOUNT_ID`
- `AWS_REGION`
- `VPC_ID`
- `PUBLIC_SUBNET_IDS`

Optional variables:

- `ENVIRONMENT_NAME`, default `prod`
- `STACK_NAME`, default `card-onboarding-services-${ENVIRONMENT_NAME}`
- `IMAGE_TAG`, default current git short SHA

`PUBLIC_SUBNET_IDS` should be a comma-separated list, for example `subnet-aaaaaaaa,subnet-bbbbbbbb`.

The deployment Make targets validate that the configured VPC and subnets exist in `AWS_REGION` before starting CloudFormation. If a first deploy fails during stack creation, CloudFormation can leave the stack in `ROLLBACK_COMPLETE`; that state cannot be updated. Delete the failed stack record, fix the VPC/subnet CI variables, then rerun the deployment:

```sh
aws cloudformation delete-stack --region ap-southeast-1 --stack-name card-onboarding-services-prod
aws cloudformation wait stack-delete-complete --region ap-southeast-1 --stack-name card-onboarding-services-prod
```

You can also run the same deployment manually:

Example:

```powershell
$env:AWS_ACCOUNT_ID = "123456789012"
$env:AWS_REGION = "ap-southeast-1"
$env:VPC_ID = "vpc-xxxxxxxx"
$env:PUBLIC_SUBNET_IDS = "subnet-aaaaaaaa,subnet-bbbbbbbb"
make deploy-production
```

## Local Mode

Local in-memory mode is controlled by the Go service configuration, not by this template. When `REQUEST_STATUS_TABLE_NAME` and `ACCOUNT_DETAILS_TABLE_NAME` are not set, `onboard-service` continues to use the in-memory stores.
