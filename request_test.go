package httpreq

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var responseData = `{"success": true,"data": "done!"}`
var token = `123456`

type Data struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
}

func TestMethods(t *testing.T) {

	var table = []struct {
		name string
	}{
		{"get"},
		{"post"},
		{"put"},
		{"delete"},
		{"postjson"},
	}

	// Start a local HTTP server
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/get":
				t.Log("Get request executed")
				require.Equal(t, req.Method, "GET")
			case "/post":
				t.Log("Post request executed")
				require.Equal(t, req.Method, "POST")
			case "/put":
				t.Log("Put request executed")
				require.Equal(t, req.Method, "PUT")
			case "/delete":
				t.Log("Delete request executed")
				require.Equal(t, req.Method, "DELETE")
			case "/postjson":
				t.Log("PostJSON request executed")
				require.Equal(t, req.Method, "POST")
			}
		}),
	)
	defer server.Close()

	for _, row := range table {
		url := fmt.Sprintf("%s/"+row.name, server.URL)
		r := New(url)

		switch row.name {
		case "get":
			_, err := r.Get()
			require.NoError(t, err)
		case "post":
			_, err := r.Post()
			require.NoError(t, err)
		case "put":
			_, err := r.Put()
			require.NoError(t, err)
		case "delete":
			_, err := r.Delete()
			require.NoError(t, err)
		case "postjson":
			_, err := r.PostJSON()
			require.NoError(t, err)
		}
	}
}

func TestPost(t *testing.T) {

	// Start a local HTTP server
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			// Test request
			require.Equal(t, req.URL.String(), "/post")
			require.Equal(t, req.Method, "POST")
			require.Equal(t, req.Header.Get("X-SecurityToken"), token)
			require.Equal(t, req.Header.Get("Content-Type"), "application/json")
			require.Equal(t, req.Header.Get("Test-Header"), "this is a test")

			rw.Header().Set("Content-Type", "application/json")
			_, err := rw.Write([]byte(responseData))
			require.NoError(t, err)
		}),
	)
	defer server.Close()

	url := fmt.Sprintf("%s/post", server.URL)

	r := New(url)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	r.SetTLSConfig(tlsConfig)
	r.SetTimeout(30 * time.Second)
	r.SetHeaders(map[string]string{"Test-Header": "this is a test"})
	r.SetContentType("application/json")

	info := &Data{
		FirstName: "John",
		LastName:  "Doe",
		Age:       42,
	}

	jsonData, err := json.Marshal(info)
	require.NoError(t, err)

	// Set body
	r.SetBody(jsonData)

	// Send Request
	resp, err := r.Post()
	require.NoError(t, err)

	// Get Response body
	sonuc, err := resp.Body()
	require.NoError(t, err)

	require.Equal(t, string(sonuc), responseData)
}

func TestPostJSON(t *testing.T) {

	// Start a local HTTP server
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, req.URL.String(), "/post-json")
			require.Equal(t, req.Method, "POST")

			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("Test-Header", "test response header")

			_, err := rw.Write([]byte(responseData))
			require.NoError(t, err)
		}),
	)
	defer server.Close()

	url := fmt.Sprintf("%s/post-json", server.URL)

	r := New(url)

	info := &Data{
		FirstName: "John",
		LastName:  "Doe",
		Age:       42,
	}

	jsonData, err := json.Marshal(info)
	require.NoError(t, err)

	// Set body
	r.SetBody(jsonData)

	// Send Request
	resp, err := r.PostJSON()
	require.NoError(t, err)

	response := resp.Response()
	require.NotEqual(t, nil, response)

	statusCode := resp.StatusCode()
	require.Equal(t, 200, statusCode)

	headers := resp.Headers()
	headerValue, ok := headers["Test-Header"]
	require.True(t, ok)
	require.Equal(t, "test response header", headerValue[0])

	sonuc, err := resp.Body()
	require.NoError(t, err)

	err = resp.Close()
	require.NoError(t, err)

	require.Equal(t, string(sonuc), responseData)
}

func TestSetFormEarlyError(t *testing.T) {
	r := &Req{err: fmt.Errorf("Test Error")}
	files := []map[string]string{}
	fields := []map[string]string{}
	r.SetForm(files, fields)
	require.Error(t, r.err)
}

func TestSetFormFileError(t *testing.T) {
	r := New("")
	files := []map[string]string{}
	file := make(map[string]string)
	file["file"] = "/wrong/path"
	files = append(files, file)

	fields := []map[string]string{}
	r.SetForm(files, fields)
	require.Error(t, r.err)
}

func TestSetForm(t *testing.T) {

	// Create new http request instance with URL
	r := New("")

	// File list
	f, err := ioutil.TempFile("", "_httpreq_set_form_file_*")
	require.NoError(t, err)
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	files := []map[string]string{}
	file := make(map[string]string)
	file["file"] = f.Name()
	files = append(files, file)

	// Form fields
	fields := []map[string]string{}
	field := make(map[string]string)
	field["taskId"] = "123456"
	fields = append(fields, field)

	// Add Files
	r.SetForm(files, fields)
	require.NoError(t, r.err)
}

func TestSendEarlyError(t *testing.T) {
	r := &Req{err: fmt.Errorf("Test Error")}
	resp, err := r.send("GET")
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestSendGenerateURLError(t *testing.T) {
	// % should causes error at url.Parse
	r := New("%")
	resp, err := r.send("GET")
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestSendRequestError(t *testing.T) {
	// wrong-host should causes request error
	r := New("wrong-host")
	resp, err := r.send("GET")
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestGenerateURL(t *testing.T) {
	// % should causes error at url.Parse
	url, err := generateURL("%")
	require.Error(t, err)
	require.Nil(t, url)
}

func TestSetBodyXML(t *testing.T) {
	r := New("")
	r.SetBodyXML()
	require.Equal(t, "application/xml; charset=UTF-8", r.request.Header.Get("Content-Type"))
}

func TestSetTransport(t *testing.T) {
	r := New("")
	transportConfig := &http.Transport{
		MaxIdleConns: 2,
	}
	r.SetTransport(transportConfig)
	require.Equal(t, transportConfig, r.client.Transport)
}
