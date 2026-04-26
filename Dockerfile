FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.sum .
COPY go.mod .
RUN go mod tidy
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o products .

FROM alpine:3.14
WORKDIR /app
COPY --from=builder /app/products .
EXPOSE 42069
CMD ["./products"]
