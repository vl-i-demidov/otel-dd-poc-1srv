# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.18-buster AS build

#
WORKDIR /app

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY ./*.go ./

RUN go build -o /main-app

##
## Deploy
##
FROM ubuntu:22.04

WORKDIR /

COPY --from=build /main-app /app

EXPOSE 8001

ENTRYPOINT ["/app"]