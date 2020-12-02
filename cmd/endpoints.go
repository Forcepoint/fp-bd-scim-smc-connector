package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var Routes = []Route{
	{
		Name:        "CreateUser",
		Method:      "POST",
		Pattern:     "/api/v1/CreateUsers",
		HandlerFunc: AddUser,
	},
	{
		Name:        "UpdateUser",
		Method:      "POST",
		Pattern:     "/api/v1/UpdateUsers",
		HandlerFunc: UpdateUser,
	},
	{
		Name:        "GetUsers",
		Method:      "GET",
		Pattern:     "/api/v1/Users",
		HandlerFunc: GetUsers,
	},

	{
		Name:        "DeleteUser",
		Method:      "DELETE",
		Pattern:     "/api/v1/Users/{id}",
		HandlerFunc: DeleteUser,
	},

	{
		Name:        "entrypoints",
		Method:      "GET",
		Pattern:     "/api/v1/Entypoints",
		HandlerFunc: EntryPoints,
	},
	{
		Name:        "TokenPermission",
		Method:      "POST",
		Pattern:     "/api/v1/TokenPermission",
		HandlerFunc: TokenPermission,
	},
}
var RoutesCopy []Route

func AddRoutes(router *mux.Router) *mux.Router {
	for _, route := range Routes {
		router.Methods(route.Method).Path(route.Pattern).Handler(route.HandlerFunc)
		RoutesCopy = append(RoutesCopy, route)
	}
	return router
}

func GetEntryPoints() map[string]string {
	entryPoints := make(map[string]string)
	for _, route := range RoutesCopy {
		entryPoints[route.Name] = fmt.Sprintf("http://%s:%s%s",
			viper.GetString("connector.hostname"), viper.GetString("connector.port"), route.Pattern)
	}
	return entryPoints
}
