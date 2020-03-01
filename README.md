# Open311 Services

This repos contains the back end implementation of the Social Torch Open311 Services.  The YAML file specifies the AWS cloud formation template, which spins up the necessary infrastructure to use API Gateway, Cognito, S3 buckets, DynamoDB, and corresponding lambda functions.

See this nice blog by [Alex Edwards](https://www.alexedwards.net/) for more details on hooking up the AWS bits: [How to build a Serverless API with Go and AWS Lambda](https://www.alexedwards.net/blog/serverless-api-with-go-and-aws-lambda)

## Get Started

Ensure your GOPATH is set in your environment. Once that is set, the `make` commands below will work.

```bash
$ > mkdir -p $GOPATH/src/github.com/social-torch
$ > cd  $GOPATH/src/github.com/social-torch
$ > git clone git@github.com:social-torch/open311-services
```

## Dependencies

```bash
$ > sudo yum install jq or sudo apt-get install jq
# Install Go dependencies
$ > make install
```
This installs:

$ > go get github.com/aws/aws-lambda-go/events
$ > go get github.com/aws/aws-lambda-go/lambda
$ > go get github.com/oklog/ulid
$ > go get github.com/stretchr/testify/assert


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

Ensure your [AWS credentials](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) are set up properly for the account to which you wish to deploy. Create a `~/.aws/credentials` file. Go to AWS IAM (Identity and access management). Go to Access management, users, (your user), Summary, then the Security credentials tab and pick the Access keys. IF you've lost the secret key, create a new Access Key ID and you'll get the secret key (once and only once)

### Create .env

```bash
AWS_ACCOUNT_ID=1234567890
AWS_BUCKET_NAME=your-bucket-name-for-cloudformation-package-data
AWS_REGION=us-west-1
AWS_STACK_NAME=your-cloudformation-stack-name
AWS_STAGE=Prod
AWS_USER_POOL=your-cognito-pool-ARN
AWS_IMAGE_BUCKET_NAME=name-of-bucket-to-store-mobile-image-uploads
```

The `AWS_ACCOUNT_ID` can be found by logging into AWS, Clicking on your name and pulling down "My Account". Your Account ID should be the first item under Account Settings.
The `AWS_BUCKET_NAME` is your S3 bucket name, acessible via The Services... S3. 
`AWS_REGION` is the region your services are deployed to, typically us-east-1
`AWS_STACK_NAME` is the name of your CloudFormation stack, (AWS services, search for Cloud Formation and find the stack you want to use)
`AWS_USER_POOL` is found by searching for the Cognito Service, then manage your User Pool. Select the pool that comes up and the ARN will be under Pool ARN
`AWS_IMAGE_BUCKET_NAME` is the S3 bucket your images will go into

### Command
The .env files should be placed in your open311-services directory (next to the Makefile).

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

## Cognito App Client

When creating the AppClient ID, do NOT create a secret key.

## API Gateway OPTIONS Resource

It seems to also be required that the default OPTIONS HTTP method that is generated from the SAM Template needs to be unauthorized.  Info here: [https://github.com/awslabs/aws-sam-cli/issues/1381](https://github.com/awslabs/aws-sam-cli/issues/1381)

1. Go to the API Gatway console
2. Find "Resources", select the OPTIONS method on the resource/URL you'd like to alter
3. In the Method Request block, change Authorization to NONE.

Not sure if the preflight check for CORS needs to be unauthorized. AWS Amplify might not correctly form the request header for OPTIONS, because I see it place the authorization header as a VALUE of the Access-Control-Request-Headers header, which makes no sense.

## Logging access to the API Gateway:

For debugging purposes, follow this simple 22 step process to see a log of your API calls: [https://aws.amazon.com/premiumsupport/knowledge-center/api-gateway-cloudwatch-logs/](https://aws.amazon.com/premiumsupport/knowledge-center/api-gateway-cloudwatch-logs/)

Can also use AWS X-Ray to see invocations through the API - Got to API Gateway, under API, select "Stages", the Log/Tracing tab, then enable Xray checkbox.

In X-ray, you can go to "Traces" and select your API you want to inspect in the text box at he top. 

## API Gateway notes

Remember, any changes to the API will require youto Deploy the app to the proper stage. This was also required for changing Authorizers.

## Security Note

Until we automate it in the YAML, you must manually add a security policy for the CitiesRole, RequestRole, UsersRole and ServicesRole to access DynamoDB. You must also attach a policy for the ImagesRole to access the appropriate S3 images bucket.

When accessing the cloud API, your request will need an authorization token.
