package main

import (
//	"encoding/json"
//	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
//	"github.com/social-torch/open311-services/repository"
)

var infoLogger = log.New(os.Stdout, "INFO\t", 0)
var warningLogger = log.New(os.Stderr, "WARNING\t", log.Lshortfile)
var errorLogger = log.New(os.Stderr, "ERROR\t", log.Lshortfile)

func AddConfirmedUser(req events.CognitoEventUserPoolsPostConfirmationRequest ) (events.CognitoEventUserPoolsPostConfirmationResponse , error) {
		infoLogger.Println(fmt.Sprintf("User confirmed \n %v", req))
		return events.CognitoEventUserPoolsPostConfirmationResponse{}, nil
}

func main() {
	lambda.Start(AddConfirmedUser)
}
