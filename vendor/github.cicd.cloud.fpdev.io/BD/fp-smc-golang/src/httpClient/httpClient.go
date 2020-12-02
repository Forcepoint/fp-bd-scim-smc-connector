/*
A Http Client instance for SMC
author: Dlo Bagari
date:12/02/2020
*/

package httpClient

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

var client *http.Client

func init() {
	client = &http.Client{}
}

type SmcRequest struct {
	MethodName string
	Url        string
	BodyData   io.Reader
	Headers    map[string]string
	RequestObj *http.Request
}

//Add a header to HTTP header
func (r *SmcRequest) AddHeader(key string, value string) *SmcRequest {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[key] = value
	return r
}

//add multiple headers to HTTP Header
func (r *SmcRequest) AddHeaders(headers map[string]string) *SmcRequest {
	for key, value := range headers {
		r.Headers[key] = value
	}
	return r
}

// define the type of the Method for HTTP request
func (r *SmcRequest) Method(method string) *SmcRequest {
	r.MethodName = strings.ToUpper(method)
	return r

}

//Add Body to HTTP request
func (r *SmcRequest) Body(body io.Reader) *SmcRequest {
	r.BodyData = body
	return r
}

//Generate HTTP request
func (r *SmcRequest) GenerateRequest() error {
	req, err := http.NewRequest(r.MethodName, r.Url, r.BodyData)
	if err != nil {
		return errors.New("failed in generating a http request")
	}
	r.RequestObj = req
	if r.Headers != nil {
		for k, v := range r.Headers {
			req.Header.Set(k, v)
		}
	}
	return nil
}

//Run HTTP request
func (r *SmcRequest) Run() (*http.Response, error) {
	return client.Do(r.RequestObj)
}
