package repository

import "errors"

// Single service (type) offered via Open311
type Service struct {
	ServiceCode	string	`json:service_code`
	ServiceName	string	`json:service_name`
	Description	string	`json:description`
	Metadata		bool		`json:metadata`
	Type				string	`json:type`
	Keywords		string	`json:keywords`
	Group				string	`json:group`
}

// Service definition associated with a service code.
// These attributes can be unique to the city/jurisdiction
type ServiceDefinition struct {
	ServiceCode string							`json:service_code`
	Attributes  []ServiceAttribute `json:attributes`
}

// Single attribute extension for a service
type ServiceAttribute struct {
	Code 								string						`json:code`
	DataType						string						`json:datatype`
	Variable						bool							`json:variable`
	Required						bool							`json:required`
	Order								int32							`json:order`
	Description 				string						`json:description`
	DataTypeDescription	string						`json:datatype_description`
	Values							[]AttributeValue	`json:values`
}

// Possible value for ServiceAttribute that defines lists
type AttributeValue struct {
	Key 	string	`json:key`
	name	string	`json:name`
}


// Issues that have been reported as service requests.  Location
// is submitted via lat/long or address
type Request struct {
	ServiceRequestId	string	`json:service_request_id`
	ServiceCode 			string	`json:service_code`
	ServiceName				string	`json:service_name`
	Description 			string	`json:description`
	Address 					string	`json:address`
	AddressId					string	`json:address_id`
	ZipCode						int32		`json:zipcode`
	Latitude					float32 `json:lat`
	Longitude 				float32	`json:lon`
	MediaUrl					string	`json:media_url`
	AgencyResponsible string 	`json:agency_responsible`
	ServiceNotice			string	`json:service_notice`
	Status 						string 	`json:"name"`
	StatusNotes  			string  `json:status_notes`
	RequestedDateTime	string	`json:requested_datetime`
	UpdatedDateTime		string	`json:update_datetime`
	ExpectedDateTime	string	`json:expected_datetime`
}


func allServices() []Service {
	return []Service{
		Service{"001",
		 				"Cans left out 24x7",
		 				"Garbage or recycling cans left out for more than 24 hours after collection.",
		 				true,
		 				"realtime",
		 				"",
						"sanitation" },
		Service{"002",
						"Construction plate shifted",
						"Metal construction plate covering the street or sidewalk has been moved.",
						true,
						"realtime",
						"",
						"street"},
		Service{"003",
						"Curb or curb ramp defect",
						"Sidewalk curb or ramp has problems such as cracking, missing pieces, holes, and/or chipped curb.",
						true,
						"realtime",
						"",
						"street"},
	}
}

// GetRequest returns service information with code
func GetService(code string) (*Service, error) {
	for _, s := range allServices() {
		if s.ServiceCode == code {
			return &s, nil
		}
	}
	return nil, errors.New("Service not found")
}

// GetServices returns all Services
func GetServices() []Service {
	return allServices()
}


func allRequests() []Request {
	return []Request{
		Request{
				"638349",
				"003",
				"Sidewalk and Curb Issues",
				"",
				"8TH AVE and JUDAH ST",
				"545483",
				12309,
				42.8170100653,
				-73.9246079682,
				"http://www.google.com",
				"",
				"",
				"open",
				"",
				"2010-04-19T06:37:38-08:00",
		    "2010-04-19T06:37:38-08:00",
		    "2010-04-19T06:37:38-08:00",
		},
	}
}

// GetRequest returns request with id
func GetRequest(id string) (*Request, error) {
	for _, r := range allRequests() {
		if r.ServiceRequestId == id {
			return &r, nil
		}
	}
	return nil, errors.New("Request not found")
}

// GetRequests returns all Requests
func GetRequests() []Request {
	return allRequests()
}
