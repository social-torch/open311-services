package repository

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/sony/sonyflake"
	"strconv"
	"time"
)

const (
	ServicesTable = "Services"
	RequestsTable = "Requests"
	AwsRegion     = endpoints.UsEast1RegionID // "us-east-1" // US East (N. Virginia).
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
	ServiceRequestId  string  `json:"service_request_id"` // The unique ID of the service request created.
	Status            string  `json:"status"`             // The current status of the service request.
	StatusNotes       string  `json:"status_notes"`       // Explanation of why status was changed to current state or more details on current status than conveyed with status alone.
	ServiceName       string  `json:"service_name"`       // The human readable name of the service request type
	ServiceCode       string  `json:"service_code"`       // The unique identifier for the service request type
	Description       string  `json:"description"`        // A full description of the request or report submitted.
	AgencyResponsible string  `json:"agency_responsible"` // The agency responsible for fulfilling or otherwise addressing the service request.
	ServiceNotice     string  `json:"service_notice"`     // Information about the action expected to fulfill the request or otherwise address the information reported.
	RequestedDateTime string  `json:"requested_datetime"` // The date and time when the service request was made.
	UpdatedDateTime   string  `json:"update_datetime"`    // The date and time when the service request was last modified. For requests with status=closed, this will be the date the request was closed.
	ExpectedDateTime  string  `json:"expected_datetime"`  // The date and time when the service request can be expected to be fulfilled. This may be based on a service-specific service level agreement.
	Address           string  `json:"address"`            // Human readable address or description of location.
	AddressId         string  `json:"address_id"`         // The internal address ID used by a jurisdictions master address repository or other addressing system.
	ZipCode           int32   `json:"zipcode"`            // The postal code for the location of the service request.
	Latitude          float32 `json:"lat"`                // latitude using the (WGS84) projection.
	Longitude         float32 `json:"lon"`                // longitude using the (WGS84) projection.
	MediaUrl          string  `json:"media_url"`          // A URL to media associated with the request, eg an image.
	//Values              []AttributeValue `json:"values"`  //TODO enable this to grow with the things Aussie wants
}

type RequestResponse struct {
	ServiceRequestID string `json:"service_request_id"` // The unique ID of the service request created.
	ServiceNotice    string `json:"service_notice"`     // Information about the action expected to fulfill the request or otherwise address the information reported
	AccountID        string `json:"account_id"`         // Unique ID for the user account of the person submitting the request
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
	result, err := svc.Scan(params)
	if err != nil {
		return nil, fmt.Errorf("\n repository: unable to get all services from database with the following parameters: %+v. \n  %s", params, err)
	}

	services := []Service{}

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

	if request.ServiceRequestId == "" {
		return Request{}, &RequestIdNotFoundErr{"request not found"}
	}

	return request, err
}

func SubmitRequest(request Request) (RequestResponse, error) {
	svc, err := createDynamoClient()
	if err != nil {
		return RequestResponse{}, err
	}

	// Get unique identifier by which this new request will be submitted.
	requestID, err := genRequestID()
	if err != nil {
		return RequestResponse{}, fmt.Errorf("\nrepository: failed to generate unique id for new request. \n  %s", err)
	}
	request.ServiceRequestId = requestID

	// Assign requested_datetime
	t := time.Now()
	request.RequestedDateTime = t.Format(time.RFC3339)

	//Initialize new request as "open"
	request.Status = "open"

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

	var response RequestResponse
	response.ServiceRequestID = requestID

	return response, err
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
		fmt.Printf("\nERROR: repository/IsValidServiceCode: unable to establish session with AWS \n  %s", err)
		return false //TODO better handle errors

	}

	filt := expression.Contains(expression.Name("service_code"), ServiceCode)
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		fmt.Printf("\nERROR: repository: "+
			"While checking if service existed, Got error building database expression. \n   %s", err)
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(ServicesTable),
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Printf("\nERROR: repository: "+
			"Query API call failed while checking if Service Code was valid. \n   %s", err)
	}

	if *result.Count == 1 { //Since "Contains" matches substrings and services codes should be unique, valid IDs match only once
		return true
	}

	return false
}

func genRequestID() (string, error) {
	// TODO Replace with something a little more readable. Make assumptions on performance required.
	// Perhaps return something based on calendar day of request and an incrementing number (atomically stored in DB??

	// Initialize sonyfake - see usage details at https://github.com/sony/sonyflake
	var st sonyflake.Settings
	//st.MachineID = awsutil.AmazonEC2MachineID //uncomment if running in EC2  // look here for timeout errors
	flake := sonyflake.NewSonyflake(st)

	if flake == nil {
		return "", fmt.Errorf("repository/sonyflake: error creating Unique ID generator. \n  sonyflake settings %+v", st)
	}

	// Generate relatively short unique ID
	id, err := flake.NextID()
	if err != nil {
		return "", fmt.Errorf("repository/sonyflake: error generating Unique ID. \n  sonyflake: %+v \n  %s", flake, err)
	}
	idString := strconv.FormatUint(id, 16) // sonyflake uses Hex
	return idString, err
}
