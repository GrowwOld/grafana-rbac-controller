# Grafana RBAC Controller  
[![GitHub license](https://img.shields.io/github/license/Groww/grafana-rbac-controller?color=51C838)](https://github.com/Groww/grafana-rbac-controller/blob/main/LICENSE)   [![GitHub issues](https://img.shields.io/github/issues/Groww/grafana-rbac-controller?color=51C838)](https://github.com/Groww/grafana-rbac-controller/issues)  ![Release](https://img.shields.io/github/v/release/Groww/grafana-rbac-controller)

![Grafana](https://img.shields.io/badge/grafana-%23F46800.svg?style=for-the-badge&logo=grafana&logoColor=white)   ![Google](https://img.shields.io/badge/google-4285F4?style=for-the-badge&logo=google&logoColor=white)   ![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)   ![Nginx](https://img.shields.io/badge/nginx-%23009639.svg?style=for-the-badge&logo=nginx&logoColor=white)

Grafana is an observability stack that allows you to monitor and analyse metrics, logs, and traces. 

![](https://github.com/grafana/grafana/blob/main/docs/logo-horizontal.png)

Grafana users and resources such as dashboards, data sources and plugins can be separated from each other using Grafana 
organisations which provides a fully isolated Grafana experience (with different set of users and their respective roles, 
dashboards and other resources in each organisation). Hence Grafana organisations support multi-tenancy in a single instance of Grafana

## Why Grafana RBAC Controller?
When a new user has to be onboarded to Grafana, organisation admins have to manually add the user to their respective organisations 
and grant appropriate user roles. Similarly organisation admins have to manually modify user permissions when required.
Grafana LDAP authentication lets you map LDAP groups to Grafana organisations and roles. 
However, our requirement is to have Grafana org-role mapping based on google group membership rather than LDAP group membership.

> **Requirement: To update users with appropriate roles in various Grafana organisations based on the user’s google group membership**

Grafana RBAC Controller is a proxy layer around the grafana service that provides users access to Grafana based on their google group membership. The mapping beteween grafana org-role and google group is maintained as a configuration file. We propose the following architecture for updating user permissions to various Grafana organisations on login. </br>

The working of Grafana RBAC Controller is explained in much detail [here](https://tech.groww.in/google-groups-to-manage-grafana-roles-grafana-rbac-controller-8b7efa0f081a)

</br>
<img src="https://github.com/Groww/grafana-rbac-controller/blob/main/assets/architecture.png" width="75%">

## To build Grafana RBAC Controller

### Requirements:
1. Grafana setup in the kubernetes cluster - [Link](https://grafana.com/docs/grafana/next/setup-grafana/installation/kubernetes/) 
2. `auth.proxy` enabled on Grafana and other configurations to [run Grafana behind a reverse proxy](https://grafana.com/tutorials/run-grafana-behind-a-proxy/)
> **Note:** You can also refer to the grafana manifest [here](https://github.com/Groww/grafana-rbac-controller/blob/main/examples/deploy-grafana.yaml) for deploying and configuring Grafana
3. Registered domain name (Eg: groww.in)
4. OAuth client ID and secret - [Link](https://developers.google.com/adwords/api/docs/guides/authentication#create_a_client_id_and_client_secret)
  - Authorized JavaScript origins: `https://grafana.domain.in`
  - Authorized redirect URIs: `https://grafana.domain.in/oauth2/callback`
5. SSL certificates for the registered domain
6. Google Admin SDK service account with `admin.directory.group.readonly` permission - [Link](https://developers.google.com/admin-sdk/directory/v1/guides/delegation)


The entire setup can be built in kubernetes using [helm-charts](https://github.com/Groww/grafana-rbac-controller/tree/main/helm-charts). 
Ensure to make the following changes beforing deploying the setup:
- modify values enclosed within angular bracket - <> (such as hostname, OAuth credentials, etc...) in [values.yaml](https://github.com/Groww/grafana-rbac-controller/blob/main/helm-charts/values-test.yaml)
- update the [org-permissions.json](https://github.com/Groww/grafana-rbac-controller/blob/main/helm-charts/values-test.yaml#L272)
based on your requirements of updating user roles to the existing Grafana organisations

You can deploy the setup in kubernetes using the following command:
```
cd helm-charts
helm upgrade --install grafana-rbac-controller .
```

## To use Grafana RBAC Controller
Grafana can be accessed normally via its proxy layer from its registered URL (say grafana.domain.in)

Every time a user logs in to Grafana via the proxy by logging in using their Google credentials (OAuth authentication):
- the user's list of google groups is fetched
- user's role in various Grafana organisation is decided based on [org-permissions.json](https://github.com/Groww/grafana-rbac-controller/blob/main/helm-charts/values-test.yaml#L272)
- Grafana HTTP API calls are made to update the user roles
- user gets directed to Grafana without impacting the user experience

You can set timeout to the controller API [here](https://github.com/Groww/grafana-rbac-controller/blob/main/helm-charts/values-test.yaml#L233).
In case the Google Admin API or Grafana HTPP API calls take a long time to respond, the user will be directed to Grafana and the user role update process
will occur in the background without disturbing the user experience. Note that default timeout duration is set at 5s.

### Metrics exposed:
Prometheus metrics regarding the number of grafana-role-update errors are exposed by controller container at port 9080:
- `grafana_role_update_errors`: Prometheus counter metrics which increments on events when updating of user roles to Grafana failed due to unpredictable reasons. Check the pod logs for debugging.
- `timeout_errors`: Prometheus counter metrics which increments on events when controller API timed out

Golang application [health metrics](https://prometheus.io/docs/guides/go-application/) are also exposed along with the above metrics
