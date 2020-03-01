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
    if req.Resource == "/service/{id}" {
      id := req.PathParameters["id"]
      return getService(id)
    }

    if req.Resource == "/services" {
      return getServices()
    }
  }
  return clientError(http.StatusMethodNotAllowed, errors.New("method must be 'GET'"))
}

func getService(id string) (events.APIGatewayProxyResponse, error) {
  service, err := repository.GetService(id)
  if err != nil {
    switch err.(type) {
    case *repository.ServiceCodeNotFoundErr:
      errorMessage := fmt.Errorf("%s. service_code '%s' not in database", err, id)
      return clientError(http.StatusNotFound, errorMessage)
    default:
      return serverError(http.StatusInternalServerError, err)
    }
  }

  body, err := json.Marshal(&service)
  if err != nil {
    return serverError(http.StatusInternalServerError, errors.New("error marshalling GetService() struct"))
  }

  return events.APIGatewayProxyResponse{
    StatusCode: http.StatusOK,
    Headers:    map[string]string{"content-type": "application/json"},
    Body:       string(body),
  }, nil
}

func getServices() (events.APIGatewayProxyResponse, error) {
  services, err := repository.GetServices()
  if err != nil {
    return serverError(http.StatusInternalServerError, err)
  }

  body, err := json.Marshal(services)
  if err != nil {
    return serverError(http.StatusInternalServerError, errors.New("error marshalling GetServices() struct"))
  }

  return events.APIGatewayProxyResponse{
    StatusCode: http.StatusOK,
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
