package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/smc"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/utils"
	"github.cicd.cloud.fpdev.io/BD/scim-smc-connector/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	emptyString = ""
)

// get all SMC users
func SmcUsers(id string) ([]map[string]string, error) {
	var users []map[string]string
	err := SmcInstance.Login()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	body, err := SmcInstance.GetAllAdmins()
	if err != nil {
		return users, err
	}
	result, _ := utils.ResponseToMap(body)
	admins := result["result"]
	if id == emptyString {
		users = admins
	} else {
		if strings.Contains(id, "@") {
			parts := strings.Split(id, "@")
			id = parts[0]
		}
		for _, admin := range admins {
			adminUrl := admin["href"]
			userInfo, err := GetUserSMCInfo(adminUrl)
			if err != nil {
				return users, err
			}
			parts := strings.Split(userInfo.LdapUser, "/")
			adminId := parts[len(parts)-1]
			if admin["name"] == id || adminId == id {
				users = append(users, admin)
				break
			}
		}
	}
	err = SmcInstance.Logout()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	return users, nil
}

// extract user's info from SMC
func SmcUsersWithDetails(users []map[string]string) ([]UserInfo, error) {
	err := SmcInstance.Login()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	var usersInfo []UserInfo
	for _, v := range users {
		href := v["href"]
		body, err := SmcInstance.GetHttp(href)
		if err != nil {
			return usersInfo, err
		}
		buff, err := ioutil.ReadAll(body.Body)
		if err != nil {
			return usersInfo, err
		}
		var info UserInfo
		if err := json.Unmarshal(buff, &info); err != nil {
			return usersInfo, err
		}
		usersInfo = append(usersInfo, info)
	}
	err = SmcInstance.Logout()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	return usersInfo, nil
}

//validate if a given username is exist in SMC
func validateUser(userName string) (bool, error) {
	users, err := SmcUsers(userName)
	if err != nil {
		return false, err
	}
	if users == nil || len(users) == 0 {
		return false, errors.New("user does not exists: " + userName)
	}
	if len(users) > 1 {
		return false, errors.New("more then one user exists with the same user name")
	}
	return true, nil
}

// create a new user
func CreateUser(userName string, active bool) (string, int, error) {
	var returnError error
	returnError = nil
	httpStatus := http.StatusCreated
	err := SmcInstance.Login()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	var userLdap smc.LDAPUser
	//find external LDAP Auth
	ldapAuthService, err := SmcInstance.FindExternalLdap()
	if err != nil {
		returnError = err
		httpStatus = http.StatusBadRequest
		return "", httpStatus, returnError
	}
	authMethod := ldapAuthService["href"]
	ldapDomain, err := SmcInstance.ExternalLdapDomain(viper.GetString("LDAP_DOMAIN"))
	if err != nil {
		returnError = err
		httpStatus = http.StatusBadRequest
		return "", httpStatus, returnError
	}
	url := ldapDomain["href"] + "/browse"
	azureAd, err := SmcInstance.GetHttp(url)
	rep, err := utils.ResponseToMap(azureAd.Body)
	if err != nil {
		returnError = err
		httpStatus = http.StatusBadRequest
		return "", httpStatus, returnError
	}
	for _, r := range rep["result"] {
		if r["name"] == "AADDC Users" {
			users, _ := SmcInstance.FindAllUsers(r["href"])
			for _, user := range users {
				u, _ := SmcInstance.ExternalAldapUser(user["href"])
				if u.Name == userName {
					userLdap = u
				}
			}
		}
	}

	var userHref string
	for _, l := range userLdap.Link {
		if l["rel"] == "self" {
			userHref = l["href"]
		}
	}
	permissions, superUser := generateDefaultPermissions()
	user := smc.UserCreation{
		Name:                   userName,
		Enabled:                active,
		AllowSudo:              viper.GetBool("ROLES.ALLOW_SUDO"),
		ConsoleSuperuser:       viper.GetBool("ROLES.CONSOLE_SUPPER_USER"),
		AllowedToLoginInShared: viper.GetBool("ROLES.ALLOW_TO_LOGS_IN_SHARED"),
		EngineTarget:           []string{},
		LocalAdmin:             false,
		Superuser:              superUser,
		CanUseApi:              viper.GetBool("ROLES.CAN_USE_API"),
		Comment:                nil,
		AuthMethod:             authMethod,
		LdapUser:               userHref,
		Permissions:            permissions,
	}

	if !SmcInstance.SetCookie {
		err := SmcInstance.Login()
		if err != nil {
			logrus.Fatal(err.Error())
		}
	}
	_, httpStatus, err = SmcInstance.CreateAdmin(&user)
	if err != nil {
		returnError = err
		logrus.Debug("Error4: " + err.Error())
	}
	if httpStatus == http.StatusUnprocessableEntity {
		returnError = errors.New(fmt.Sprintf("User name %s is already exist", userName))
	}
	if httpStatus != http.StatusCreated {
		returnError = errors.New(fmt.Sprintf("unexpected http status code: %d is recieved ", httpStatus))
	}
	err = SmcInstance.Logout()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	return userHref, httpStatus, returnError
}

