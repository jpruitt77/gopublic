//This package is based on github.com/foomo/soap

package soap

import "net/http"

// OperationHandlerFunc runs the actual business logic - request is whatever you constructed in RequestFactoryFunc
type OperationHandlerFunc func(request interface{}, w http.ResponseWriter, httpRequest *http.Request) (response interface{}, header interface{}, err error)

// RequestFactoryFunc constructs a request object for OperationHandlerFunc
type RequestFactoryFunc func() interface{}

type DummyContent struct{}

type operationHander struct {
	requestFactory RequestFactoryFunc
	handler        OperationHandlerFunc
}

type responseWriter struct {
	w             http.ResponseWriter
	outputStarted bool
}

// Server a SOAP server, which can be run standalone or used as a http.HandlerFunc
type Server struct {
	handlers              map[string]map[string]map[string]*operationHander
	Marshaller            XMLMarshaller
	ContentType           string
	SoapVersion           string
	Port                  string
	Path                  string
	CheckHeader           func(request EnvelopeRequest) error
	CreateEnvelopeRequest func() EnvelopeRequest
}

// NewServer construct a new SOAP server
func NewServer(port string, path string) *Server {
	s := &Server{
		handlers:    make(map[string]map[string]map[string]*operationHander),
		Marshaller:  newDefaultMarshaller(),
		ContentType: SoapContentType11,
		SoapVersion: SoapVersion11,
		Port:        port,
		Path:        path,
		CheckHeader: func(request EnvelopeRequest) error { return nil },
		CreateEnvelopeRequest: func() EnvelopeRequest {
			return EnvelopeRequest{
				Body: BodyRequest{
					Content: &DummyContent{},
				},
			}
		},
	}
	return s
}
