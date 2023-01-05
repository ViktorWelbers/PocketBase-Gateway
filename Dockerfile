FROM golang:1.19 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /api-gateway

EXPOSE 8080 8080
ENTRYPOINT ["/api-gateway", "serve", "--http=0.0.0.0:8080"]