// enable or disable a user
func EnableDisableUser(userId string) (bool, error) {
	user, err := SmcUsers(userId)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("user not found")
	}
	err = SmcInstance.Login()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	for _, u := range user {
		if u["name"] == userId {
			response, err := SmcInstance.DisableEnableUser(userId, u["href"])
			if err != nil {
				return false, err
			}
			if response.StatusCode != http.StatusOK {
				return false, errors.New(fmt.Sprintf("unexpected http status: %d ", response.StatusCode))
			}
		}
	}
	err = SmcInstance.Logout()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	return true, nil
}

// load all exists SMC roles which can be assigned to a user
func GetRoles() (map[string]string, error) {
	smcRoles := make(map[string][]map[string]string)
	roles := make(map[string]string)
	response, err := SmcInstance.GetHttp(fmt.Sprintf("http://%s:%s/%s/elements/role",
		SmcInstance.Hostname, SmcInstance.Port, SmcInstance.APIVersion))
	if err != nil {
		return roles, err
	}
	if response == nil {
		return roles, err
	}
	if response.StatusCode == http.StatusOK {

		buff, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return roles, err
		}
		if err := json.Unmarshal(buff, &smcRoles); err != nil {
			return roles, err
		}
	}
	for _, role := range smcRoles["result"] {
		roles[role["name"]] = role["href"]
	}

	return roles, nil
}

// get users name from SMC
func GetAllUsersName() ([]string, error) {
	var users []string
	body, err := SmcInstance.GetAllAdmins()
	if err != nil {
		return users, err
	}
	result, err := utils.ResponseToMap(body)
	if err != nil {
		return users, err
	}
	for _, user := range result["result"] {
		users = append(users, user["name"])
	}
	return users, nil
}

// map between users and roles
func MapUsersToRoles() (map[string][]string, error) {
	usersWithRoles := make(map[string][]string)
	groupNames := strings.Split("Editor,Operator,Owner,Viewer,Superuser,NSX Role,Logs Viewer,Reports Manager,Monitor",
		",")
	smcUsers, err := GetAllUsersName()
	if err != nil {

	}
	ldapDom, err := SmcInstance.ExternalLdapDomain(viper.GetString("LDAP_DOMAIN"))
	if err != nil {
		return usersWithRoles, err
	}
	url := ldapDom["href"] + "/browse"
	azureAd, err := SmcInstance.GetHttp(url)
	if err != nil {
		return usersWithRoles, err
	}
	if azureAd == nil {
		return usersWithRoles, errors.New("got an empty response while mapping users to roles")
	}
	rep, _ := utils.ResponseToMap(azureAd.Body)
	for _, r := range rep["result"] {
		// find all groups and users of azure AD
		if r["name"] == "AADDC Users" {
			groups, _ := SmcInstance.FindAllGroups(r["href"])
			for _, group := range groups {
				if !lib.StringInSlice(group["name"], groupNames) {
					continue
				}
				users, _ := SmcInstance.FindAllUsers(group["href"])
				for _, user := range users {
					u, _ := SmcInstance.ExternalAldapUser(user["href"])
					//is user is a SMC user
					if lib.StringInSlice(u.Name, smcUsers) {
						if _, ok := usersWithRoles[u.Name]; !ok {
							usersWithRoles[u.Name] = []string{}
						}
						usersWithRoles[u.Name] = append(usersWithRoles[u.Name], group["name"])
					}
				}
			}
		}
	}
	return usersWithRoles, nil
}

