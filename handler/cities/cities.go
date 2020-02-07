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

// Route requests
func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		if req.Resource == "/city/{id}" {
			id := req.PathParameters["id"]
			return getCity(id)
		}

		if req.Resource == "/cities" {
			return getCities()
		}

	case "POST":
		if req.Resource == "/city/onboard" {
			return submitRequest(req)
		}
	}
	return clientError(http.StatusMethodNotAllowed, errors.New("method must be 'GET'"))

}

func getCity(id string) (events.APIGatewayProxyResponse, error) {
	city, err := repository.GetCity(id)
	if err != nil {
		switch err.(type) {
		case *repository.CityNotFoundErr:
			errorMessage := fmt.Errorf("%s.  city_name '%s' not in database", err, id)
			return clientError(http.StatusNotFound, errorMessage)
		default:
			return serverError(http.StatusInternalServerError, err)
		}
	}

	body, err := json.Marshal(&city)
	if err != nil {
		return serverError(http.StatusInternalServerError, errors.New("error marshalling GetCity() struct"))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       string(body),
	}, nil
}

func getCities() (events.APIGatewayProxyResponse, error) {
	cities, err := repository.GetCities()
	if err != nil {
		return serverError(http.StatusInternalServerError, err)
	}

	body, err := json.Marshal(cities)
	if err != nil {
		return serverError(http.StatusInternalServerError, errors.New("error marshalling GetCities() struct"))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
//		Headers:    map[string]string{"content-type": "application/json", "Access-Control-Allow-Origin": "*", "Access-Control-Allow-Headers":"Content-Type"},
		Body:       string(body),
	}, nil
}

func submitRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := req.Headers["from"] // accountID must be added to header in client app
	if userID == "" {             // but just in case the client app doesn't, track request as a guest
		userID = "guest"
	}

	var onboardingRequest repository.OnboardingRequest
	err := json.Unmarshal([]byte(req.Body), &onboardingRequest)
	if err != nil {
		return clientError(http.StatusUnprocessableEntity, errors.New("error unmarshalling onboarding request JSON. Check syntax"))
	}

	// Make sure minimum amount of information in order to create onboarding request
	if onboardingRequest.City == "" && onboardingRequest.State =="" {
		return clientError(http.StatusBadRequest, errors.New("City and State must be specified"))
	}

	// Create onboarding request and load into DynamoDB table
	response, err := repository.AddOnboardingRequest(onboardingRequest, userID)
	if err != nil {
		return serverError(http.StatusInternalServerError, err)
	}

	body, err := json.Marshal(response)
	if err != nil {
		return serverError(http.StatusInternalServerError, errors.New("unable to marshal JSON for response"))
	}

	infoLogger.Println("New onboarding request submitted")

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers:    map[string]string{"content-type": "application/json"},
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
