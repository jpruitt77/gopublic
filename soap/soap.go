//This package is based on github.com/foomo/soap

package soap

import (
	"encoding/xml"
)

// SOAP 1.1 and SOAP 1.2 must expect different ContentTypes and Namespaces.

const (
	SoapVersion11 = "1.1"
	SoapVersion12 = "1.2"

	SoapContentType11 = "text/xml; charset=\"utf-8\""
	SoapContentType12 = "application/soap+xml; charset=\"utf-8\""

	NamespaceSoap11 = "http://schemas.xmlsoap.org/soap/envelope/"
	NamespaceSoap12 = "http://www.w3.org/2003/05/soap-envelope"

	XmlNSXsd = "http://www.w3.org/2001/XMLSchema"
	XmlNSXsi = "http://www.w3.org/2001/XMLSchema-instance"
)

// Verbose be verbose
var Verbose = false

// Envelope type
type Envelope struct {
	XMLName   xml.Name `xml:"soap:Envelope"`
	XmlNSSoap string   `xml:"xmlns:soap,attr,omitempty"`
	XmlNSXsd  string   `xml:"xmlns:xsd,attr"`
	XmlNSXsi  string   `xml:"xmlns:xsi,attr"`
	Header    Header
	Body      Body
}

// Header type
type Header struct {
	XMLName xml.Name `xml:"soap:Header"`

	Content interface{}
}

// Body type
type Body struct {
	XMLName xml.Name `xml:"soap:Body"`

	Fault               *Fault      `xml:",omitempty"`
	Content             interface{} `xml:",omitempty"`
	SOAPBodyContentType string      `xml:"-"`
}

// Fault type
type Fault struct {
	XMLName xml.Name `xml:"soap:Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}

// EnvelopeRequest type
type EnvelopeRequest struct {
	XMLName xml.Name `xml:"Envelope"`

	Header interface{}
	Body   BodyRequest
}

// HeaderRequest type
type HeaderRequest struct {
	XMLName xml.Name `xml:"Header"`

	Content interface{}
}

// BodyRequest type
type BodyRequest struct {
	XMLName xml.Name `xml:"Body"`

	Fault               *FaultRequest `xml:",omitempty"`
	Content             interface{}   `xml:",omitempty"`
	SOAPBodyContentType string        `xml:"-"`
}

// FaultRequest type
type FaultRequest struct {
	XMLName xml.Name `xml:"Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}

// UnmarshalXML implement xml.Unmarshaler
func (b *BodyRequest) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if b.Content == nil {
		return xml.UnmarshalError("Content must be a pointer to a struct")
	}

	var (
		token    xml.Token
		err      error
		consumed bool
	)

Loop:
	for {
		if token, err = d.Token(); err != nil {
			return err
		}

		if token == nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			l(se.Name.Space)
			if consumed {
				return xml.UnmarshalError("Found multiple elements inside SOAP body; not wrapped-document/literal WS-I compliant")
			} else if se.Name.Local == "Fault" {
				b.Fault = &FaultRequest{}
				b.Content = nil

				err = d.DecodeElement(b.Fault, &se)
				if err != nil {
					return err
				}

				consumed = true
			} else {
				b.SOAPBodyContentType = se.Name.Local
				if err = d.DecodeElement(b.Content, &se); err != nil {
					return err
				}

				consumed = true
			}
		case xml.EndElement:
			break Loop
		}
	}

	return nil
}

func (f *Fault) Error() string {
	return f.String
}
