FROM golang:1.26-alpine AS build

WORKDIR /src
RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/agenttrace ./cmd/agenttrace

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=build /out/agenttrace /usr/local/bin/agenttrace

EXPOSE 6006 4317
VOLUME ["/app/data"]

ENTRYPOINT ["agenttrace"]
CMD ["serve", "--http-addr=:6006", "--grpc-addr=:4317"]
