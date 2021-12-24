#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./mhub
./gencert.sh localhost
docker build -f Dockerfile -t mhub:latest .