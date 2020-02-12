FROM golang:1.13 AS builder
WORKDIR /deploy
ENV CGO_ENABLED=0
COPY . .
RUN go build -ldflags="-s -w"

FROM gcr.io/distroless/static
COPY --from=builder /deploy/github-deployments .
ENTRYPOINT ["/github-deployments"]
