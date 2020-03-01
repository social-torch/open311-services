package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var infoLogger = log.New(os.Stdout, "INFO\t", 0)
var warningLogger = log.New(os.Stderr, "WARNING\t", log.Lshortfile)
var errorLogger = log.New(os.Stderr, "ERROR\t", log.Lshortfile)

// Route requests
func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		if req.Resource == "/images/fetch/{key}" {
			key := req.PathParameters["key"]
			return getPresignedURLForFetch(key)
		}

		if req.Resource == "/images/store/{key}" {
			key := req.PathParameters["key"]
			return getPresignedURLForStore(key)
		}
	}
	return clientError(http.StatusMethodNotAllowed, errors.New("Method must be 'GET'"))

}

// Get presigned S3 URL to retrieve an image
func getPresignedURLForFetch(key string) (events.APIGatewayProxyResponse, error) {
	bucket := os.Getenv("IMAGE_BUCKET")
	svc := s3.New(session.New())
	req, _ := svc.GetObjectRequest( &s3.GetObjectInput {
		Bucket: aws.String(bucket),
		Key: aws.String(key) } )

	urlStr, err := req.Presign(10 * time.Minute)
	if err != nil {
		errorLogger.Println(err)
		return serverError(http.StatusInternalServerError, errors.New("Error retreiving presigned S3 URL for retrieving"))
	}

	infoLogger.Println("Presigned URL  ", urlStr)
	body, _ := json.Marshal( &struct {
																			 URL      string  `json:"url"`
																		 }{
																			 URL: urlStr,
																		 })

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"content-type": "application/json", "Access-Control-Allow-Origin": "*"},
		Body:       string(body),
	}, nil
}

// Get presigned S3 URL to store an image
func getPresignedURLForStore(key string) (events.APIGatewayProxyResponse, error) {
	bucket := os.Getenv("IMAGE_BUCKET")
	svc := s3.New(session.New())
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key) } )

	urlStr, err := req.Presign(10 * time.Minute)
	if err != nil {
		errorLogger.Println(err)
		return serverError(http.StatusInternalServerError, errors.New("Error retreiving presigned S3 URL for storing"))
	}

	infoLogger.Println("Presigned URL  ", urlStr)
	body, _ := json.Marshal( &struct {
																			 URL      string  `json:"url"`
																		 }{
																			 URL: urlStr,
																		 })

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
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
