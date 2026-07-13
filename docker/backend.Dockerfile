# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS dev
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest
WORKDIR /app
CMD ["air"]

FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM alpine:3.20 AS prod
RUN apk add --no-cache ca-certificates
COPY --from=build /out/api /usr/local/bin/api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
