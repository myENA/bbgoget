FROM golang:1-alpine AS bld
MAINTAINER Nathan Johnson <njohnson@ena.com>
COPY . /go/src/github.com/myENA/bbgoget
WORKDIR /go/src/github.com/myENA/bbgoget
RUN go install

FROM alpine:latest
COPY --from=bld /go/bin/bbgoget /usr/local/bin/bbgoget
EXPOSE 8800/tcp
ENTRYPOINT ["/usr/local/bin/bbgoget"]
