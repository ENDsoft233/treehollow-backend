#!/bin/bash
# sudo docker run -d -e MYSQL_ROOT_PASSWORD=treehollow114514 --name mysql8 -d -p 36602:3306 -v mysql_data8_35:/var/lib/mysql -v /home/ubuntu/treehollow-backend/data/mysql/log:/var/log/ -v /home/ubuntu/treehollow-backend/data/mysql/my.cnf:/etc/mysql/my.cnf --network treehollow-backend_default mysql:8.0.35
echo start build backend
export GOPROXY=https://mirrors.aliyun.com/goproxy && go install ./...
echo build successfully, going to run after 3s
sleep 3
/go/bin/treehollow-v3-push-api &
/go/bin/treehollow-v3-security-api &
/go/bin/treehollow-v3-services-api &
/go/bin/treehollow-v3-fallback