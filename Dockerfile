FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

RUN go mod download

COPY . .

RUN go build -o register1 .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/register1 .

EXPOSE 8080

CMD ["./register1"]
