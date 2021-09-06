package httpreq

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var logger Logger = NewBuiltinLogger()

// Req is main struct for requests
type Req struct {
	request *http.Request
	client  *http.Client
	address string
	err     error
}

// New creates a new HTTP Request
func New(address string) *Req {

	r := new(Req)

	r.address = address

	r.request = &http.Request{
		Method: "GET",
		Header: make(http.Header),
	}

	r.client = &http.Client{
		Transport: http.DefaultTransport.(*http.Transport),
		Timeout:   time.Second * 30,
	}

	return r
}

// SetTLSConfig changes the request TLS client configuration
func (r *Req) SetTLSConfig(c *tls.Config) *Req {
	r.client.Transport.(*http.Transport).TLSClientConfig = c
	return r
}

// SetTimeout changes the request timeout
func (r *Req) SetTimeout(d time.Duration) *Req {
	r.client.Timeout = d
	return r
}

// SetHeaders sets request headers
func (r *Req) SetHeaders(headers map[string]string) *Req {
	if len(headers) > 0 {
		for k, v := range headers {
			r.request.Header.Set(k, v)
		}
	}
	return r
}

// SetContentType sets content type of request
func (r *Req) SetContentType(contentType string) *Req {
	r.request.Header.Set("Content-Type", contentType)
	return r
}

//SetTransport sets transport configuration of request
func (r *Req) SetTransport(transport *http.Transport) *Req {
	r.client.Transport = transport
	return r
}

// SetBody sets request body
func (r *Req) SetBody(data []byte) *Req {
	r.request.Body = ioutil.NopCloser(bytes.NewReader(data))
	r.request.GetBody = func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewReader(data)), nil
	}

	r.request.ContentLength = int64(len(data))
	return r
}

//SetBodyXML sets content type as XML.
func (r *Req) SetBodyXML() *Req {
	r.SetContentType("application/xml; charset=UTF-8")
	return r
}

// SetForm creates form and add files and data to form.
func (r *Req) SetForm(files []map[string]string, fields []map[string]string) *Req {

	// If there is an error in chain, then do nothing and return early
	if r.err != nil {
		return r
	}

	var b bytes.Buffer

	w := multipart.NewWriter(&b)

	for _, file := range files {
		for key, value := range file {
			if err := createFormFile(w, key, value); err != nil {
				logger.Errorf("Failed to create form file %s as %s Error: %v", key, value, err)
				r.err = err
				return r
			}
		}
	}

	for _, field := range fields {
		for k, v := range field {
			err := w.WriteField(k, v)
			if err != nil {
				logger.Errorf("Can't write field %s as %s Error: %v", k, v, err)
				r.err = err
				return r
			}
		}
	}

	if err := w.Close(); err != nil {
		logger.Errorf("Can't close multipart writer Error: %v", err)
		r.err = err
		return r
	}

	r.request.Body = ioutil.NopCloser(bytes.NewReader(b.Bytes()))

	// GetBody is required to be set for protecting body on redirections
	r.request.GetBody = func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewReader(b.Bytes())), nil
	}

	r.request.ContentLength = int64(b.Len())
	r.request.Header.Set("Content-Type", w.FormDataContentType())

	return r
}

// SetProxy sets proxy URL to http client
func (r *Req) SetProxy(u string) *Req {
	proxyURL, err := url.Parse(u)
	if err != nil {
		r.err = err
		return r
	}
	transport := http.Transport{}
	transport.Proxy = http.ProxyURL(proxyURL)

	r.client.Transport = &transport
	return r
}

// Get is a get http request
func (r *Req) Get() (*Response, error) {
	return r.send(http.MethodGet)
}

// Post is a post http request
func (r *Req) Post() (*Response, error) {
	return r.send(http.MethodPost)
}

// PostJSON is a POST http request as JSON
func (r *Req) PostJSON() (*Response, error) {
	r.SetContentType("application/json")
	return r.send(http.MethodPost)
}

// Put is a put http request
func (r *Req) Put() (*Response, error) {
	return r.send(http.MethodPut)
}

// Delete is a delete http request
func (r *Req) Delete() (*Response, error) {
	return r.send(http.MethodDelete)
}

// Send HTTP request
func (r *Req) send(method string) (*Response, error) {

	// If there is an error in chain, then do nothing and return error
	if r.err != nil {
		return nil, r.err
	}

	if r.request.ContentLength > 0 && r.request.GetBody == nil {
		return nil, errors.New("request.GetBody cannot be nil because it prevents redirection when content length>0")
	}

	// Set method
	r.request.Method = method

	// Set URL
	URL, err := generateURL(r.address)
	if err != nil {
		logger.Errorf("Error generating URL: %s, %v", r.address, err)
		return nil, err
	}
	r.request.URL = URL

	// Execute request and get response
	resp, err := r.client.Do(r.request)
	if err != nil {
		logger.Errorf("Error sending HTTP request: %s, %v", URL, err)
		return nil, err
	}

	// Build Response
	response := &Response{
		resp: resp,
	}

	return response, nil
}

// generateURL generates URL from address
func generateURL(address string) (*url.URL, error) {
	address = strings.ToLower(address)

	// Parse URL
	parsedURL, err := url.Parse(address)
	if err != nil {
		logger.Errorf("URL parsing error: %s, %v", address, err)
		return nil, err
	}

	return parsedURL, nil
}

// createFormFile reads defined files and adds to form
func createFormFile(w *multipart.Writer, key, value string) error {
	part, err := w.CreateFormFile(key, value)
	if err != nil {
		logger.Errorf("Failed to create form data from file %v Error: %v", value, err)
		return err
	}

	f, err := os.Open(value)
	if err != nil {
		logger.Errorf("Failed to open file %s Error: %v", value, err)
		return err
	}

	defer func() {
		if err = f.Close(); err != nil {
			logger.Errorf("Failed to close file %s Error: %v", value, err)
		}
	}()

	_, err = io.Copy(part, f)
	if err != nil {
		logger.Errorf("Can't copy file %s Error: %v", value, err)
		return err
	}

	return nil
}
