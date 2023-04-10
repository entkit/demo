FROM golang:1.19-alpine

WORKDIR /app
ADD . /app

# Deps
RUN apk add --no-cache alpine-sdk
RUN apk add --no-cache nodejs npm

# Build
RUN ls -la
RUN go mod tidy
RUN go generate
RUN go build -o /usr/local/bin/entkit ./my-server/*.go

# Run binary
ENTRYPOINT ["entkit"]