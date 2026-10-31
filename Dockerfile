FROM alpine:latest
COPY checkssl /checkssl
USER nobody
ENTRYPOINT ["/checkssl"]
