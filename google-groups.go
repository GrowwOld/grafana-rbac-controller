package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	apex "github.com/apex/log"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

func getGoogleAPIClient(userEmail string, ServiceAccountFilePath string) (*admin.Service, error) {

	ctx := context.Background()

	jsonCredentials, err := ioutil.ReadFile(ServiceAccountFilePath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, admin.AdminDirectoryGroupScope)
	if err != nil {
		return nil, fmt.Errorf("JWTConfigFromJSON: %v", err)
	}
	config.Subject = userEmail
	ts := config.TokenSource(ctx)

	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("NewService: %v", err)
	}
	return srv, nil
}

func getGoogleGroups(loginId string, logs *apex.Entry) ([]string, error) {
	// Create a client that will interact with Google Admin API
	adminID := os.Getenv("GOOGLE-WORKSPACE-ADMIN")

	srv, err := getGoogleAPIClient(adminID, "./credentials/credentials.json")
	if err != nil {
		logs.Error("Error in getting Google API client")
		return nil, err
	}

	logs.Info("Created the Google API client")

	// List all groups in "groww.in" that the user is a part of
	groupsReport, err := srv.Groups.List().Domain("groww.in").UserKey(loginId).Do()
	if err != nil {
		logs.Error("Error in getting response from the Google Admin API")
		return nil, err
	}

	groups := []string{}
	logs.Info("Got response from the Google Admin API")
	if len(groupsReport.Groups) != 0 {
		for _, u := range groupsReport.Groups {
			groups = append(groups, u.Email)
		}
	}

	logs.Info("Google groups: " + strings.Join(groups, ", "))
	return groups, nil
}
