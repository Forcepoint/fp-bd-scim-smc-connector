package utils

import (
	"bytes"
	"encoding/json"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/httpClient"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/interfaces"
	"net/http"
)

/*
Read resources by using GET requests. To save network bandwidth and avoid transferring the complete body
in the response, HEAD requests and ETags are supported.
*/
func Get(url, cookie string) (*http.Response, error) {
	smcRequest, err := buildSmcRequest("GET", url, cookie, nil)
	if err != nil {
		return nil, err
	}
	return smcRequest.Run()
}

func GetSubElements(url, cookie string, element interfaces.SmcElement) (*http.Response, error) {
	smcRequest, err := buildSmcRequest("GET", url, cookie, element)
	if err != nil {
		return nil, err
	}
	return smcRequest.Run()
}

/*
Create resources by using POST requests on the URI of the collection that lists all elements.
The URI of the created resource is returned in the Location header field
Trigger actions such as Policy Uploads by using POST requests
*/
func Create(url, cookie string, element interfaces.SmcElement) (*http.Response, error) {
	smcRequest, err := buildSmcRequest("POST", url, cookie, element)
	if err != nil {
		return nil, err
	}
	return smcRequest.Run()
}

func CreateSubElement(url, cookie string, element interfaces.SmcElement) (*http.Response, error) {
	smcRequest, err := buildSmcRequest("POST", url, cookie, element)
	if err != nil {
		return nil, err
	}
	return smcRequest.Run()
}

/*
Update resources by using PUT requests. All updates are conditional and rely on ETags
*/
func Update(url, cookie string, element interfaces.SmcElement) (*http.Response, error) {
	smcRequest, err := buildSmcRequest("PUT", url, cookie, element)
	if err != nil {
		return nil, err
	}
	return smcRequest.Run()
}

func UpdateSubElement(url, cookie string, element interfaces.SmcElement) (*http.Response, error) {
	smcRequest, err := buildSmcRequest("PUT", url, cookie, element)
	if err != nil {
		return nil, err
	}
	return smcRequest.Run()
}

func Delete(url, cookie string, element interfaces.SmcElement) (*http.Response, error) {
	smcRequest, err := buildSmcRequest("DELETE", url, cookie, element)
	if err != nil {
		return nil, err
	}
	return smcRequest.Run()
}

func DeleteSubElement(url, cookie string, element interfaces.SmcElement) (*http.Response, error) {
	smcRequest, err := buildSmcRequest("DELETE", url, cookie, element)
	if err != nil {
		return nil, err
	}
	return smcRequest.Run()
}

func buildSmcRequest(action, url, cookie string, element interfaces.SmcElement) (*httpClient.SmcRequest, error) {
	var buffer *bytes.Buffer
	var smcRequest httpClient.SmcRequest

	if element == nil {
		buffer = bytes.NewBuffer(nil)
	} else {
		rawBytes, err := json.Marshal(element)
		if err != nil {
			return nil, err
		}
		buffer = bytes.NewBuffer(rawBytes)
	}

	smcRequest = httpClient.SmcRequest{
		MethodName: action,
		Url:        url,
		BodyData:   buffer,
		Headers:    defaultHeaders(cookie),
		RequestObj: nil,
	}

	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}

	return &smcRequest, nil
}

func defaultHeaders(cookie string) (headers map[string]string) {
	headers = make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Cookie"] = cookie
	return
}
