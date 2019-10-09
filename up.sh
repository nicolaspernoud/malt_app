#!/bin/bash
# Working directory
WD="$(
    cd "$(dirname "$0")"
    pwd -P
)"
# Down services
$WD/down.sh
# Source env
set -a
. $WD/.env.keycloak
set +a

# START KEYCLOAK
./keycloak/keycloak-up.sh

# START QOR
docker run -d --name malt-qor \
    --restart unless-stopped \
    -v /etc/localtime:/etc/localtime:ro \
    -v /etc/timezone:/etc/timezone:ro \
    -v $WD/data:/app/data:rw \
    --user "1000" \
    --net=host \
    -e REDIRECT_URL=${REDIRECT_URL} \
    -e CLIENT_ID=${CLIENT_ID} \
    -e CLIENT_SECRET=${CLIENT_SECRET} \
    -e AUTH_URL=${AUTH_URL} \
    -e TOKEN_URL=${TOKEN_URL} \
    -e USERINFO_URL=${USERINFO_URL} \
    -e LOGOUT_URL=${LOGOUT_URL} \
    -e ADMIN_GROUP=${ADMIN_GROUP} \
    malt-qor

# Let it start...
sleep 20
echo  "*** Ready to go ! ***"
docker logs malt-qor
