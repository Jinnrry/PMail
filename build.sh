#!/bin/bash

# 编译前端网站
echo "start building fe"
cd fe 
npm install && npm run build
cd ..

# 删除 http_server 下的 dist 文件夹
echo "delete dist folder"
rm -rf server/listen/http_server/dist

# 复制 fe/dist 到 http_server
echo "copy dist folder"
cp -r fe/dist server/listen/http_server/

# 编译 go 语言项目
echo "start building server"
cd server
go build -o pmail main.go

cp pmail ~/test
cd ..
echo "build success"
