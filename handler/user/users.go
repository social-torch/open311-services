package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/social-torch/open311-services/repository"
)

var infoLogger = log.New(os.Stdout, "INFO\t", 0)
var warningLogger = log.New(os.Stderr, "WARNING\t", log.Lshortfile)
var errorLogger = log.New(os.Stderr, "ERROR\t", log.Lshortfile)

/// Route request
func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		if req.Resource == "/user/{id}" {
			id := req.PathParameters["id"]
			return getUser(id)
		}
	case "POST":
		if req.Resource == "/feedback" {
			return submitFeedback(req)
		}
	}
	return clientError(http.StatusMethodNotAllowed, errors.New("method must be 'GET' or 'POST'"))
}

func getUser(accountID string) (events.APIGatewayProxyResponse, error) {
	user, err := repository.GetUser(accountID)
	if err != nil {
		switch err.(type) {
		case *repository.AccountIDNotFoundErr:
			errorMessage := fmt.Errorf("%s. account_id: '%s' not in database", err, accountID)
			return clientError(http.StatusNotFound, errorMessage)
		default:
			return serverError(http.StatusInternalServerError, err)
		}
	}

	body, err := json.Marshal(&user)
	if err != nil {
		return serverError(http.StatusInternalServerError, errors.New("error marshalling User struct"))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"content-type": "application/json", "Access-Control-Allow-Origin": "*"},
		Body:       string(body),
	}, nil
}

func submitFeedback(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var feedback repository.Feedback
	err := json.Unmarshal([]byte(req.Body), &feedback)
	if err != nil {
		return clientError(http.StatusUnprocessableEntity, errors.New("error unmarshalling feedback JSON. Check syntax"))
	}

	// Load feedback into DynamoDB table
	response, err := repository.AddFeedback(feedback)
	if err != nil {
		return serverError(http.StatusInternalServerError, err)
	}

	body, err := json.Marshal(response)
	if err != nil {
		return serverError(http.StatusInternalServerError, errors.New("unable to marshal JSON for response"))
	}

	infoLogger.Println("Feedback submitted")

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers:    map[string]string{"content-type": "application/json", "Access-Control-Allow-Origin": "*"},
		Body:       string(body),
	}, nil
}

func serverError(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"content-type": "text/plain"},
		Body:       http.StatusText(statusCode) + ": " + err.Error(),
	}, nil
}

func clientError(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	warningLogger.Println(err.Error())
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"content-type": "text/plain"},
		Body:       http.StatusText(statusCode) + ": " + err.Error(),
	}, nil
}

func main() {
	lambda.Start(router)
}
