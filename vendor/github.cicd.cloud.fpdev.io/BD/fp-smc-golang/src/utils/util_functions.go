/*
An endpoint instance of SMC
author: Dlo Bagari
date:12/02/2020
*/
package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

//convert a response body (io.Reader) to map.
//read the content of the body and converted to a map object
func ResponseToMap(body io.Reader) (map[string][]map[string]string, error) {
	buff, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var objMap map[string][]map[string]string
	if err := json.Unmarshal(buff, &objMap); err != nil {
		return nil, err
	}
	return objMap, nil
}

func ParseResponseToStruct(io io.ReadCloser, obj interface{}) (err error) {
	body, err := ioutil.ReadAll(io)

	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &obj)

	if err != nil {
		return err
	}

	return nil
}
