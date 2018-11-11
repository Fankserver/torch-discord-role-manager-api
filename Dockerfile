FROM golang:1.11.1-alpine as builder
WORKDIR /go/src/github.com/fankserver/torch-discord-role-manager-api
COPY . .
RUN apk add --no-cache alpine-sdk \
    && go get ./... \
    && go build -o app .

FROM alpine:latest
RUN adduser -D -u 1000 rolemanger
USER rolemanger

# Add app
COPY --from=builder /go/src/github.com/fankserver/torch-discord-role-manager-api/app /app

# This container will be executable
ENTRYPOINT ["/app"]