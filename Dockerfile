# ref: https://docs.docker.com/language/golang/build-images/
# syntax=docker/dockerfile:1

## Build
FROM golang:1.19 AS build-base

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go test ./... -v
RUN CGO_ENABLED=0 go build -o ./out/expenses-app .
EXPOSE 2565

## Deploy
FROM alpine:3.16.2

COPY --from=build-base /app/out/expenses-app /app/expenses-app
CMD ["/app/expenses-app"]