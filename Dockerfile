FROM golang:1-alpine AS builder

ENV GOPATH=/go
RUN mkdir -p "/go/src/github.com/szazeski/checkssl"
WORKDIR "/go/src/github.com/szazeski/checkssl"

#COPY go.mod .
#COPY go.sum .
#RUN go mod download

COPY * ./

RUN go build -o /checkssl

## Deploy
FROM alpine:latest
MAINTAINER steve@checkssl.org
LABEL build_date="2022-10-03"
LABEL built_version="0.4.4"

WORKDIR /
COPY --from=builder /checkssl /checkssl

ENTRYPOINT ["/checkssl"]