# Open311 Services

This repos contains the back end implementation of the Social Torch Open311 Services.  The YAML file specifies the AWS cloud formation template, which spins up the necessary infrastructure to use API Gateway, Cognito, S3 buckets, DynamoDB, and corresponding lambda functions.

See this nice blog by [Alex Edwards](https://www.alexedwards.net/) for more details on hooking up the AWS bits: [How to build a Serverless API with Go and AWS Lambda](https://www.alexedwards.net/blog/serverless-api-with-go-and-aws-lambda)

## Get Started

```bash
$ > mkdir -p $GOPATH/src/github.com/social-torch
$ > cd  $GOPATH/src/github.com/social-torch
$ > git clone git@github.com:social-torch/open311-services
```

## Dependencies

```bash
$ > go get github.com/aws/aws-lambda-go/events
$ > go get github.com/aws/aws-lambda-go/lambda
$ > go get github.com/stretchr/testify/assert
$ > sudo yum install jq
```

Also depends on [docker](https://docs.docker.com/v17.09/engine/installation/), the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-linux.html) and the [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)

## Build

```bash
# Build binary
$ > make build
```

## Test

```bash
# Test Go Code
$ > make test
```

AWS provides [SAM Local](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/sam-cli-command-reference-sam-local-start-api.html) to run serverless applications locally for quick development and testing.

```bash
# Start a local, containerized instantiation of endpoints and lambda functions
$ > make run
```

## Deploy to Cloud

Ensure your [AWS credentials](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) are set up properly for the account to which you wish to deploy.

### Create .env

```bash
AWS_ACCOUNT_ID=1234567890
AWS_BUCKET_NAME=your-bucket-name-for-cloudformation-package-data
AWS_STACK_NAME=your-cloudformation-stack-name
AWS_REGION=us-west-1
AWS_STACK_NAME=your-cloudformation-stack-name
AWS_STAGE=stage-to-deploy-to(dev,qa,prod)
AWS_USER_POOL=arn-of-cognito-userpool
AWS_IMAGE_BUCKET_NAME=name-of-bucket-to-store-mobile-image-uploads
```

### Command

```bash
# Create S3 Bucket
$ > make configure

# Upload data to S3 Bucket
$ > make package

# Deploy CloudFormation Stack
$ > make deploy
```

## Usage

```bash
$ > make outputs

[
  {
    "OutputKey": "URL",
    "OutputValue": "https://random-id.execute-api.us-west-1.amazonaws.com/Prod",
    "Description": "URL for HTTPS Endpoint"
  }
]

$ > curl https://random-id.execute-api.us-west-1.amazonaws.com/Stage/services

TODO:  Show all calls

$ > curl https://random-id.execute-api.us-west-1.amazonaws.com/Stage/requests

TODO:  Show all calls
```

## Security Note

Until we automate it in the YAML, you must manually add a security policy for the CitiesRole, RequestRole, UsersRole and ServicesRole to access DynamoDB. You must also attach a policy for the ImagesRole to access the appropriate S3 images bucket.

When accessing the cloud API, your request will need an authorization token.