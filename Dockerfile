FROM golang:1-alpine AS builder

COPY go.mod .
#COPY go.sum .
RUN go mod download

COPY * ./

RUN go build -o /checkssl


## Deploy
FROM alpine:latest

WORKDIR /

COPY --from=build /checkssl /checkssl

USER nonroot:nonroot

ENTRYPOINT ["/checkssl"]