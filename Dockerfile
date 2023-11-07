FROM golang:latest as build

ENV CGO_ENABLED=1

ENV GOOS=linux
ENV GOARCH=amd64

RUN apt-get update && apt-get install -y clang

COPY . /app

WORKDIR /app

CMD go build -o /app/hot_cache_linux -a
