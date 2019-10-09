# The Malt App

This application is based upon [QOR](https://github.com/qor/qor).

## Install

Clone the repository and `cd malt-app`.

## Use with docker

### Build the container

`./build.sh`

If it doesn't build, remove or rename the './data/business.db' file.

### Start the services

`./up.sh`

### Usage

Open [the admin interface](http://localhost:8081/admin?locale=fr-FR), log with username : demo and password : demo and try the UI.

### User configuration

Open [the keycloak interface](http://localhost:8080), log with username : admin and password : admin and manage your users (in the "malt" realm).

## Use directly (for coding)

### Setup

Golang must me installed and $GOPATH environnment variable set.
Exec `./setup.sh`.

### Start keycloak

`./keycloak-up.sh`

### Start the app

Start the application with `./start-without-docker.sh`.

### Debug

Use the "Debug Server with Keycloak" VS Code debug configuration.

### Code

Alter the 'models.go' file to add your own models and restart the app. They are automatically added to the admin interface.
Access them through the api by using the model route + `.json` (example : http://localhost:8081/admin/employees.json).
Use the common REST methods on this endpoint to create, alter and delete entities.