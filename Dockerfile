FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o main .

FROM scratch
COPY --from=builder /app/main /
COPY --from=builder /app/templates /templates
COPY --from=builder /app/static /static

EXPOSE 8080
ENV GIN_MODE=release
ENV PORT=8080

CMD ["/main"]