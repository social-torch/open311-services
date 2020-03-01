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

/// Route requests
func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		if req.Resource == "/request/{id}" {
			id := req.PathParameters["id"]
			return getRequest(id)
		}

		if req.Resource == "/requests" {
			return getRequests()
		}

	case "POST":
		return submitRequest(req)
	}
	return clientError(http.StatusMethodNotAllowed, errors.New("method must be 'GET' or 'POST'"))
}

func getRequest(id string) (events.APIGatewayProxyResponse, error) {
	request, err := repository.GetRequest(id)
	if err != nil {
		switch err.(type) {
		case *repository.RequestIdNotFoundErr:
			errorMessage := fmt.Errorf("%s. service_request_id '%s' not in database", err, id)
			return clientError(http.StatusNotFound, errorMessage)
		default:
			return serverError(http.StatusInternalServerError, err)
		}
	}

	body, err := json.Marshal(&request)
	if err != nil {
		return serverError(http.StatusInternalServerError, errors.New("error marshalling GetRequest() struct"))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"content-type": "application/json", "Access-Control-Allow-Origin": "*"},
		Body:       string(body),
	}, nil
}

func getRequests() (events.APIGatewayProxyResponse, error) {
	requests, err := repository.GetRequests()
	if err != nil {
		return serverError(http.StatusInternalServerError, err)
	}

	body, err := json.Marshal(requests)
	if err != nil {
		return serverError(http.StatusInternalServerError, errors.New("error marshalling GetRequests() struct"))
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"content-type": "application/json", "Access-Control-Allow-Origin": "*"},
		Body:       string(body),
	}, nil
}

func submitRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	userID := req.Headers["from"] // accountID must be added to header in client app
	if userID == "" {             // but just in case the client app doesn't, track request as a guest
		userID = "guest"
	}

	var Open311request repository.Request
	err := json.Unmarshal([]byte(req.Body), &Open311request)
	if err != nil {
		return clientError(http.StatusUnprocessableEntity, errors.New("error unmarshalling Request JSON. Check syntax"))
	}

	// Make sure Request has minimum amount of information in order to create new 311 request
	// Check that service code exists in Services table
	if !repository.IsValidServiceCode(Open311request.ServiceCode) {
		return clientError(http.StatusBadRequest, errors.New("invalid Service Code: "+Open311request.ServiceCode))
	}

	// Check that request has a location
	if Open311request.Address == "" && (Open311request.Latitude == 0 && Open311request.Longitude == 0) {
		return clientError(http.StatusBadRequest, errors.New("no location included in request"))
	}

	var response repository.RequestResponse
	// If this is a new request, initialize a new request.  If this is an existing request, update it
	if Open311request.ServiceRequestID == "" {
		// Create new Open311 Request and load into DynamoDB Requests table
		response, err = repository.SubmitRequest(Open311request, userID)
		infoLogger.Println("New request submitted: " + response.ServiceRequestID)
	} else {
		// Update existing Open311 Request in DynamoDB Requests table
		response, err = repository.UpdateRequest(Open311request, userID)
		infoLogger.Println("Request updated: " + response.ServiceRequestID)
	}

	if err != nil {
		return serverError(http.StatusInternalServerError, err)
	}

	body, err := json.Marshal(response)
	if err != nil {
		return serverError(http.StatusInternalServerError, errors.New("unable to marshal JSON for request response"))
	}

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
