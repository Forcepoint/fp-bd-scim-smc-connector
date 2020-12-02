package smc

import (
	"fmt"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/interfaces"
	"github.com/creasty/defaults"
	"net/http"
)

func (s *Smc) GetElements(element interfaces.SmcElement) (*http.Response, error) {
	//This sets the default 'TypeOf' value in the struct to allow pulling the url
	//from the entrypoint map in the SMC client
	if err := defaults.Set(element); err != nil {
		return nil, err
	}
	return element.Get(s.EntryPoints[element.GetTypeOf()], s.cookie)
}

func (s *Smc) GetElement(element interfaces.SmcElement, id int) (*http.Response, error) {
	//This sets the default 'TypeOf' value in the struct to allow pulling the url
	//from the entrypoint map in the SMC client
	if err := defaults.Set(element); err != nil {
		return nil, err
	}
	return element.Get(fmt.Sprintf(s.EntryPoints[element.GetTypeOf()]+"/%d", id), s.cookie)
}

func (s *Smc) GetSubElements(parentElement, subElement interfaces.SmcElement, parentId int) (*http.Response, error) {
	if err := defaults.Set(parentElement); err != nil {
		return nil, err
	}
	if err := defaults.Set(subElement); err != nil {
		return nil, err
	}

	getUrl := fmt.Sprintf("%s/%d/%s", s.EntryPoints[parentElement.GetTypeOf()], parentId, subElement.GetTypeOf())

	return parentElement.GetSubElements(getUrl, s.cookie)
}

func (s *Smc) CreateElement(element interfaces.SmcElement) (*http.Response, error) {
	if err := defaults.Set(element); err != nil {
		return nil, err
	}
	return element.Create(s.EntryPoints[element.GetTypeOf()], s.cookie)
}

func (s *Smc) CreateSubElement(parentElement, subElement interfaces.SmcElement, parentId int) (*http.Response, error) {
	if err := defaults.Set(parentElement); err != nil {
		return nil, err
	}
	if err := defaults.Set(subElement); err != nil {
		return nil, err
	}

	createUrl := fmt.Sprintf("%s/%d/%s", s.EntryPoints[parentElement.GetTypeOf()], parentId, subElement.GetTypeOf())

	return subElement.CreateSubElement(createUrl, s.cookie)
}

func (s *Smc) UpdateElement(element interfaces.SmcElement) (*http.Response, error) {
	if err := defaults.Set(element); err != nil {
		return nil, err
	}
	return element.Update(s.EntryPoints[element.GetTypeOf()], s.cookie)
}

func (s *Smc) UpdateSubElement(parentElement, subElement interfaces.SmcElement, parentId int) (*http.Response, error) {
	if err := defaults.Set(parentElement); err != nil {
		return nil, err
	}
	if err := defaults.Set(subElement); err != nil {
		return nil, err
	}

	updateUrl := fmt.Sprintf("%s/%d/%s", s.EntryPoints[parentElement.GetTypeOf()], parentId, subElement.GetTypeOf())

	return subElement.UpdateSubElement(updateUrl, s.cookie)
}

func (s *Smc) DeleteElement(element interfaces.SmcElement) (*http.Response, error) {
	if err := defaults.Set(element); err != nil {
		return nil, err
	}
	return element.Delete(s.EntryPoints[element.GetTypeOf()], s.cookie)
}

func (s *Smc) DeleteSubElement(parentElement, subElement interfaces.SmcElement, parentId int) (*http.Response, error) {
	if err := defaults.Set(parentElement); err != nil {
		return nil, err
	}
	if err := defaults.Set(subElement); err != nil {
		return nil, err
	}

	deleteUrl := fmt.Sprintf("%s/%d/%s", s.EntryPoints[parentElement.GetTypeOf()], parentId, subElement.GetTypeOf())

	return subElement.DeleteSubElement(deleteUrl, s.cookie)
}
