FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o example ./example

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder --chmod=777 /app/example /usr/local/bin/example

WORKDIR /data/config
ENTRYPOINT ["/usr/local/bin/example"]
CMD ["serve"]