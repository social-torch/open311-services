package repository

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oklog/ulid"
)

const (
	ServicesTable = "Services"
	RequestsTable = "Requests"
	CitiesTable   = "Cities"
	UserTable     = "Users"
	AwsRegion     = endpoints.UsEast1RegionID // "us-east-1" // US East (N. Virginia).
)
const (
	RequestOpen   = "open"   // Open311 Request Status - it has been reported
	RequestClosed = "closed" // Open311 Request Status - it has been resolved
)

// Single service (type) offered via Open311
type Service struct {
	ServiceCode string   `json:"service_code"`
	ServiceName string   `json:"service_name"`
	Description string   `json:"description"`
	Metadata    bool     `json:"metadata"`
	Type        string   `json:"type"`
	Keywords    []string `json:"keywords"`
	Group       string   `json:"group"`
}

// Service definition associated with a service code.
// These attributes can be unique to the city/jurisdiction
type ServiceDefinition struct {
	ServiceCode string             `json:"service_code"`
	Attributes  []ServiceAttribute `json:"attributes"`
}

// Single attribute extension for a service
type ServiceAttribute struct {
	Code                string           `json:"code"`
	DataType            string           `json:"datatype"`
	Variable            bool             `json:"variable"`
	Required            bool             `json:"required"`
	Order               int32            `json:"order"`
	Description         string           `json:"description"`
	DataTypeDescription string           `json:"datatype_description"`
	Values              []AttributeValue `json:"values"`
}

// Possible value for ServiceAttribute that defines lists
type AttributeValue struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Issues that have been reported as service requests.  Location is submitted via lat/long or address
type Request struct {
	ServiceRequestID  string           `json:"service_request_id"` // The unique ID of the service request created.
	Status            string           `json:"status"`             // The current status of the service request.
	StatusNotes       string           `json:"status_notes"`       // Explanation of why status was changed to current state or more details on current status than conveyed with status alone.
	ServiceName       string           `json:"service_name"`       // The human readable name of the service request type
	ServiceCode       string           `json:"service_code"`       // The unique identifier for the service request type
	Description       string           `json:"description"`        // A full description of the request or report submitted.
	AgencyResponsible string           `json:"agency_responsible"` // The agency responsible for fulfilling or otherwise addressing the service request.
	ServiceNotice     string           `json:"service_notice"`     // Information about the action expected to fulfill the request or otherwise address the information reported.
	RequestedDateTime string           `json:"requested_datetime"` // The date and time when the service request was made.
	UpdatedDateTime   string           `json:"update_datetime"`    // The date and time when the service request was last modified. For requests with status=closed, this will be the date the request was closed.
	ExpectedDateTime  string           `json:"expected_datetime"`  // The date and time when the service request can be expected to be fulfilled. This may be based on a service-specific service level agreement.
	Address           string           `json:"address"`            // Human readable address or description of location.
	AddressID         string           `json:"address_id"`         // The internal address ID used by a jurisdictions master address repository or other addressing system.
	ZipCode           int32            `json:"zipcode"`            // The postal code for the location of the service request.
	Latitude          float32          `json:"lat"`                // latitude using the (WGS84) projection.
	Longitude         float32          `json:"lon"`                // longitude using the (WGS84) projection.
	MediaURL          string           `json:"media_url"`          // A URL to media associated with the request, eg an image.
	Values            []AttributeValue `json:"values"`             // Enables future expansion
}

type RequestResponse struct {
	ServiceRequestID string `json:"service_request_id"` // The unique ID of the service request created.
	ServiceNotice    string `json:"service_notice"`     // Information about the action expected to fulfill the request or otherwise address the information reported
	AccountID        string `json:"account_id"`         // Unique ID for the user account of the person submitting the request
}

type User struct {
	UserID            string    `json:"user_id"`            // Unique ID of Open311 User
	SubmittedRequests []Request `json:"submitted_requests"` // Slice of requests user has made
	WatchedRequests   []Request `json:"watched_requests"`   // Slice of request user is watching
}

// Assumes each jurisdiction has its own AWS endpoint
type City struct {
	CityName string `json:"city_name"`
	Endpoint string `json:"endpoint"`
}

type ServiceCodeNotFoundErr struct {
	message string
}

func (e *ServiceCodeNotFoundErr) Error() string {
	return e.message
}

type RequestIdNotFoundErr struct {
	message string
}

