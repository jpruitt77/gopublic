//This package is based on github.com/foomo/soap

package soap

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
)

func (s *Server) UseSoap11() {
	s.SoapVersion = SoapVersion11
	s.ContentType = SoapContentType11
}

func (s *Server) UseSoap12() {
	s.SoapVersion = SoapVersion12
	s.ContentType = SoapContentType12
}

// RegisterHandler register to handle an operation
func (s *Server) RegisterHandler(path string, action string, messageType string, requestFactory RequestFactoryFunc, operationHandlerFunc OperationHandlerFunc) {
	_, pathHandlersOK := s.handlers[path]
	if !pathHandlersOK {
		s.handlers[path] = make(map[string]map[string]*operationHander)
	}
	_, ok := s.handlers[path][action]
	if !ok {
		s.handlers[path][action] = make(map[string]*operationHander)
	}
	s.handlers[path][action][messageType] = &operationHander{
		handler:        operationHandlerFunc,
		requestFactory: requestFactory,
	}
}

func (s *Server) handleError(err error, w http.ResponseWriter) {
	// has to write a soap fault
	l("handling error:", err)
	responseEnvelope := createResponse(&Fault{
		String: err.Error(),
	}, nil)
	xmlBytes, xmlErr := s.Marshaller.Marshal(responseEnvelope, s.SoapVersion)
	if xmlErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not marshal soap fault for: " + err.Error() + " xmlError: " + xmlErr.Error()))
	} else {
		addSOAPHeader(w, len(xmlBytes), s.ContentType)
		w.Write(xmlBytes)
	}
}

// WriteHeader first sets header like content-type and then writes the header
func (s *Server) WriteHeader(w http.ResponseWriter, code int) {
	setContentType(w, s.ContentType)
	w.WriteHeader(code)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	soapAction := r.Header.Get("SOAPAction")
	l("ServeHTTP method:", r.Method, ", path:", r.URL.Path, ", SOAPAction", "\""+soapAction+"\"")
	// we have a valid request time to call the handler
	w = &responseWriter{
		w:             w,
		outputStarted: false,
	}
	switch r.Method {
	case "POST":
		l("incoming POST")

		// Read raw post request
		soapRequestBytes, err := io.ReadAll(r.Body)
		if check(err, "could not read POST:: ", w, s) {
			return
		}

		// Check if a handler exist for the path requested
		pathHandlers, pathHandlerOK := s.handlers[r.URL.Path]
		if check(pathHandlerOK, "unknown path", w, s) {
			return
		}

		// Check if a there is a handler for the SoapAction retrieved from header
		actionHandlers, ok := pathHandlers[soapAction]
		if check(ok, "unknown action \""+soapAction+"\"", w, s) {
			return
		}

		// We need to find out what is in the body
		// Unmarshal raw request into probeEnvelope
		probeEnvelope := s.CreateEnvelopeRequest()
		err = s.Marshaller.Unmarshal(soapRequestBytes, &probeEnvelope, s.SoapVersion)
		if check(err, "could not probe soap body content:: ", w, s) {
			return
		}

		// Check if the type returned from the body of the request
		// matches any actionHandlers and return the matching handler
		t := probeEnvelope.Body.SOAPBodyContentType
		l("found content type", t)
		actionHandler, ok := actionHandlers[t]
		if check(ok, "no action handler for content type: \""+t+"\"", w, s) {
			return
		}

		// Return the correct interface/struct for actionHandler
		request := actionHandler.requestFactory()
		probeEnvelope.Body.Content = request

		// Finish unmarshalling request with the returned interface/struct
		err = xml.Unmarshal(soapRequestBytes, &probeEnvelope)
		if check(err, "could not unmarshal request:: ", w, s) {
			return
		}

		l("request", jsonDump(probeEnvelope))

		// Check if the Multispeak header is present and valid from sender
		if check(s.CheckHeader(probeEnvelope), "invalid MultiSpeak header", w, s) {
			l("invalid Multispeak header: maybe wrong user and password")
			return
		}

		// Proccess request and retrieve response
		response, header, err := actionHandler.handler(request, w, r)
		if check(err, "", w, s) {
			l("action handler threw up")
			return
		}

		l("result", jsonDump(response))
		if !w.(*responseWriter).outputStarted {
			// Create an Envelope struct with response
			responseEnvelope := createResponse(response, header)

			// Marshal Envelope
			xmlBytes, err := s.Marshaller.Marshal(responseEnvelope, s.SoapVersion)
			check(err, "could not marshal response:: ", w, s)

			// Add header info and write/send response
			addSOAPHeader(w, len(xmlBytes), s.ContentType)
			_, writeErr := w.Write(xmlBytes)
			if writeErr != nil {
				l(writeErr)
			}
		} else {
			l("action handler sent its own output")
		}

	default:
		// this will be a soap fault !?
		s.handleError(errors.New("this is a soap service - you have to POST soap requests"), w)
	}
}

// ListenAndServe run standalone
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}
