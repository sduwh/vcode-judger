#!/usr/bin/env bash

echo "start pack vcode-judger..."
# build target for linux
echo "setting golang env..."
go env -w GO111MODULE=on && go env -w GOPROXY="https://goproxy.cn,direct"

echo "start download golang module..."
go mod download && go mod tidy

echo "start build golang program..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./target/vcode-judger

echo "Start build docker images..."
echo "Login docker registry"
docker login --username=$1 registry.cn-hangzhou.aliyuncs.com

gitTag=$(git describe --tags $(git rev-list --tags --max-count=1))
echo "get the newest tag: ${gitTag}"
dockerTag=$(echo ${gitTag} | egrep '(v[0-9]*\.[0-9]*\.[0-9]*)' -o)

echo "build images..."
docker build --tag registry.cn-hangzhou.aliyuncs.com/vcodeteam/vcode-judger:${dockerTag} -f ./Dockerfile .
docker tag $(docker image ls -q registry.cn-hangzhou.aliyuncs.com/vcodeteam/vcode-judger:${dockerTag}) registry.cn-hangzhou.aliyuncs.com/vcodeteam/vcode-judger:latest

echo "push images..."
docker push registry.cn-hangzhou.aliyuncs.com/vcodeteam/vcode-judger:${dockerTag}
docker push registry.cn-hangzhou.aliyuncs.com/vcodeteam/vcode-judger:latest

echo "pack vcode-judger docker image success..."
