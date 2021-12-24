FROM scratch
ADD mhub mhub
ADD ca.pem ca.pem
ADD server.pem server.pem
ADD server.key server.key
EXPOSE 8083
ENTRYPOINT ["/mhub"]