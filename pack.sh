#!/bin/bash

echo "start pack vcode-judger..."
# build target for linux
echo "setting golang env..."
go env -w GO111MODULE=on && go env -w GOPROXY="https://goproxy.cn,direct"

echo "start download golang module..."
go mod download && go mod tidy

echo "start build golang program..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./target/vcode-judger

# get git tag
# shellcheck disable=SC2046
tag=$(git describe --tags $(git rev-list --tags --max-count=1))
echo "get the newest tag: ${tag}"

# build docker
echo "start build docker..."
docker build --tag vcode-judger:"${tag}" -f ./Dockerfile .

echo "pack vcode-judger docker image success..."
