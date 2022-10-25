package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"

	apex "github.com/apex/log"
	sdk "github.com/grafana/grafana-api-golang-client"
)

var passwordBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890%#"

func getGrafanaClient(logs *apex.Entry) (*sdk.Client, error) {

	// Connect to Grafana API as 'admin'
	grafanaURL := os.Getenv("GRAFANA-ENDPOINT")
	password := os.Getenv("GRAFANA-ADMIN-PASSWORD")

	client, err := sdk.New(grafanaURL, sdk.Config{BasicAuth: url.UserPassword("admin", password), NumRetries: 2})
	if err != nil {
		logs.Error("Error connecting to Grafana API server: " + err.Error())
		return nil, err
	}
	return client, nil
}

// Check if the user is already present in Grafana
// If present, get the user ID
// If not present, add the user and get the user ID
func getGrafanaUserId(client *sdk.Client, loginId string, logs *apex.Entry) (int64, error) {

	var userId int64
	user, err := client.UserByEmail(loginId)
	if err != nil {
		logs.Error("User not found: " + err.Error())

		// Generating random password - https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
		password := make([]byte, 10)
		randomInt := rand.Int63()
		for i := 0; i < 10; {
			idx := int(randomInt & 63)
			password[i] = passwordBytes[idx]
			randomInt >>= 6
			i++
		}

		var userDetails = sdk.User{Email: loginId, Password: string(password)}
		user, err := client.CreateUser(userDetails)

		if err != nil {
			logs.Error("Couldn't create user: " + err.Error())
			return -1, err

		} else {
			logs.Info("Created new User")
			userId = user
		}

	} else {
		userId = user.ID
	}
	logs.Info("User ID: " + fmt.Sprintf("%d", userId))
	return userId, nil
}

// Update the user roles in Grafana
func updateUserPermission(loginId string, orgRoles map[string]string, logs *apex.Entry) error {

	client, err := getGrafanaClient(logs)
	if err != nil {
		return err
	}

	userId, err := getGrafanaUserId(client, loginId, logs)
	if err != nil {
		return err
	}

	// Update the org-roles for the user
	for org, role := range orgRoles {
		currOrg, err := client.OrgByName(org)
		if err != nil {
			logs.Info("Invalid organization - " + org)

		} else {
			if role == "delete-user" {
				err = client.RemoveOrgUser(currOrg.ID, userId)

			} else {
				err = client.UpdateOrgUser(currOrg.ID, userId, role)

				// If UpdateOrgUser fails, the user is new the org
				// Add the user to the org with their respective roles
				if err != nil {

					err = client.AddOrgUser(currOrg.ID, loginId, role)
					if err != nil {
						logs.Info("Couldn't add user to org - " + org)
					}
				}
			}
		}
	}

	return nil
}
