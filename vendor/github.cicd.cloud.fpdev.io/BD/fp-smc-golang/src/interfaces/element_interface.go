package interfaces

import (
	"net/http"
)

type SmcElement interface {
	Get(url, cookie string) (*http.Response, error)
	GetSubElements(url, cookie string) (*http.Response, error)
	Create(url, cookie string) (*http.Response, error)
	CreateSubElement(url, cookie string) (*http.Response, error)
	Update(url, cookie string) (*http.Response, error)
	UpdateSubElement(url, cookie string) (*http.Response, error)
	Delete(url, cookie string) (*http.Response, error)
	DeleteSubElement(url, cookie string) (*http.Response, error)
	GetTypeOf() string
}
