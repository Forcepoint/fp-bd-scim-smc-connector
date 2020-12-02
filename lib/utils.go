package lib

import (
	"errors"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/smc"
	"os"
	"strings"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func CompareRoles(p1 []smc.Permission, p2 []smc.Permission) bool {
	if len(p1) == 0 && len(p2) == 1 {
		parts := strings.Split(p2[0].RoleRef, "/")
		index := parts[len(parts)-1]
		if index == "1" {
			return true
		}
	}
	if len(p1) != len(p2) {
		return false
	}
	for _, p := range p1 {
		matched := false
		for _, n := range p2 {
			if p.RoleRef == n.RoleRef {
				matched = true
			}
		}
		if matched == false {
			return matched
		}
	}
	return true
}

func ExtractName(email string) (string, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", errors.New("user loginName is not an email address: " + email)
	}
	return parts[0], nil
}
