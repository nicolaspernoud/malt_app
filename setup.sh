#!/bin/bash
echo "GOPATH is ${GOPATH}"

# Install dependencies and update
go get -u
go mod tidy

# Get last version of templates
WD=$(pwd)
cd $GOPATH
go get github.com/qor/admin
go get github.com/qor/i18n
cd $WD

# Copy templates
rm -rf ./vendor/github.com/qor/
mkdir -p ./vendor/github.com/qor/admin/
cp -r $GOPATH/src/github.com/qor/admin/views/ ./vendor/github.com/qor/admin/
cp -r $GOPATH/src/github.com/qor/i18n/views/ ./vendor/github.com/qor/i18n/
