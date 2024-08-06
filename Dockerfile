FROM golang:1-alpine AS builder

ENV GOPATH=/go
RUN mkdir -p "/go/src/github.com/szazeski/checkssl"
WORKDIR "/go/src/github.com/szazeski/checkssl"

#COPY go.mod .
#COPY go.sum .
#RUN go mod download

COPY . .

RUN go build -o /checkssl

## Deploy
FROM alpine:latest
MAINTAINER steve@checkssl.org
LABEL build_date="2024-04-21"
LABEL built_version="0.5.1"

WORKDIR /
COPY --from=builder /checkssl /checkssl

ENTRYPOINT ["/checkssl"]


# docker buildx build --platform linux/arm/v7,linux/arm64,linux/amd64 --tag szazeski/checkssl:0.5.1 --tag szazeski/checkssl:latest --push .
# you may need to do a `docker buildx create --use`
