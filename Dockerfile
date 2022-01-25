FROM scratch
ADD mhub mhub
EXPOSE 8083
ENTRYPOINT ["/mhub"]