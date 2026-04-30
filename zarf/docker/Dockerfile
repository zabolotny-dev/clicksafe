FROM golang:1.25.9-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /clicksafe_api api/service/main.go
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /clicksafe_tool api/tooling/main.go

FROM alpine:3.21

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /clicksafe_api /clicksafe_api
COPY --from=builder /clicksafe_tool /clicksafe_tool

RUN chmod +x /clicksafe_api /clicksafe_tool

RUN mkdir -p /var/www/my-app/uploads && chown appuser:appgroup /var/www/my-app/uploads

USER appuser

EXPOSE 8080

CMD ["/clicksafe_api"]