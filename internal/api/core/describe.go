package core

// API Description will be empty until RunServer() is called.
var apiDescription APIDescription

type APIDescription struct {
	Endpoints map[string]EndpointDescription `json:"endpoints"`
}

type EndpointDescription struct {
	RequestType  string `json:"request-type"`
	ResponseType string `json:"response-type"`
	Description  string `json:"description"`
}

func SetAPIDescription(description APIDescription) {
	apiDescription = description
}

func GetAPIDescription() APIDescription {
	return apiDescription
}
