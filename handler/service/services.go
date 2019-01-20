package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/social-torch/open311-services/repository"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

// Route requests
func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    switch req.HTTPMethod {
    case "GET":
        return getServices()
    default:
        return clientError(http.StatusMethodNotAllowed)
    }
}

func getServices() (events.APIGatewayProxyResponse, error) {
	services := repository.GetServices()
	body, err := json.Marshal(services)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Unable to marshal JSON", StatusCode: 500}, nil
	}
	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
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