func (e *RequestIdNotFoundErr) Error() string {
	return e.message
}

type CityNotFoundErr struct {
	message string
}

func (e *CityNotFoundErr) Error() string {
	return e.message
}

// GetServices returns array of all Open311 Services in DynamoBD Service Table
func GetServices() ([]Service, error) {
	return allServices()
}

func allServices() ([]Service, error) {
	svc, err := createDynamoClient()
	if err != nil {
		return []Service{}, err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String(ServicesTable),
	}

	// Make the DynamoDB Query API call
	// TODO handle pagination
	result, err := svc.Scan(params)
	if err != nil {
		return nil, fmt.Errorf("\n repository: unable to get all services from database with the following parameters: %+v. \n  %s", params, err)
	}

	services := []Service{}

	// TODO - investigate UnmarshalListOfMaps here
	// For each service, unmarshal and add to array of services
	for _, i := range result.Items {
		service := Service{}
		err = dynamodbattribute.UnmarshalMap(i, &service)
		if err != nil {
			return services, fmt.Errorf("\n repository: Failed to unmarshal record: \n %+v \n   %s", i, err)
		}

		services = append(services, service)
	}
	return services, err
}

// GetService takes a service code UUID, looks up that service in DynamoDB and returns the corresponding
// Open311 Service struct.  If the requested service code is not in the database, a ServiceCodeNotFoundErr error is set
func GetService(code string) (Service, error) {
	svc, err := createDynamoClient()
	if err != nil {
		return Service{}, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(ServicesTable),
		Key: map[string]*dynamodb.AttributeValue{
			"service_code": {
				S: aws.String(code),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return Service{}, fmt.Errorf("\n repository: unable to get specified service from database with the following input: \n  %+v. \n   %s", input, err)
	}

	service := Service{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &service)
	if err != nil {
		return service, fmt.Errorf("\n repository: Failed to unmarshal service record from database: \n  %+v. \n   %s", result.Item, err)
	}

	if service.ServiceCode == "" {
		return service, &ServiceCodeNotFoundErr{"service not found"}
	}

	return service, err
}

// GetRequests returns array of all Open311 Requests in DynamoBD Requests Table
func GetRequests() ([]Request, error) {
	return allRequests()
}

func allRequests() ([]Request, error) {
	svc, err := createDynamoClient()
	if err != nil {
		return []Request{}, err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String(RequestsTable),
	}

	// Make the DynamoDB Query API call
	// TODO handle pagination
	result, err := svc.Scan(params)
	if err != nil {
		return nil, fmt.Errorf("repository: unable to get all requests from database with the following parameters: %+v. \n %s", params, err)
	}

	requests := []Request{}

	// for each request, unmarshal and add to array of all requests
	for _, i := range result.Items {
		request := Request{}
		err = dynamodbattribute.UnmarshalMap(i, &request)
		if err != nil {
			return requests, fmt.Errorf("repository: Failed to unmarshal record: %+v. \n %s", i, err)
		}

		requests = append(requests, request)
	}
	return requests, err
}

// GetRequest takes a service_request_id, looks up that request in DynamoDB and returns the corresponding
// Open311 Request struct.  If the service_request_id is not in the database, a RequestIdNotFoundErr error is set
func GetRequest(id string) (Request, error) {
	svc, err := createDynamoClient()
	if err != nil {
		return Request{}, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(RequestsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"service_request_id": {
				S: aws.String(id),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return Request{}, fmt.Errorf("repository: unable to get specified request from database with the following input: %+v \n %s", input, err)
	}

	request := Request{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &request)
	if err != nil {
		return request, fmt.Errorf("repository: Failed to unmarshal request record from database: %+v. \n %s", result.Item, err)
	}

	if request.ServiceRequestID == "" {
		return Request{}, &RequestIdNotFoundErr{"request not found"}
	}

	return request, err
}

func SubmitRequest(request Request, userID string) (RequestResponse, error) {
	svc, err := createDynamoClient()
	if err != nil {
		return RequestResponse{}, err
	}

	// Get unique identifier by which this new request will be submitted.
	requestID, err := genRequestID()
	if err != nil {
		return RequestResponse{}, fmt.Errorf("\nrepository: failed to generate unique id for new request. \n  %s", err)
	}
	request.ServiceRequestID = requestID

	// Assign requested_datetime
	t := time.Now()
	request.RequestedDateTime = t.Format(time.RFC3339)

	//Initialize new request as "open"
	request.Status = RequestOpen

	// Initialize service name
	service, _ := GetService(request.ServiceCode)
	request.ServiceName = service.ServiceName

	av, err := dynamodbattribute.MarshalMap(request)
	if err != nil {
		return RequestResponse{}, fmt.Errorf("repository: Failed to marshal request:\n %+v. \n  %s", request, err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(RequestsTable),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return RequestResponse{}, fmt.Errorf("repository: failed to put new request in database: \n input: %+v. \n %s", input, err)
	}

	// Add new request to list of requests created by this user
	_, err = trackUserRequest(userID, requestID)
	if err != nil {
		//log error
	}

	var response RequestResponse
	response.AccountID = userID
	response.ServiceRequestID = requestID

	return response, err
}

// trackUserRequest updates the Users table to append a request to the list of requsts a user has created
func trackUserRequest(userID string, requestID string) (*dynamodb.UpdateItemOutput, error) {
	// TODO implement funcion
	// see: https://stackoverflow.com/questions/31288085/how-to-append-a-value-to-list-attribute-on-aws-dynamodb
	// handle cases where userID is "Guest" or nil
	return nil, nil
}

// createDynamoClient is a convenience function to establish a session with AWS and
// returns a new instance of the DynamoDB client
func createDynamoClient() (*dynamodb.DynamoDB, error) {

	// Initial credentials loaded from SDK's default credential chain. Such as
	// the environment, shared credentials (~/.aws/credentials), or EC2 Instance
	// Role.

	// Create the session that the DynamoDB service will use.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AwsRegion)},
	)
	if err != nil {
		return nil, fmt.Errorf("\n repository: unable to establish session with AWS \n  %s", err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	return svc, nil
}

func IsValidServiceCode(ServiceCode string) bool {
	svc, err := createDynamoClient()
	if err != nil {
		// TODO send this to os.Stderr so the AWS cloudwatch logs pick it up
		fmt.Printf("\nERROR: repository/IsValidServiceCode: unable to establish session with AWS \n  %s", err)
		return false
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(ServicesTable),
		Key: map[string]*dynamodb.AttributeValue{
			"service_code": {
				S: aws.String(ServiceCode),
			},
		},
	}
	response, err := svc.GetItem(input)
	if err != nil {
		// TODO send this to os.Stderr so the AWS cloudwatch logs pick it up
		fmt.Printf("\nERROR: repository: "+
			"Query API call failed while checking if Service Code was valid. \n   %s", err)
	}

	// If there is no matching item, GetItem does not return any data and there will be no Item element in the response.
	if response.Item == nil {
		return false
	}

	return true
}

func genRequestID() (string, error) {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	id, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		return "", fmt.Errorf("\n repository: Unable to generate request id:\n  %s", err)
	}
	reqID := "SR-" + id.String()
	return reqID, nil
}

func GetCities() ([]City, error) {
	return allCities()
}

func allCities() ([]City, error) {
	svc, err := createDynamoClient()
	if err != nil {
		return []City{}, err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String(CitiesTable),
	}

	// Make the DynamoDB Query API call
	// TODO handle pagination
	result, err := svc.Scan(params)
	if err != nil {
		return nil, fmt.Errorf("\n repository: unable to get all cities from database with the following parameters: %+v. \n  %s", params, err)
	}

	cities := []City{}

	// For each city, unmarshal and add to array of cities
	for _, i := range result.Items {
		city := City{}
		err = dynamodbattribute.UnmarshalMap(i, &city)
		if err != nil {
			return cities, fmt.Errorf("\n repository: Failed to unmarshal record: \n %+v \n   %s", i, err)
		}

		cities = append(cities, city)
	}
	return cities, err
}

func GetCity(id string) (City, error) {
	svc, err := createDynamoClient()
	if err != nil {
		return City{}, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(CitiesTable),
		Key: map[string]*dynamodb.AttributeValue{
			"city_name": {
				S: aws.String(id),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return City{}, fmt.Errorf("\n repository: unable to get specified city from database with the following input: \n  %+v. \n   %s", input, err)
	}

	city := City{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &city)
	if err != nil {
		return city, fmt.Errorf("\n repository: Failed to unmarshal city record from database: \n  %+v. \n   %s", result.Item, err)
	}

	if city.CityName == "" {
		return city, &CityNotFoundErr{"city not found"}
	}

	return city, err
}
