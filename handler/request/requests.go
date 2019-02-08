package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/social-torch/open311-services/repository"
	"log"
	"net/http"
	"os"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

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
		//body := req.body
		//Parse
		//body
		return submitRequest(req)
	}
	return clientError(http.StatusMethodNotAllowed)
}

func getRequest(id string) (events.APIGatewayProxyResponse, error) {
	request, _ := (repository.GetRequest(id)) //TODO use value mechanics instead of pointer mechanics
	body, err := json.Marshal(request)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Unable to marshal JSON", StatusCode: 500}, nil
	}
	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
}

func getRequests() (events.APIGatewayProxyResponse, error) {
	requests, _ := repository.GetRequests() //TODO use value mechanics instead of pointer mechanics
	body, err := json.Marshal(requests)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Unable to marshal JSON", StatusCode: 500}, nil
	}
	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
}

func submitRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if req.Headers["Content-Type"] != "application/json" {
		return clientError(http.StatusNotAcceptable)
	}

	var Open311request repository.Request
	err := json.Unmarshal([]byte(req.Body), &Open311request)
	if err != nil {
		return clientError(http.StatusUnprocessableEntity)
	}

	// Make sure Request has minimum amount of information in order to create new 311 request
	//Check that service code exists in Services table
	if !repository.IsValidServiceCode(Open311request.ServiceCode) {
		return clientError(http.StatusBadRequest)
	}

	//Check that request has a location
	if Open311request.Address == "" && (Open311request.Latitude == 0 && Open311request.Longitude == 0) {
		return clientError(http.StatusBadRequest)
	}

	response, err := repository.SubmitRequest(Open311request)

	if err != nil {
		return serverError(err)
	}
	
	body, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Unable to marshal JSON", StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       string(body),
	}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func main() {
	lambda.Start(router)
}
