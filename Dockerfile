FROM debian:bookworm-slim AS BUILD

WORKDIR /app

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
    golang \
    && apt-get clean \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt /var/lib/dpkg /tmp/* /var/tmp/*

COPY . .

RUN go mod download \
    && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build .


FROM gcr.io/distroless/static-debian12:latest AS PRODUCTION

COPY --from=BUILD /app/backend .

CMD ["./backend"]
