package main

import (
	//	"encoding/json"
	//	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/social-torch/open311-services/repository"
)

var infoLogger = log.New(os.Stdout, "INFO\t", 0)
var warningLogger = log.New(os.Stderr, "WARNING\t", log.Lshortfile)
var errorLogger = log.New(os.Stderr, "ERROR\t", log.Lshortfile)

func addConfirmedUser(req events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	infoLogger.Println(fmt.Sprintf("User confirmed \n %v", req))

  for key, value := range req.Request.UserAttributes { // Order not specified 
		infoLogger.Println(fmt.Sprintf("Key:[%s] - [%s]", key, value))
  }
  // Not sure why but the user key is in "sub"
  // An example:
/*
RAW JSON
{{1 PostConfirmation_ConfirmSignUp us-east-1 us-east-1_uUDDwJWv9 {aws-sdk-unknown-unknown 6nk51q1bnv50onrhgcq4lv2l7n} d1a29c7f-c5c3-47db-8771-c0d9857592e0} {map[given_name:z family_name:z email:xxx@gmail.com sub:d1a29c7f-c5c3-47db-8771-c0d9857592e0 cognito:email_alias:xxx@gmail.com cognito:user_status:CONFIRMED email_verified:true]} {}}
Just USerAttributes parsed
INFO given_name z
INFO family_name z
INFO email xxx@gmail.com
INFO sub d1a29c7f-c5c3-47db-8771-c0d9857592e0
INFO cognito:email_alias xxx@gmail.com
INFO cognito:user_status CONFIRMED
INFO email_verified true
*/
  infoLogger.Println(fmt.Sprintf("User sub: %s", req.Request.UserAttributes["sub"]))
  accountID := req.Request.UserAttributes["sub"]
	err := repository.AddNewUser(accountID)

	if err != nil {
		return req, err
	}

	return req, nil
}

func main() {
	lambda.Start(addConfirmedUser)
}
