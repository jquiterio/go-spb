#!/bin/bash

openssl genrsa -out ca.key 2048
openssl req -new -x509 -key ca.key -out ca.pem -days 3650 -subj '/CN=root ca'
openssl genrsa -out server.key 2048
openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/CN=localhost" -out server.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:localhost") -days 365 -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out server.pem
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj '/CN=localhost' 
openssl x509 -req -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out client.pem -days 3650