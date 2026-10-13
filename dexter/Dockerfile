FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /dexterd ./cmd/dexterd

FROM alpine:3.21
RUN apk add --no-cache git
COPY --from=build /dexterd /usr/local/bin/dexterd
ENTRYPOINT ["dexterd"]
