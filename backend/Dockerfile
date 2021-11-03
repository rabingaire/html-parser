FROM golang:alpine as build

WORKDIR /app

# Copy go source code and install dependency
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

RUN go build -o html-parser

# production environment
FROM alpine:latest

WORKDIR /app

ENV GIN_MODE=release

COPY --from=build /app/html-parser .

EXPOSE 8000

ENTRYPOINT ["./html-parser"]
