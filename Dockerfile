FROM golang:1-alpine AS bld
MAINTAINER Nathan Johnson <njohnson@ena.com>
WORKDIR /go
RUN apk --no-cache add git && \
go get -u github.com/myENA/bbgoget

FROM alpine:latest
COPY --from=bld /go/bin/bbgoget /usr/local/bin/bbgoget
EXPOSE 8800/tcp
ENTRYPOINT ["/usr/local/bin/bbgoget"]