// this function will be called in a goroutine, the goal of this function is to read all azure ldap groups and apply
// the required roles on their members.
func ApplyRoles() {
	if !SmcInstance.SetCookie {
		err := SmcInstance.Login()
		if err != nil {
			logrus.Fatal(err.Error())
		}
	}
	roles, err := GetRoles()
	if err != nil || len(roles) == 0 {
		time.Sleep(2 * time.Minute)
		roles, err = GetRoles()
		if err != nil {
			logrus.Errorf("Error occur in loading all roles. Error: %s", err)
			return
		}
	}

	usersUrl := make(map[string]string)
	admins, _ := SmcInstance.GetAllAdmins()
	if err != nil {
		logrus.Errorf("Error occur in updating the user's roles. Error: %s", err)
		return

	}
	if admins == nil {
		logrus.Error("Failed to communicate with Forcepoint SMC in order to load exists users. ensure the hostname and the --key are correct")
		return
	}
	result, _ := utils.ResponseToMap(admins)
	for _, u := range result["result"] {
		usersUrl[u["name"]] = u["href"]
	}

	usersWithRoles, err := MapUsersToRoles()
	for user, userRole := range usersWithRoles {
		var appliedRoles []string
		if len(userRole) != 0 {
			permissions := make(map[string][]smc.Permission)
			permissions["permission"] = []smc.Permission{}
			userData, err := GetUserData(usersUrl[user])
			if err != nil {
				logrus.Errorf("Error occur in loading user's SMC data. Error: %s", err)
				return
			}
			for _, role := range userRole {
				roleUrl := roles[role]
				if role == "Superuser" {
					grantedElements := fmt.Sprintf("http://%s:%s/%s/elements/access_control_list/7",
						SmcInstance.Hostname,
						SmcInstance.Port,
						SmcInstance.APIVersion)
					permission := smc.Permission{
						GrantedDomainRef: fmt.Sprintf("http://%s:%s/%s/elements/admin_domain/1",
							SmcInstance.Hostname,
							SmcInstance.Port,
							SmcInstance.APIVersion),
						GrantedElements: []string{grantedElements},
						RoleRef:         roleUrl,
					}
					permissions["permission"] = make([]smc.Permission, 0)
					permissions["permission"] = append(permissions["permission"], permission)
					userData.Superuser = true
					appliedRoles = []string{role}
					break
				} else {
					grantedElements := fmt.Sprintf("http://%s:%s/%s/elements/access_control_list/7",
						SmcInstance.Hostname,
						SmcInstance.Port,
						SmcInstance.APIVersion)
					permission := smc.Permission{
						GrantedDomainRef: fmt.Sprintf("http://%s:%s/%s/elements/admin_domain/1",
							SmcInstance.Hostname,
							SmcInstance.Port,
							SmcInstance.APIVersion),
						GrantedElements: []string{grantedElements},
						RoleRef:         roleUrl,
					}
					permissions["permission"] = append(permissions["permission"], permission)
				}
				userData.Superuser = false
				appliedRoles = append(appliedRoles, role)
			}
			result := lib.CompareRoles(userData.Permissions["permission"], permissions["permission"])
			if !result {
				userData.Permissions = permissions
				response, err := SmcInstance.UpdateUser(&userData)
				if err != nil {
					logrus.Errorf("Error occur in updating the user's roles. Error: %s", err)
				}
				if response.StatusCode != http.StatusOK {
					logrus.Errorf("Error occur in updating the user's roles. http status code: %d",
						response.StatusCode)
				} else {
					newRoles := strings.Join(appliedRoles, ", ")
					logrus.Infof("new roles: user=%s, roles: %s", user, newRoles)
				}

				time.Sleep(1 * time.Second)
			}
		}
	}

	err = SmcInstance.Logout()
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

// get all info related to a SMC user
func GetUserData(userUrl string) (smc.UserData, error) {
	var userData smc.UserData
	response, err := SmcInstance.GetHttp(userUrl)
	if response == nil {
		return userData, err
	}
	if err != nil {
		return userData, err
	}
	buffer, err := ioutil.ReadAll(response.Body)
	if err := json.Unmarshal(buffer, &userData); err != nil {
		return userData, err
	}
	return userData, nil
}

// generate the defaults roles and permissions for a new users.
// the default roles and permissions can be defined in the config file
func generateDefaultPermissions() (map[string][]smc.Permission, bool) {
	perNames := []string{"VIEWER", "LOGS_VIEWER",
		"REPORTS_MANAGER", "OWNER", "OPERATOR", "MONITOR", "EDITOR",
		"NSX_ROLE"}
	permissions := make(map[string][]smc.Permission)
	permissions["permission"] = []smc.Permission{}
	if !SmcInstance.SetCookie {
		err := SmcInstance.Login()
		if err != nil {
			logrus.Fatal(err.Error())
		}
	}
	roles, err := GetRoles()
	if err != nil {
		logrus.Errorf("Error in getting all exist roles from SMC: %s", err.Error())
	}
	err = SmcInstance.Logout()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	if viper.GetBool("ROLES.PERMISSIONS.SUPPER_USER") {
		grantedElements := fmt.Sprintf("http://%s:%s/%s/elements/access_control_list/7",
			SmcInstance.Hostname,
			SmcInstance.Port,
			SmcInstance.APIVersion)
		permission := smc.Permission{
			GrantedDomainRef: fmt.Sprintf("http://%s:%s/%s/elements/admin_domain/1",
				SmcInstance.Hostname,
				SmcInstance.Port,
				SmcInstance.APIVersion),
			GrantedElements: []string{grantedElements},
			RoleRef:         roles["Superuser"],
		}
		permissions["permission"] = append(permissions["permission"], permission)
		return permissions, true
	} else {
		for _, p := range perNames {
			viperName := fmt.Sprintf("ROLES.PERMISSIONS.%s", p)
			if viper.GetBool(viperName) {
				roleName := strings.ReplaceAll(p, "_", " ")
				roleName = strings.ToLower(roleName)
				roleName = strings.Title(roleName)
				grantedElements := fmt.Sprintf("http://%s:%s/%s/elements/access_control_list/7",
					SmcInstance.Hostname,
					SmcInstance.Port,
					SmcInstance.APIVersion)
				permission := smc.Permission{
					GrantedDomainRef: fmt.Sprintf("http://%s:%s/%s/elements/admin_domain/1",
						SmcInstance.Hostname,
						SmcInstance.Port,
						SmcInstance.APIVersion),
					GrantedElements: []string{grantedElements},
					RoleRef:         roles[roleName],
				}
				permissions["permission"] = append(permissions["permission"], permission)
			}
		}
	}
	return permissions, false
}

func GetUserSMCInfo(href string) (UserInfo, error) {
	err := SmcInstance.Login()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	var usersInfo UserInfo
	body, err := SmcInstance.GetHttp(href)
	if err != nil {
		return usersInfo, err
	}
	buff, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return usersInfo, err
	}
	if err := json.Unmarshal(buff, &usersInfo); err != nil {
		return usersInfo, err
	}
	err = SmcInstance.Logout()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	return usersInfo, nil
}

