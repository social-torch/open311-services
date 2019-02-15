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
		if req.Resource == "/city/{id}" {
			id := req.PathParameters["id"]
			return getCity(id)
		}

		if req.Resource == "/cities" {
			return getCities()
		}
	}
	return clientError(http.StatusMethodNotAllowed)
}

func getCity(id string) (events.APIGatewayProxyResponse, error) {
	city, err := repository.GetCity(id)
	if err != nil {
		switch err.(type) {
		case *repository.CityNotFoundErr:
			errorMessage := fmt.Sprintf("city handler error: \n %s \n  city_name: %s not in repository", err, id)
			errorLogger.Println(errorMessage)
			return events.APIGatewayProxyResponse{Body: errorMessage, StatusCode: 404}, nil
		default:
			return serverError(err)
		}
	}

	body, err := json.Marshal(&city)
	if err != nil {
		//TODO throw server error here instead
		return serverError(fmt.Errorf("city handler: unable to marshal service: \n %+v \n %s", city, err))
	}

	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
}

func getCities() (events.APIGatewayProxyResponse, error) {
	cities, err := repository.GetCities()
	if err != nil {
		return serverError(err)
	}

	body, err := json.Marshal(cities)
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
