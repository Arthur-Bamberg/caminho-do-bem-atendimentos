FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates wget
COPY --from=build /server /usr/local/bin/server
EXPOSE 8080
HEALTHCHECK --interval=2s --timeout=3s --retries=20 \
  CMD wget -qO- http://127.0.0.1:8080/api/saude || exit 1
ENTRYPOINT ["/usr/local/bin/server"]
