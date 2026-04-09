FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/thewar ./main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /out/thewar /app/thewar

EXPOSE 8080
CMD ["/app/thewar"]