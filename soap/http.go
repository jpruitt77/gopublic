package soap

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

var xmlHeader = []byte(xml.Header)

const xmlHeaderLen = 39

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.outputStarted = true

	b = append(xmlHeader, b...)

	if Verbose {
		l("writing response: " + string(b))
	}

	return w.w.Write(b)
}

func (w *responseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func addSOAPHeader(w http.ResponseWriter, contentLength int, contentType string) {
	setContentType(w, contentType)
	w.Header().Set("Content-Length", fmt.Sprint(contentLength+xmlHeaderLen))
}
