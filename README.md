# Open311 Services

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
Also depends on docker, the AWS CLI and AWS SAM CLI

## Build
```bash
# Build binary
$ > make build

# Test Go Code
$ > make test
```

## Deploy

### Create .env

```bash
AWS_ACCOUNT_ID=1234567890
AWS_BUCKET_NAME=your-bucket-name-for-cloudformation-package-data
AWS_STACK_NAME=your-cloudformation-stack-name
AWS_REGION=us-west-1
```

### Install AWS CLI

```bash
$ > brew install awscli
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
