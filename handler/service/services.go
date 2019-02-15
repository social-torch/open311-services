package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/social-torch/open311-services/repository"
	"log"
	"net/http"
	"os"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

// Route requests
func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		if req.Resource == "/service/{id}" {
			id := req.PathParameters["id"]
			return getService(id)
		}

		if req.Resource == "/services" {
			return getServices()
		}
	}
	return clientError(http.StatusMethodNotAllowed)
}

func getService(id string) (events.APIGatewayProxyResponse, error) {
	service, err := repository.GetService(id)
	if err != nil {
		switch err.(type) {
		case *repository.ServiceCodeNotFoundErr:
			errorMessage := fmt.Sprintf("service handler error: \n %s \n  service_code: %s not in repository", err, id)
			errorLogger.Println(errorMessage)
			return events.APIGatewayProxyResponse{Body: errorMessage, StatusCode: 404}, nil
		default:
			return serverError(err)
		}
	}

	body, err := json.Marshal(&service)
	if err != nil {
		//TODO throw server error here instead
		return serverError(fmt.Errorf("service handler: unable to marshal service: \n %+v \n %s", service, err))
	}

	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
}

func getServices() (events.APIGatewayProxyResponse, error) {
	services, err := repository.GetServices()
	if err != nil {
		return serverError(err)
	}

	body, err := json.Marshal(services)
	if err != nil {
		//TODO throw server error here instead
		return events.APIGatewayProxyResponse{Body: "Unable to marshal JSON", StatusCode: 500}, nil
	}
	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())
	//TODO  Need to provide more context to the user based on error
	// See https://wiki.open311.org/GeoReport_v2/#errors
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	//TODO provide more context
	// See https://wiki.open311.org/GeoReport_v2/#errors
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func main() {
	lambda.Start(router)
}
