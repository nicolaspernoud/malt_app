#!/bin/bash
set -a
. ./.env.keycloak
set +a
./keycloak/keycloak-up.sh
go run main.go
