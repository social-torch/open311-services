include .env

clean:
	@rm -rf dist
	@mkdir -p dist

build: clean
	@for dir in `ls handler`; do \
		GOOS=linux go build -o dist/handler/$$dir github.com/social-torch/open311-services/handler/$$dir; \
	done

run:
	sam local start-api

install:
	go get github.com/aws/aws-sdk-go
	go get github.com/aws/aws-lambda-go/events
	go get github.com/aws/aws-lambda-go/lambda
	go get github.com/oklog/ulid
	go get github.com/stretchr/testify/assert

test:
	go test ./... --cover

configure:
	aws s3api create-bucket \
		--bucket $(AWS_BUCKET_NAME) \
		--region $(AWS_REGION)

package: build
	@aws cloudformation package \
		--template-file template.yml \
		--s3-bucket $(AWS_BUCKET_NAME) \
		--region $(AWS_REGION) \
		--output-template-file package.yml

deploy:
	@aws cloudformation deploy \
		--template-file package.yml \
		--region $(AWS_REGION) \
		--capabilities CAPABILITY_IAM \
		--stack-name $(AWS_STACK_NAME)

describe:
	@aws cloudformation describe-stacks \
		--region $(AWS_REGION) \
		--stack-name $(AWS_STACK_NAME) \

outputs:
	@aws cloudformation describe-stacks \
		--region $(AWS_REGION) \
		--stack-name $(AWS_STACK_NAME) | jq -r '.Stacks[0].Outputs'

url:
	@aws cloudformation describe-stacks \
		--region $(AWS_REGION) \
		--stack-name $(AWS_STACK_NAME)| jq -r '.Stacks[0].Outputs[0].OutputValue' -j
