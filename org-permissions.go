package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	apex "github.com/apex/log"
)

type orgPermissions struct {
	Groups map[string]string `json:"groups"`
	Users  map[string]string `json:"users"`
}

type permissionConfig struct {
	DefaultViewOrg []string                  `json:"default-viewer-org"`
	OrgPermissions map[string]orgPermissions `json:"org-permissions"`
}

func readConfig() (permissionConfig, error) {
	jsonFile, err := os.Open("./grafana_orgs/org-permissions.json")
	defer jsonFile.Close()

	var result permissionConfig
	// if we os.Open returns an error then handle it
	if err != nil {
		var result permissionConfig
		return result, err

	} else {
		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return result, err
		}
		err = json.Unmarshal([]byte(byteValue), &result)
		if err != nil {
			return result, err
		}
		return result, nil
	}
}

func getUserPermission(loginId string, logs *apex.Entry) (map[string]string, error) {

	// Mapping roles to values - for finding appropriate roles based on the google-group memberships
	var roles = make(map[string]int64)
	roles["Viewer"] = 1
	roles["Editor"] = 2
	roles["Admin"] = 3

	var config permissionConfig
	var userGroups []string

	config, err := readConfig()
	if err != nil {
		logs.Error("Error in reading org-permission config-file: " + err.Error())
		return nil, err
	}

	userGroups, err = getGoogleGroups(loginId, logs)
	if err != nil {
		logs.Error("Error in getting google groups: " + err.Error())
		return nil, err
	}

	// counter for checking the user has a role in any one organization
	flag := 0

	// Find the highest role the user is supposed to get
	orgRoles := make(map[string]string)
	for org, perms := range config.OrgPermissions {

		if roles[perms.Users[loginId]] > 0 {
			orgRoles[org] = perms.Users[loginId]
			// flag = 1 when the user has some role in any one organization
			flag = 1

		} else {
			var maxRoleId int64 = 0
			var maxRole string = "delete-user"

			var groups map[string]string = perms.Groups
			for i := 0; i < len(userGroups); i++ {
				group := userGroups[i]
				if roles[groups[group]] > maxRoleId {
					maxRoleId = roles[groups[group]]
					maxRole = groups[group]
				}
			}

			orgRoles[org] = maxRole
			if maxRoleId != 0 {
				// flag = 1 when the user has some role in any one of the organizations
				flag = 1
			}
		}
	}

	// if the user has no roles in any Grafana organizations
	// grant the user "viewer" role in the defaultViewOrg list
	if flag == 0 {
		for _, ele := range config.DefaultViewOrg {
			orgRoles[ele] = "Viewer"
		}
	}

	// for logging purposes : print the user org-roles
	for org, role := range orgRoles {
		logs.Info(org + ":" + role)
	}

	return orgRoles, nil
}
