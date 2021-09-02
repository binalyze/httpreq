package httpreq

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponse(t *testing.T) {

	// Start a local HTTP server
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Test-Header", "this is response")
			rw.Header().Set("Content-Type", "application/json")

			_, err := rw.Write([]byte(responseData))
			require.NoError(t, err)
		}),
	)
	defer server.Close()

	url := fmt.Sprintf("%s/post", server.URL)

	resp, err := New(url).Get()
	require.NoError(t, err)

	// Test Response()
	require.Equal(t, resp.resp, resp.Response())

	// Test StatusCode()
	statusCode := resp.StatusCode()
	require.Equal(t, 200, statusCode)

	// Test Headers()
	headers := resp.Headers()
	headerValue, ok := headers["Test-Header"]
	require.True(t, ok)
	require.Equal(t, "this is response", headerValue[0])

	// Test Body()
	result, err := resp.Body()
	require.NoError(t, err)
	require.Equal(t, string(result), responseData)

	// Test Close()
	err = resp.Close()
	require.NoError(t, err)
}

// Test StatusCode 0
func TestStatusCodeErrort(t *testing.T) {
	resp := Response{}
	statusCode := resp.StatusCode()
	require.Equal(t, 0, statusCode)
}

// Test StatusCode 0
func TestHeadersError(t *testing.T) {
	resp := Response{}
	require.Equal(t, 0, len(resp.Headers()))
}

// Test StatusCode 0
func TestBodyError(t *testing.T) {
	resp := Response{}
	_, err := resp.Body()
	require.Error(t, err)
}

// Test Close Error
func TestCloseError(t *testing.T) {
	resp := Response{}
	err := resp.Close()
	require.NoError(t, err)
}

func TestReadBody(t *testing.T) {

	resp := &Response{data: []byte("Hello World")}
	_, err := resp.readBody()
	require.NoError(t, err)

	resp = &Response{}
	_, err = resp.readBody()
	require.Error(t, err)

	resp = &Response{resp: &http.Response{}}
	_, err = resp.readBody()
	require.Error(t, err)

	resp = &Response{resp: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString("Hello World"))}}
	_, err = resp.readBody()
	require.NoError(t, err)
}

func TestSaveFile(t *testing.T) {

	srcFile, err := ioutil.TempFile("", "source-file-*.png")
	require.NoError(t, err)
	defer os.Remove(srcFile.Name())

	message := randStringBytes(30)

	_, err = srcFile.Write([]byte(message))
	require.NoError(t, err)

	// Start a local HTTP server
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.ServeFile(rw, req, srcFile.Name())
		}),
	)
	defer server.Close()

	url := fmt.Sprintf("%s/get", server.URL)

	resp, err := New(url).Get()
	require.NoError(t, err)

	dstFile, err := ioutil.TempFile("", "dest-file-*.png")
	require.NoError(t, err)
	defer os.Remove(dstFile.Name())

	// Test Body()
	err = resp.SaveFile(dstFile.Name())
	require.NoError(t, err)

	// Read destination file
	data, err := ioutil.ReadFile(dstFile.Name())
	require.NoError(t, err)

	require.Equal(t, string(data), message)

	// Test Close()
	err = resp.Close()
	require.NoError(t, err)
}

func TestSaveFileError(t *testing.T) {

	// readBody error
	resp := &Response{}
	err := resp.SaveFile("")
	require.Error(t, err)

	// len(data) error
	resp = &Response{resp: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(""))}}
	err = resp.SaveFile("")
	require.Error(t, err)

	// os.Create error
	resp = &Response{resp: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString("hello world"))}}
	err = resp.SaveFile("/wrong/path")
	require.Error(t, err)
}

func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestDownloadFile(t *testing.T) {

	contentType := "binary/octet-stream"
	url, content, downloadDir := testSetupDownloadFile(t, contentType,
		func(f string) string { return fmt.Sprintf("attachment;filename=%q", f) })

	resp, err := New(url).Get()
	require.NoError(t, err)

	ctype, filePath, err := resp.DownloadFile(downloadDir)
	require.NoError(t, err)
	require.Equal(t, contentType, ctype)

	// Read destination file
	data, err := ioutil.ReadFile(filePath)

	require.NoError(t, err)
	require.Equal(t, string(data), content)
	require.NoError(t, resp.Close())
}

func TestDownloadFile_MissingHeader(t *testing.T) {

	url, _, downloadDir := testSetupDownloadFile(t, "", nil)

	resp, err := New(url).Get()
	require.NoError(t, err)

	_, _, err = resp.DownloadFile(downloadDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "content-disposition header missing")
}

func TestDownloadFile_BadHeader(t *testing.T) {

	url, _, downloadDir := testSetupDownloadFile(t, "", func(s string) string { return "/" })

	resp, err := New(url).Get()
	require.NoError(t, err)

	_, _, err = resp.DownloadFile(downloadDir)
	require.Error(t, err)
}

func TestDownloadFile_WrongHeader(t *testing.T) {

	url, _, downloadDir := testSetupDownloadFile(t, "", func(string) string { return "attachment;" })

	resp, err := New(url).Get()
	require.NoError(t, err)

	_, _, err = resp.DownloadFile(downloadDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "filename missing")
}

func testSetupDownloadFile(t *testing.T, contentType string, contentDisp func(string) string) (url, fileContent, downloadDir string) {
	t.Helper()

	srcFile, err := ioutil.TempFile("", "*")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Remove(srcFile.Name())
	})

	fileContent = randStringBytes(30)

	_, err = srcFile.Write([]byte(fileContent))
	require.NoError(t, err)

	// Start a local HTTP server
	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if contentDisp != nil {
				rw.Header().Set(
					"content-disposition",
					contentDisp(filepath.Base(srcFile.Name())),
				)
			}

			if contentType != "" {
				rw.Header().Set("content-type", contentType)
			}

			http.ServeFile(rw, req, srcFile.Name())
		}),
	)
	t.Cleanup(func() { server.Close() })

	downloadDir, err = ioutil.TempDir("", t.Name()+"*")
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Remove(downloadDir) })

	return server.URL, fileContent, downloadDir
}
