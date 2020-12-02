/*
An interface for an endpoint instance of SMC
author: Dlo Bagari
date:12/02/2020
*/

package interfaces

// an interface for endpoint. an endpoint would have the defined behaviors
type EndPoint interface {
	Login() error
	Logout() error
}
