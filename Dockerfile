FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o example ./example

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /data/config
COPY --from=builder /app/example /example
ENTRYPOINT ["/example"]
CMD ["serve"]