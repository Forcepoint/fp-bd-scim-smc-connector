/*
An endpoint instance of SMC
author: Dlo Bagari
date:12/02/2020
*/

package smc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/httpClient"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/smc/responses"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/utils"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Smc struct {
	APIVersion  string            `json:"apiVersion"`
	Hostname    string            `json:"hostname"`
	Port        string            `json:"port"`
	AccessKey   string            `json:"accessKey"`
	EntryPoints map[string]string `json:"entry_point"`
	SetCookie   bool
	cookie      string
}

type entryPointStore struct {
	EntryPoint []entryPoint `json:"entry_point"`
}
type entryPoint struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

//this method is used when we login, it load all exists entryPoints in a map object.
func (s *Smc) loadEntryPoints() error {
	endPoint := fmt.Sprintf("http://%s:%s/%s/api", s.Hostname, s.Port, s.APIVersion)
	smcRequest := httpClient.SmcRequest{
		MethodName: "GET",
		Url:        endPoint,
		BodyData:   nil,
		Headers:    nil,
		RequestObj: nil,
	}
	if err := smcRequest.GenerateRequest(); err != nil {
		return err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("received  unexpected http status code"))
	}
	var entryPointStore entryPointStore
	if err := json.NewDecoder(response.Body).Decode(&entryPointStore); err != nil {
		return errors.New("failed in decoding EntryPoints")
	}
	if s.EntryPoints == nil {
		s.EntryPoints = make(map[string]string)
	}
	for _, entry := range entryPointStore.EntryPoint {
		s.EntryPoints[entry.Rel] = entry.Href
	}
	return nil
}

// the login function: login into an SMC instance and open a http session
// once this function is been called, Smc.setCookie will be True and Smc.Cookie will contain a valid cookie
func (s *Smc) Login() error {
	//chick if there is already an open session.
	if s.SetCookie && s.cookie != "" {
		return nil
	}
	if err := validateSmcField(s); err != nil {
		return err
	}
	endPoint := fmt.Sprintf("http://%s:%s/%s/login", s.Hostname, s.Port, s.APIVersion)

	requestBody, _ := json.Marshal(map[string]string{
		"domain":            "Shared Domain",
		"authenticationkey": s.AccessKey,
	})
	smcRequest := httpClient.SmcRequest{
		MethodName: "POST",
		Url:        endPoint,
		BodyData:   bytes.NewBuffer(requestBody),
		Headers:    nil,
		RequestObj: nil,
	}
	if err := smcRequest.AddHeader("Content-Type", "application/json").GenerateRequest(); err != nil {
		return err
	}

	resp, err := smcRequest.Run()
	if resp == nil {
		return errors.New("request timeout in login to SMC: an empty response is received ")
	}
	if err != nil {
		return errors.Wrap(err, "An error occurs during login process")
	} else {
		if resp.StatusCode != http.StatusOK {
			return errors.Wrap(err, fmt.Sprintf("unexpected http status %d received", resp.StatusCode))
		}
	}
	// read the cookie from the header
	setCookie := resp.Header.Get("Set-Cookie")
	if setCookie == "" {
		return errors.New("login response does not contain any cookies")
	}
	s.cookie = strings.Split(setCookie, ";")[0]
	s.SetCookie = true
	if err := s.loadEntryPoints(); err != nil {
		return errors.New("Failed in loading EntryPoints")
	}
	return nil
}

func retrieveLatestApiVersion(hostname string, port string) (*responses.ApiVersionResponse, error) {
	endPoint := fmt.Sprintf("http://%s:%s/api", hostname, port)
	resp, err := http.Get(endPoint)

	if resp == nil {
		return nil, errors.New("request timeout: an empty response is received ")
	}

	if err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, errors.Wrap(err, fmt.Sprintf("unexpected http status %d received", resp.StatusCode))
		}
	}

	apiVersionResponse := &responses.ApiVersionResponse{}

	err = utils.ParseResponseToStruct(resp.Body, apiVersionResponse)

	if err != nil {
		return nil, errors.New("error parsing API version response to struct")
	}

	return apiVersionResponse, nil
}

