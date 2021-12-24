openssl genrsa -out ca.key 2048
openssl req -new -x509 -key ca.key -out ca.pem -days 3650 -subj '/CN=ca'
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj '/CN=server'
openssl x509 -req -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out server.pem -days 3650
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj '/CN=client'
openssl x509 -req -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out client.pem -days 3650