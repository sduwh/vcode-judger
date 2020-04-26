FROM alpine:3.9.6

COPY ./target /app

COPY ./config/config.yaml /app/config/config.yaml

WORKDIR /app

ENTRYPOINT ["./vcode-judger"]