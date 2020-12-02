package cmd

import (
	"bytes"
	"errors"
	"fmt"
	errorWrapper "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"os/exec"
	"strings"
	"syscall"
)

type AzureCLI struct {
	IsLogin bool
}

// login to azure
func (a AzureCLI) Login() error {
	if !a.IsLogin {
		var stdout, stderr bytes.Buffer
		if viper.GetString("AZURE_ADMIN_LOGIN_NAME") == "" {
			logrus.Fatal("the field AZURE_ADMIN_LOGIN_NAME in the config file is empty. Please add your azure administrator login name to the config file")
		}
		if viper.GetString("AZURE_ADMIN_LOGIN_PASSWORD") == "" {
			fmt.Printf("Enter the current password for '%s' and press Enter: ",
				viper.GetString("AZURE_ADMIN_LOGIN_NAME"))
			bytePassword, err := terminal.ReadPassword(syscall.Stdin)
			if err != nil {
				return err
			}
			password := string(bytePassword)
			fmt.Println() // do not remove it
			if len(password) == 0 {
				return errors.New("please enter a valid password")
			}
			viper.Set("AZURE_ADMIN_LOGIN_PASSWORD", strings.TrimSpace(password))
		}
		//login to azure
		c1 := fmt.Sprintf("az login -u %s -p '%s'",
			viper.GetString("AZURE_ADMIN_LOGIN_NAME"),
			viper.GetString("AZURE_ADMIN_LOGIN_PASSWORD"))
		exe := exec.Command("sh", "-c", c1)
		exe.Stderr = &stderr
		exe.Stdout = &stdout
		err := exe.Run()
		errorResult := string(stderr.Bytes())
		if len(errorResult) != 0 {
			if strings.Contains(errorResult, "Error validating credentials due to invalid username or password") {
				return errors.New("error in validating credentials due to invalid username or password")
			}
			return errors.New(errorResult)
		}
		if err != nil {
			return errors.New("failed in executing the azure login command")
		}
		a.IsLogin = true
	}
	return nil
}

// azure logout
func (a AzureCLI) Logout() error {
	exe := exec.Command("sh", "-c", "az logout")
	err := exe.Run()
	if err != nil {
		err = errorWrapper.Wrap(err, "Failed in executing the azure logout command")
		return err
	}
	a.IsLogin = false
	return nil
}

func GetAssignedUsersNames(usersId []string) ([]string, error) {
	var names []string
	for _, id := range usersId {
		c := fmt.Sprintf("az ad user show --id %s --query mailNickname -o tsv", id)
		output, err := ExecuteCmd(c)
		if err != nil {
			return nil, err
		}
		output = strings.TrimSpace(output)
		names = append(names, output)
	}
	return names, nil
}

func GetAppAssignedUsers(appName string) ([]string, error) {
	c := fmt.Sprintf("az ad sp list --display-name '%s' --query [].objectId -o tsv", appName)
	appId, err := ExecuteCmd(c)
	if err != nil {
		return nil, err
	}
	appId = strings.TrimSpace(appId)
	if appId == "" {
		return nil, errors.New("failed in reading the app id")
	}
	c = fmt.Sprintf("az rest --method GET --uri https://graph.microsoft.com/beta/servicePrincipals/%s/appRoleAssignedTo --query value[].principalId -o tsv", appId)
	output, err := ExecuteCmd(c)
	if err != nil {
		return nil, err
	}
	output = strings.TrimSpace(output)
	return strings.Split(output, "\n"), nil
}

func ExecuteCmd(cmd string) (string, error) {
	var stdout, stderr bytes.Buffer
	exe := exec.Command("sh", "-c", cmd)
	exe.Stderr = &stderr
	exe.Stdout = &stdout
	err := exe.Run()
	errorResult := string(stderr.Bytes())
	if len(errorResult) != 0 && !strings.Contains(errorResult, "deprecated") {
		return "", errors.New(errorResult)
	}
	if err != nil && !strings.Contains(errorResult, "deprecated") {
		return "", errors.New(fmt.Sprintf("failed in executing the azure command: %s", cmd))
	}
	output := string(stdout.Bytes())
	if len(output) != 0 {
		return output, nil
	}
	return "", nil
}

func GetAllAzureUsers() ([]string, error) {
	c := "az ad user list --query [].mailNickname -o tsv"
	output, err := ExecuteCmd(c)
	if err != nil {
		return nil, err
	}
	output = strings.TrimSpace(output)
	return strings.Split(output, "\n"), nil
}
