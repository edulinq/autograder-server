package core

// Server routes will be empty until RunServer() is called.
var serverRoutes []Route

type APIDescription struct {
	Endpoints map[string]EndpointDescription `json:"endpoints"`
}

type EndpointDescription struct {
	RequestType  string `json:"request-type"`
	ResponseType string `json:"response-type"`
}

func SetServerRoutes(routes []Route) {
	serverRoutes = routes
}

func GetServerRoutes() []Route {
	return serverRoutes
}

func Describe(routes []Route) *APIDescription {
	endpointMap := make(map[string]EndpointDescription)
	for _, route := range routes {
		apiRoute, ok := route.(*APIRoute)
		if !ok {
			continue
		}

		endpointMap[apiRoute.GetBasePath()] = EndpointDescription{
			RequestType:  apiRoute.RequestType.String(),
			ResponseType: apiRoute.ResponseType.String(),
		}
	}

	return &APIDescription{
		Endpoints: endpointMap,
	}
}
