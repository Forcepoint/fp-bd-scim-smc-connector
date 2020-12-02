package cmd

import (
	"encoding/json"
	"fmt"
	"github.cicd.cloud.fpdev.io/BD/scim-smc-connector/lib"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	_ "regexp"
	"strings"
)

type foundUsers struct {
	TotalUsers int        `json:"total_users"`
	Users      []UserInfo `json:"users"`
}

type ErrorTemplate struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}

type Operation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

//get a list of exists admins/admin from forcepoint SMC
func GetUsers(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	var userName string
	var smcUsers foundUsers
	var userScim []map[string]interface{}
	if filterQuery, ok := args["id"]; ok {
		userName = filterQuery[0]
	} else {
		userName = ""
	}
	users, err := SmcUsers(userName)
	if err != nil {
		loggerWithField(r).Error(err.Error())
	}
	if users == nil {
		smcUsers = foundUsers{
			TotalUsers: 0,
			Users:      nil,
		}
	} else {
		usersInfo, err := SmcUsersWithDetails(users)
		if err != nil {
			loggerWithField(r).Error(err.Error())
		}
		smcUsers = foundUsers{TotalUsers: len(usersInfo), Users: usersInfo}
		userScim = userScimInfo(smcUsers)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&userScim)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	loggerWithField(r).Infof("Get users")
	return
}

func EntryPoints(w http.ResponseWriter, r *http.Request) {
	entryPoints := GetEntryPoints()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(entryPoints)
	if err != nil {
		loggerWithField(r).Error(err.Error())
	}
	loggerWithField(r).Info("Get API EntryPoints request")
	return
}

func TokenPermission(w http.ResponseWriter, r *http.Request) {
	result := make(map[string]string)
	userInfo := struct {
		UserName    string `json:"username"`
		Password    string `json:"password"`
		ProductName string `json:"productName"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		loggerWithField(r).Error(err.Error())
	}
	if userInfo.ProductName != viper.GetString("SMC.NAME") {
		loggerWithField(r).Error(fmt.Sprintf("the given Product name (%s) does not match the connector's name",
			userInfo.ProductName))
		w.WriteHeader(http.StatusBadRequest)
		result["allow"] = "false"
		result["reason"] = "The product name is not valid"
		err := json.NewEncoder(w).Encode(result)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}

	if userInfo.Password != viper.GetString("SMC.KEY") {
		loggerWithField(r).Error("the given password is not valid")
		w.WriteHeader(http.StatusBadRequest)
		result["allow"] = "false"
		result["reason"] = "The given password is not valid"
		err := json.NewEncoder(w).Encode(result)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}
	//chick of user exists in SMC
	validUser, err := validateUser(userInfo.UserName)
	if err != nil {
		loggerWithField(r).Error(err.Error())
		if !validUser {
			w.WriteHeader(http.StatusBadRequest)
			result["allow"] = "false"
			result["reason"] = err.Error()
			err := json.NewEncoder(w).Encode(result)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return
		}
	}
	result["allow"] = "true"
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	loggerWithField(r).Infof("Give permission for getting an access token for user: %s", userInfo.UserName)
}

func loggerWithField(r *http.Request) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"RequestMethod": r.Method, "RequestURL": r.RequestURI, "RemoteAddress": r.RemoteAddr,
	})
}

func userScimInfo(users foundUsers) []map[string]interface{} {
	var userList []map[string]interface{}
	for _, u := range users.Users {
		parts := strings.Split(u.LdapUser, "/")
		userId := ""
		if len(parts) != 0 {
			userId = parts[len(parts)-1]
		}
		userMap := make(map[string]interface{})
		userMap["active"] = u.Enable
		userMap["name"] = u.Name
		userMap["id"] = userId
		userList = append(userList, userMap)
	}
	return userList
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	userInfo := struct {
		LoginName   string `json:"login_name"`
		DisplayName string `json:"display_name"`
		Active      bool   `json:"active"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		loggerWithField(r).Error(err.Error())
	}
	userName, err := lib.ExtractName(userInfo.LoginName)
	if err != nil {
		logrus.Fatal(err)
	}
	userALDAPurl, httpStatus, err := CreateUser(userName, userInfo.Active)
	if err != nil {
		loggerWithField(r).Error(err.Error())
		w.WriteHeader(httpStatus)
		return
	}
	if httpStatus != http.StatusCreated {
		w.WriteHeader(httpStatus)
		return
	}
	responseBody := make(map[string]string)
	responseBody["userUrl"] = userALDAPurl
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	err = json.NewEncoder(w).Encode(&responseBody)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	loggerWithField(r).Infof("User Created: %s", userName)

}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	updateJob := struct {
		UserId     string      `json:"user_id"`
		Operations []Operation `json:"operations"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&updateJob); err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	for _, op := range updateJob.Operations {
		if op.Op == "Replace" && op.Path == "active" {
			result, err := EnableDisableUser(updateJob.UserId)
			if err != nil && !result {
				w.WriteHeader(http.StatusUnprocessableEntity)
				loggerWithField(r).Error(err.Error())
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	loggerWithField(r).Infof("Updated User")
	return
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := vars["id"]
	users, err := SmcUsers(userName)
	if err != nil {
		loggerWithField(r).Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(users) == 0 {
		loggerWithField(r).Errorf("the given user id: %s not found", userName)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(users) > 1 {
		loggerWithField(r).Errorf("multiple users with id: %s are exists", userName)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user := users[0]
	if err := DeleteSmcUser(user["name"]); err != nil {
		loggerWithField(r).Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
