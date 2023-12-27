![Workflow](https://github.com/binalyze/httpreq/actions/workflows/go.yml/badge.svg)

# httpreq

`httpreq` is an http request library written with Golang to make requests and handle responses easily.


## Install 

```bash 
  go get github.com/binalyze/httpreq
```

## Overview

`httpreq` implements a friendly API over Go's existing `net/http` library.
  
`Req` and `Response` are two most important struct. You can think of `Req` as a client that initiate HTTP requests, `Resp` as a information container for the request and response. They all provide simple and convenient APIs that allows you to do a lot of things.

``` go
req := httpreq.New(url)
resp, err := req.Get()
```

## Roadmap

- Support query parameters
- Support cookies
- Support XML
- Support proxy
- Configurable transport

## Usage

Here is an example to use some helper methods of `httpreq`. You can find more examples in test files.
#### Request
```go
  // Create new request
  req := httpreq.New("https://your-address-to-send-json.com")

  // Set Timeout
  req.SetTimeout(30 * time.Second)

  // Set Header (i.e. JWT Bearer Token)
  var bearer = "Bearer " + <ACCESS TOKEN HERE>
  req.SetHeaders(map[string]string{"Authorization": bearer})

  // Set Content Type
  req.SetContentType("application/json") 

  // Set JSON raw body
  info := &Data{
    FirstName: "John",
    LastName:  "Doe",
    Age:       42,
  }

  jsonData, _ := json.Marshal(info)
	
  req.SetBody(jsonData)

  // Send Request
  resp, _ := req.Post()
```

#### Response
```go
  // Body
  result, err := resp.Body()

  // Original response
  response := resp.Response()

  // StatusCode
  statusCode := resp.StatusCode()

  // Headers
  headers := resp.Headers()
```
