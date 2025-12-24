# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/tosai ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /bin/tosai /bin/tosai
EXPOSE 9000
ENTRYPOINT ["/bin/tosai"]
