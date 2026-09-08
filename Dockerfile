FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/orders ./cmd/orders

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/orders /usr/local/bin/orders
USER app
ENTRYPOINT ["orders"]
CMD ["web"]
