# Building the server

FROM golang:alpine as server-builder

WORKDIR /server

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh build-base

ADD . .
RUN echo "GOPATH is ${GOPATH}" && \
    go get -u && \
    go mod tidy && \
    WD=$(pwd) && \
    cd $GOPATH && \
    go get github.com/qor/admin && \
    go get github.com/qor/i18n && \
    cd $WD && \
    rm -rf ./vendor/github.com/qor/ && \
    mkdir -p ./vendor/github.com/qor/admin/ && \
    cp -r $GOPATH/src/github.com/qor/admin/views/ ./vendor/github.com/qor/admin/ && \
    cp -r $GOPATH/src/github.com/qor/i18n/ ./vendor/github.com/qor/

RUN go test ./... && go build -o server

# Putting all together

FROM alpine:latest

RUN apk update && apk add ca-certificates sqlite
# ca-certificates for autocert (Let's Encrypt)

# Removing apk cache
RUN rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=server-builder /server/server /app
COPY --from=server-builder /server/vendor /app/vendor/
ADD app /app/app

ENTRYPOINT ["./server"]
