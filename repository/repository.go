package repository

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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
	ServiceCode string             `json:service_code`
	Attributes  []ServiceAttribute `json:attributes`
}

// Single attribute extension for a service
type ServiceAttribute struct {
	Code                string           `json:code`
	DataType            string           `json:datatype`
	Variable            bool             `json:variable`
	Required            bool             `json:required`
	Order               int32            `json:order`
	Description         string           `json:description`
	DataTypeDescription string           `json:datatype_description`
	Values              []AttributeValue `json:values`
}

// Possible value for ServiceAttribute that defines lists
type AttributeValue struct {
	Key  string `json:key`
	name string `json:name`
}

// Issues that have been reported as service requests.  Location
// is submitted via lat/long or address
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
}

func allServices() ([]Service, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hardcode region
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String("Services"), //TODO don't hardcode Tablename
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	services := []Service{}

	// For each service, unmarshall and add to array of services
	for _, i := range result.Items {
		service := Service{}
		err = dynamodbattribute.UnmarshalMap(i, &service)

		if err != nil {
			panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}

		services = append(services, service)
	}
	return services, err
}

// GetRequest returns service information with code
func GetService(code string) (*Service, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hardcode region
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Services"), //TODO don't hardcode tablename ?
		Key: map[string]*dynamodb.AttributeValue{
			"service_code": {
				S: aws.String(code),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	service := Service{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &service)

	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}

	if service.ServiceCode == "" {
		fmt.Println("Could not find Service")
		return nil, errors.New("Service not found")
	}

	return &service, err
}

// GetServices returns all Services
func GetServices() ([]Service, error) {
	return allServices()
}

func allRequests() ([]Request, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hardcode region
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String("Requests"), //TODO don't hardcode region
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	requests := []Request{}

	// for each request, unmarshall and add to array of all requests
	for _, i := range result.Items {
		request := Request{}
		err = dynamodbattribute.UnmarshalMap(i, &request)

		if err != nil {
			panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}

		requests = append(requests, request)
	}
	return requests, err
}

// GetRequest returns request with id
func GetRequest(id string) (*Request, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hardcode region
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Requests"), //TODO don't hardcode ?
		Key: map[string]*dynamodb.AttributeValue{
			"service_request_id": {
				S: aws.String(id),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	request := Request{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &request)

	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}

	if request.ServiceRequestId == "" {
		fmt.Println("Could not find Request")
		return nil, errors.New("Request not found")
	}

	return &request, err
}

// GetRequests returns all Requests
func GetRequests() ([]Request, error) {
	return allRequests()
}
