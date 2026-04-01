FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -o products cmd/main.go

FROM alpine:3.14
WORKDIR /app
COPY --from=builder /app/products .
EXPOSE 42069
CMD ["./products"]
