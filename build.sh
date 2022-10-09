#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./mhub
docker build -f Dockerfile -t mhub:latest .