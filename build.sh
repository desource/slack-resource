#!/bin/sh
set -e

echo ">>> copy /etc/ssl/certs/ca-certificates.crt"
certs=("/etc/ssl/certs/ca-certificates.crt" "/etc/pki/tls/certs/ca-bundle.crt" "/etc/ssl/ca-bundle.pem")
for cert in "${certs[@]}"; do
    if [ -f ${cert} ]; then
        mkdir -p etc/ssl/certs
        cat ${cert} > etc/ssl/certs/ca-certificates.crt
        break
    fi
done
echo

GOOS=linux
GOARCH=amd64
CGO_ENABLED=0

echo ">>> go build /bin/slack-resource"
go build -o bin/slack-resource main.go slack.go
echo

echo ">>> docker build quay.io/desource/slack-resource"
docker build -t quay.io/desource/slack-resource .
echo

echo ">>> docker push quay.io/desource/slack-resource"
docker push quay.io/desource/slack-resource


