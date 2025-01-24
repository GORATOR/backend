FROM golang:1.23-bookworm AS build

WORKDIR /app

COPY . .

RUN go mod download \
    && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build .


FROM gcr.io/distroless/static-debian12:latest AS production

COPY --from=BUILD /app/backend .

CMD ["./backend"]
