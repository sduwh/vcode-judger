FROM golang:1.14.2-alpine

COPY . /app

WORKDIR /app

RUN go env -w GO111MODULE=on && go env -w GOPROXY="https://goproxy.cn,direct" && go mod download && go build -o ./target/vcode-judger

ENTRYPOINT ["./target/vcode-judger"]