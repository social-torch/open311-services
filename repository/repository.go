package repository

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/sony/sonyflake"
	"strconv"
	"time"
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
	//Values              []AttributeValue `json:values`  //TODO enable this to grow with the things Aussie wants
}

type RequestResponse struct {
	ServiceRequestID string `json:"service_request_id` // The unique ID of the service request created.
	ServiceNotice    string `json:"service_notice"`    // Information about the action expected to fulfill the request or otherwise address the information reported
	AccountID        string `json"account_id"`         // Unique ID for the user account of the person submitting the request
}

// GetServices returns all Services
func GetServices() ([]Service, error) {
	return allServices()
}

func allServices() ([]Service, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hard code region
	)
	if err != nil {
		return nil, fmt.Errorf("repository: unable to establish session with AWS \n %s", err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String("Services"), //TODO don't hard code Table name
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)
	if err != nil {
		return nil, fmt.Errorf("repository: unable to get all services from database with the following parameters: %+v. \n %s", params, err)
	}

	services := []Service{}

	// For each service, unmarshal and add to array of services
	for _, i := range result.Items {
		service := Service{}
		err = dynamodbattribute.UnmarshalMap(i, &service)
		if err != nil {
			return services, fmt.Errorf("repository: Failed to unmarshal Record: %+v. \n %s", i, err)
		}

		services = append(services, service)
	}
	return services, err
}

// GetRequest returns service information with code
func GetService(code string) (*Service, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hard code region
	)
	if err != nil {
		return nil, fmt.Errorf("repository: unable to establish session with AWS \n %s", err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	input := &dynamodb.GetItemInput{
		TableName: aws.String("Services"), //TODO don't hard code table name ?
		Key: map[string]*dynamodb.AttributeValue{
			"service_code": {
				S: aws.String(code),
			},
		},
	}
	result, err := svc.GetItem(input)
	if err != nil {
		return nil, fmt.Errorf("repository: unable to get specified service from database with the following input: %+v. \n %s", input, err)
	}

	service := Service{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &service)
	if err != nil {
		return &service, fmt.Errorf("repository: Failed to unmarshal service record from database: %+v. \n %s", result.Item, err)
	}

	if service.ServiceCode == "" {
		fmt.Println("Could not find Service")
		return nil, errors.New("service not found")
	}

	return &service, err
}

func IsValidServiceCode(service_code string) bool {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hard code region
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	filt := expression.Contains(expression.Name("service_code"), service_code)
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		fmt.Println((err.Error()))
		panic("While checking if service existed, Got error building database expression.")
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String("Services"),
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println((err.Error()))
		panic("Query API call failed while checking if Service Code was valid:")
	}

	if *result.Count == 1 { //Since "Contains" matches substrings and services codes should be unique, valid IDs match only once
		return true
	}

	return false
}

// GetRequests returns all Requests
func GetRequests() ([]Request, error) {
	return allRequests()
}

func allRequests() ([]Request, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hard code region
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String("Requests"), //TODO don't hard code region
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
		Region: aws.String("us-east-1")}, //TODO don't hard code region
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Requests"), //TODO don't hard code ?
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
		return nil, errors.New("request not found")
	}

	return &request, err
}

func SubmitRequest(request Request) (RequestResponse, error) { //TODO return requestID... perhaps have dynamo increment this... has to be unique , but readable
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}, //TODO don't hard code region
	)
	//TODO handle err.  Consider using MUST above

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Get unique identifier by which this new request will be submitted.
	requestID, _ := genUniqueID()
	//requestID := "Numero Uno"
	request.ServiceRequestId = requestID

	// Assign requested_datetime
	t := time.Now()
	request.RequestedDateTime = t.Format(time.RFC3339)

	//Initialize new request as "open"
	request.Status = "open" //TODO don't hard code ?

	// Initialize service name
	service, _ := GetService(request.ServiceCode)
	request.ServiceName = service.ServiceName

	av, err := dynamodbattribute.MarshalMap(request)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal request, %v", err))
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("Requests"), //TODO don't hard code ?
	}

	_, err = svc.PutItem(input)
	if err != nil {
		panic(fmt.Sprintf("Failed to put item in database, %v", err))
	}

	var response RequestResponse
	response.ServiceRequestID = requestID

	return response, err
}

func genUniqueID() (string, error) {
	// Initialize sonyfake - see usage details at https://github.com/sony/sonyflake

	var st sonyflake.Settings
	// st.MachineID = awsutil.AmazonEC2MachineID //uncomment if running in EC2
	flake := sonyflake.NewSonyflake(st)

	if flake == nil {
		panic("Unique ID Generator, sonyflake, not created")
	}

	// Generate relatively short unique ID
	id, err := flake.NextID()

	if err != nil {
		panic(fmt.Sprintf("Error Generating Unique ID for new Request: flake.NextID() failed with %s\n", err))
	}
	idString := strconv.FormatUint(id, 16) // sonyflake uses Hex
	return idString, err
}
