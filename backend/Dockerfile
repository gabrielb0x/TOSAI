# syntax=docker/dockerfile:1

FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/autotos ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /bin/autotos /bin/autotos
EXPOSE 8080
ENTRYPOINT ["/bin/autotos"]
