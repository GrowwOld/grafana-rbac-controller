package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	apex "github.com/apex/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Will update the user to appropriate Grafana org-roles
func updateUsers(loginId string, logs *apex.Entry) error {

	if loginId == "" {
		logs.Error("The X-WEBAUTH-EMAIL header is missing")
		return nil
	}

	// Get the user's org-roles mapping
	orgRoles, err := getUserPermission(loginId, logs)
	if err != nil {
		return err
	}

	// Update user's Grafana roles
	err = updateUserPermission(loginId, orgRoles, logs)
	if err == nil {
		logs.Info("Updated user permissions")
	}

	return err
}

func updateUsers_timeout(w http.ResponseWriter, r *http.Request) {

	logs := apex.WithField("user", "server")
	// Read the timeout duration of API from env-variable
	timeoutDurationString := os.Getenv("GRAFANA-RBAC-CONTROLLER-API-TIMEOUT")
	timeoutDuration, err := strconv.Atoi(timeoutDurationString)

	// Cancel context after 5s if the converting env-variable to int returns an error
	if err != nil {
		logs.Error("Couldn't convert timeout duration from string to int. Using the default timeout - 5s")
		timeoutDuration = 5
	}

	logs.Info("Timeout duration: " + fmt.Sprintf("%d", timeoutDuration))
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutDuration)*time.Second)
	defer cancel()

	loginId := r.Header.Get("X-WEBAUTH-EMAIL")
	// Set the user to emailId for debugging in all log statement
	logs = apex.WithField("user", loginId)

	updateUserChannel := make(chan error, 1)
	go func() {
		updateUserChannel <- updateUsers(loginId, logs)
	}()

	// Wait for updateUsers to finish execution.
	// Indicate timeout error (thirdPartyAPIErrors) if the function doesn't return within 5s
	select {
	case err := <-updateUserChannel:
		if err != nil {
			logs = apex.WithField("user", "updateUser-error")
			logs.Error("Updating user failed. Check logs for more information")
			go func() {
				incrementRoleUpdateErrors()
			}()
		}
	case <-ctx.Done():
		// Once ctx.Done() returns after 5s, user will be redirected to Grafana
		// updateUsers will continue to finish it's processing in the background
		logs = apex.WithField("user", "timeout-error")
		logs.Error("updateUsers() timed out. Check logs for more information")
		go func() {
			incrementTimeoutErrors()
		}()
	}

	// Redirect user to Grafana
	logs.Info("Redirecting to Grafana...")
	logs.Info("----------------x--------------")
	http.Redirect(w, r, "/", 302)

	return
}

var PromRoleUpdateErr = promauto.NewCounter(prometheus.CounterOpts{
	Name: "grafana_role_update_errors",
	Help: "Number of user role update error",
})

var PromTimeoutErr = promauto.NewCounter(prometheus.CounterOpts{
	Name: "timeout_errors",
	Help: "Number of user timeout error",
})

func incrementRoleUpdateErrors() {
	PromRoleUpdateErr.Inc()
}

func incrementTimeoutErrors() {
	PromTimeoutErr.Inc()
}

func main() {
	fmt.Println("Starting server at port 9080 ...")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/controller", updateUsers_timeout)
	log.Fatal(http.ListenAndServe(":9080", nil))
}
