# Stage 1: Build
FROM golang:1.23-alpine AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vinyl-catalog ./cmd/api/main.go

# Stage 2: Run
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/vinyl-catalog .

RUN mkdir -p /app/uploads && chown nobody:nobody /app/uploads

EXPOSE 8080

USER nobody:nobody

ENTRYPOINT ["./vinyl-catalog"]
