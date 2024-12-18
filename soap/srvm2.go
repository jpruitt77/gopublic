package soap

import (
	"errors"
	"net/http"
)

func createResponse(response interface{}, header interface{}) Envelope {
	env := Envelope{
		XmlNSSoap: NamespaceSoap11,
		XmlNSXsd:  XmlNSXsd,
		XmlNSXsi:  XmlNSXsi,
		Header:    Header{Content: header},
		Body: Body{
			Content: response,
		},
	}

	return env
}

func check(e any, msg string, w http.ResponseWriter, s *Server) (hasError bool) {
	switch e := e.(type) {
	case error:
		if e != nil {
			s.handleError(errors.New(msg+e.Error()), w)
			hasError = true
		}
	case bool:
		if !e {
			s.handleError(errors.New(msg), w)
			hasError = true
		}
	}

	return
}