//terminate the session, and reset Smc session fields
func (s *Smc) Logout() error {
	if !s.SetCookie {
		return nil
	}
	smcRequest := httpClient.SmcRequest{
		MethodName: "PUT",
		Url:        s.EntryPoints["logout"],
		BodyData:   nil,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	if err := smcRequest.GenerateRequest(); err != nil {
		return err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNoContent {
		s.cookie = ""
		s.SetCookie = false
		s.EntryPoints = nil
	}
	return nil

}

//validate the Smc fields
func validateSmcField(s *Smc) error {
	if s.Hostname == "" {
		return errors.New("the Field HostName cannot be empty")
	}
	if s.Port == "" {
		return errors.New("the Field Port cannot be empty")
	}
	if s.APIVersion == "" {
		return errors.New("the Field APIVersion cannot be empty")
	}
	return nil
}

//Get all exists admins
func (s *Smc) GetAllAdmins() (io.Reader, error) {
	smcRequest := httpClient.SmcRequest{
		MethodName: "GET",
		Url:        s.EntryPoints["admin_user"],
		BodyData:   nil,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Failed in requesting the admins from Smc with http status: %d",
			response.StatusCode))
	}
	return response.Body, nil
}

// query an GET HTTP request
func (s *Smc) GetHttp(url string) (*http.Response, error) {
	smcRequest := httpClient.SmcRequest{
		MethodName: "GET",
		Url:        url,
		BodyData:   nil,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Create a POST HTTP request
func (s *Smc) PostHttp(url string, bodyData io.Reader) (*http.Response, error) {
	smcRequest := httpClient.SmcRequest{
		MethodName: "POST",
		Url:        url,
		BodyData:   bodyData,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *Smc) CreateAdmin(user *UserCreation) (io.Reader, int, error) {
	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	userBuffer := bytes.NewBuffer(userBytes)
	smcRequest := httpClient.SmcRequest{
		MethodName: "POST",
		Url:        s.EntryPoints["admin_user"],
		BodyData:   userBuffer,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	smcRequest.AddHeader("Content-Type", "application/json")
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, http.StatusBadRequest, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return response.Body, response.StatusCode, nil
}

// find LDAP Authentication method --> authentication_service/2
func (s *Smc) FindExternalLdap() (map[string]string, error) {
	smcRequest := httpClient.SmcRequest{
		MethodName: "GET",
		Url:        s.EntryPoints["authentication_service"],
		BodyData:   nil,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return nil, err
	}
	authServices, err := utils.ResponseToMap(response.Body)
	if err != nil {
		return nil, err
	}
	resutl := authServices["result"]
	for _, service := range resutl {
		if service["name"] == "LDAP Authentication" {
			return service, nil
		}

	}
	return nil, errors.New("no LDAP service found")
}

func (s *Smc) ExternalLdapDomain(domainName string) (map[string]string, error) {
	smcRequest := httpClient.SmcRequest{
		MethodName: "GET",
		Url:        s.EntryPoints["external_ldap_user_domain"],
		BodyData:   nil,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return nil, err
	}
	authServices, err := utils.ResponseToMap(response.Body)
	if err != nil {
		return nil, err
	}
	resutl := authServices["result"]
	for _, domain := range resutl {
		if domain["name"] == domainName {
			return domain, nil
		}
	}
	return nil, errors.New("no LDAP Domain found")
}

func (s *Smc) FindExternalActiveDirectory(domainName string) (map[string]string, error) {
	smcRequest := httpClient.SmcRequest{
		MethodName: "GET",
		Url:        s.EntryPoints["active_directory_server"],
		BodyData:   nil,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return nil, err
	}
	authServices, err := utils.ResponseToMap(response.Body)
	if err != nil {
		return nil, err
	}
	resutl := authServices["result"]
	for _, domain := range resutl {
		if domain["name"] == domainName {
			return domain, nil
		}
	}
	return nil, errors.New("no LDAP Domain found")
}

func (s *Smc) FindAllGroups(urlExternalGroup string) ([]map[string]string, error) {
	var result []map[string]string
	urlExternalGroup = urlExternalGroup + "/browse"
	resp, err := s.GetHttp(urlExternalGroup)
	if resp == nil {
		return nil, errors.New("got null point response")
	}
	if err != nil {

	}
	respData, err := utils.ResponseToMap(resp.Body)
	if err != nil {

	}
	for _, element := range respData["result"] {
		if element["type"] == "external_ldap_user_group" {
			result = append(result, element)
		}
	}
	return result, nil
}

func (s *Smc) FindAllUsers(urlExternalGroup string) ([]map[string]string, error) {
	var result []map[string]string
	urlExternalGroup = urlExternalGroup + "/browse"
	resp, err := s.GetHttp(urlExternalGroup)
	if resp == nil {
		return nil, errors.New("got null point response")
	}
	if err != nil {
		return nil, err
	}
	respData, err := utils.ResponseToMap(resp.Body)
	if err != nil {
		return nil, err
	}
	for _, element := range respData["result"] {
		if element["type"] == "external_ldap_user" {
			result = append(result, element)
		}
	}
	return result, nil
}

func (s *Smc) ExternalAldapUser(aldapUrl string) (LDAPUser, error) {
	var ldapUser LDAPUser
	user, err := s.GetHttp(aldapUrl)
	if user == nil {
		return ldapUser, errors.New("got null point response")
	}
	if err != nil {

	}
	buff, err := ioutil.ReadAll(user.Body)
	if err != nil {
		return ldapUser, err
	}
	if err := json.Unmarshal(buff, &ldapUser); err != nil {
		return ldapUser, err
	}
	return ldapUser, nil
}

//Disable or enable a user.
func (s *Smc) DisableEnableUser(userName string, userUrl string) (*http.Response, error) {
	name := strings.ReplaceAll(userName, " ", "+")
	url := fmt.Sprintf("http://%s:%s/%s/elements?filter=%s&filter_context=admin_user&exact_match=True",
		s.Hostname, s.Port, s.APIVersion, name)
	resp, err := s.GetHttp(url)
	if err != nil {
		return nil, err
	}
	etag := resp.Header.Get("Etag")
	//find the user object
	resp, err = s.GetHttp(userUrl)
	if err != nil {
		return nil, err
	}
	smcRequest := httpClient.SmcRequest{
		MethodName: "PUT",
		Url:        userUrl + "/enable_disable",
		BodyData:   resp.Body,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	smcRequest.AddHeader("Content-Type", "application/json")
	smcRequest.AddHeader("Etag", etag)
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return response, err
	}
	return response, nil
}

// update a user
func (s *Smc) UpdateUser(user *UserData) (*http.Response, error) {
	url := fmt.Sprintf("%s/%d", s.EntryPoints["admin_user"], user.Key)
	resp, err := s.GetHttp(url)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	etag := resp.Header.Get("Etag")
	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, errors.New("Failed in marshalling")
	}
	userBuffer := bytes.NewBuffer(userBytes)
	smcRequest := httpClient.SmcRequest{
		MethodName: "PUT",
		Url:        url,
		BodyData:   userBuffer,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Etag", etag)
	smcRequest.AddHeader("Cookie", s.cookie)
	smcRequest.AddHeader("Content-Type", "application/json")
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Smc) CreateActiveDirectoryLdap(ad *ActiveDirectoryLDAPS) (*http.Response, error) {
	userBytes, err := json.Marshal(ad)
	if err != nil {
		return nil, err
	}
	userBuffer := bytes.NewBuffer(userBytes)
	smcRequest := httpClient.SmcRequest{
		MethodName: "POST",
		Url:        s.EntryPoints["active_directory_server"],
		BodyData:   userBuffer,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	smcRequest.AddHeader("Content-Type", "application/json")
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Smc) CreateLdapExternalUser(ad *ExternalLDAPUser) (*http.Response, error) {
	userBytes, err := json.Marshal(ad)
	if err != nil {
		return nil, err
	}
	userBuffer := bytes.NewBuffer(userBytes)
	smcRequest := httpClient.SmcRequest{
		MethodName: "POST",
		Url:        s.EntryPoints["external_ldap_user_domain"],
		BodyData:   userBuffer,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("Cookie", s.cookie)
	smcRequest.AddHeader("Content-Type", "application/json")
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return response, err
	}
	return response, nil
}

func (s *Smc) DeleteAdmin(adminName string) (*http.Response, error) {
	if !s.SetCookie {
		return nil, errors.New("no active session exists with SMC instance, please login to SMC instance first")
	}
	body, err := s.GetAllAdmins()
	if err != nil {
		return nil, err
	}
	userHref := ""
	result, _ := utils.ResponseToMap(body)
	for _, users := range result {
		for _, user := range users {
			if user["name"] == adminName {
				userHref = user["href"]
				break
			}
		}
	}
	if userHref == "" {
		return nil, errors.New("the given adminName is not exist")
	}
	resp, err := s.GetHttp(userHref)
	if err != nil {
		return nil, err
	}
	etag := resp.Header.Get("Etag")
	smcRequest := httpClient.SmcRequest{
		MethodName: "DELETE",
		Url:        userHref,
		BodyData:   nil,
		Headers:    nil,
		RequestObj: nil,
	}
	smcRequest.AddHeader("if-match", etag)
	smcRequest.AddHeader("Cookie", s.cookie)
	if err := smcRequest.GenerateRequest(); err != nil {
		return nil, err
	}
	response, err := smcRequest.Run()
	if err != nil {
		return response, err
	}
	return response, nil
}
