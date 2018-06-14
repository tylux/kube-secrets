# build stage
FROM golang:1.10.3 AS build-container
ADD . $GOPATH/src/app
WORKDIR $GOPATH/src/app 
RUN go build -o kube-secrets . && pwd

# final stage
FROM debian:stretch-slim

RUN apt-get update && apt-get install -y ca-certificates
WORKDIR /app
COPY --from=build-container /go/src/app/kube-secrets /app/
ENV PORT 8080
EXPOSE 8080

ENTRYPOINT ./kube-secrets