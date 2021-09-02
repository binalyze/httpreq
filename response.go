package httpreq

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
)

// Response is the main struct which holds the http.Response and data.
type Response struct {
	resp *http.Response
	data []byte
}

// Response returns the original http.Response
func (r *Response) Response() *http.Response {
	return r.resp
}

// StatusCode returns the status code of the response
func (r *Response) StatusCode() int {
	if r == nil || r.resp == nil {
		return 0
	}
	return r.resp.StatusCode
}

// Headers returns the headers of the response
func (r *Response) Headers() http.Header {
	if r == nil || r.resp == nil {
		return nil
	}
	return r.resp.Header
}

// Body returns the response body
func (r *Response) Body() ([]byte, error) {
	body, err := r.readBody()
	if err != nil {
		logger.Errorf("Can not read http.Response body Error: %v", err)
		return nil, err
	}
	return body, nil
}

// DownloadFile looks for Content-Disposition header to find the filename attribute and returns the content-type
// header with saved file path that is saved under given downloadDir.
func (r *Response) DownloadFile(downloadDir string) (contentType string, filePath string, err error) {
	headers := r.Headers()
	if headers == nil {
		err = errors.New("http response headers missing")
		logger.Errorf("%v", err)
		return "", "", err
	}

	contentType = headers.Get("Content-Type")

	disposition := headers.Get("Content-Disposition")
	if disposition == "" {
		err = errors.New("content-disposition header missing")
		logger.Errorf("%v", err)
		return contentType, "", err
	}

	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		logger.Errorf("mime.ParseMediaType error: %v", err)
		return contentType, "", err
	}

	fileName := params["filename"]
	if fileName == "" {
		err = errors.New("filename missing in content-disposition")
		logger.Errorf("%v", err)
		return contentType, "", err
	}

	filePath = path.Join(downloadDir, fileName)

	err = r.SaveFile(filePath)
	if err != nil {
		logger.Errorf("cannot save file error: %v", err)
		return contentType, "", err
	}

	return contentType, filePath, nil
}

// SaveFile reads body and then saves the file defined in body
func (r *Response) SaveFile(filePath string) error {
	data, err := r.readBody()
	if err != nil {
		logger.Errorf("Can not save response to file %s Error: %v", filePath, err)
		return err
	}

	if len(data) == 0 {
		err := errors.New("Downloaded file is empty. Can not save empty response to file " + filePath)
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		logger.Errorf("Can not create file %s Error: %v", filePath, err)
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewReader(data))
	if err != nil {
		logger.Errorf("Can write to file %s Error: %v", filePath, err)
		return err
	}

	err = f.Sync()
	if err != nil {
		logger.Errorf("Can't sync file %s Error: %v", filePath, err)
		return err
	}

	return err
}

// Close closes the http.Response
func (r *Response) Close() error {
	if r == nil || r.resp == nil || r.resp.Body == nil {
		return nil
	}

	err := r.resp.Body.Close()
	if err != nil {
		logger.Errorf("Can't close http response body Error: %v", err)
		return err
	}

	return nil
}

// readBody reads the http.Response body and assigns it to the r.Body
func (r *Response) readBody() ([]byte, error) {

	// If r.data already set then return r.data
	if len(r.data) != 0 {
		return r.data, nil
	}

	// Check if Response.resp (*http.Response) is nil
	if r.resp == nil {
		err := fmt.Errorf("http.Response is nil")
		logger.Errorf("%v", err)
		return nil, err
	}

	// Check if Response.resp.Body (*http.Response.Body) is nil
	if r.resp.Body == nil {
		err := fmt.Errorf("http.Response's Body is nil")
		logger.Errorf("%v", err)
		return nil, err
	}

	// Read response body
	b, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		logger.Errorf("Can't read http.Response body Error: %v", err)
		return nil, err
	}

	// Set response readBody
	r.data = b

	// Close response body
	err = r.resp.Body.Close()
	if err != nil {
		logger.Errorf("Can't close http.Response body Error: %v", err)
		return nil, err
	}

	return b, nil
}
