FROM golang:1.22

WORKDIR /app

COPY . .

RUN cat /app/deploy/mirror > /etc/apt/sources.list.d/debian.sources

RUN git config --global --add safe.directory /app

RUN export GOPROXY=https://mirrors.aliyun.com/goproxy && go install ./...

# RUN timedatectl set-timezone Asia/Shanghai
