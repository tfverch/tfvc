FROM alpine:latest

# install git
RUN apk add --no-cache git

COPY tfvc /usr/bin/tfvc

## use a non-privileged user
RUN adduser -D tfvc
USER tfvc

# set the default entrypoint -- when this container is run, use this command
ENTRYPOINT [ "tfvc" ]
# as we specified an entrypoint, this is appended as an argument (i.e., `tfsec --help`)
CMD [ "--help" ]