func DeleteSmcUser(userName string) error {
	err := SmcInstance.Login()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	defer SmcInstance.Logout()
	resp, err := SmcInstance.DeleteAdmin(userName)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(fmt.Sprintf("got unexpected http status code: %d from deleting a user", resp.StatusCode))
	}
	return nil
}

func DetectDeletedUsers() error {
	err := SmcInstance.Login()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	defer SmcInstance.Logout()
	assignedUsers, err := GetAppAssignedUsers(viper.GetString("APP_NAME"))
	if err != nil {
		return err
	}
	assignedUserNames, err := GetAssignedUsersNames(assignedUsers)
	if err != nil {
		return err
	}
	allAzureUsers, err := GetAllAzureUsers()
	if err != nil {
		return err
	}
	allUsers, err := SmcUsers("")
	if err != nil {
		return err
	}
	var smcUsers []string
	for _, u := range allUsers {
		smcUsers = append(smcUsers, u["name"])
	}
	if err := deleteUsers(assignedUserNames, allAzureUsers, allUsers); err != nil {
		return err
	}
	return nil

}

func deleteUsers(assignedUserNames []string, allAzureUsers []string, smcUsers []map[string]string) error {
	var smcUsersNames []string
	for _, u := range smcUsers {
		smcUsersNames = append(smcUsersNames, u["name"])
	}
	for _, name := range smcUsersNames {
		if !InSlice(assignedUserNames, name) && InSlice(allAzureUsers, name) {
			if err := DeleteSmcUser(name); err != nil {
				return err
			}
			logrus.Infof("User %s is been deleted", name)
		}
	}
	return nil
}

func InSlice(list []string, name string) bool {
	for _, i := range list {
		if i == name {
			return true
		}
	}
	return false
}
