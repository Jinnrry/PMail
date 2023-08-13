# 编译前端代码
cd fe && yarn && yarn build

# 编译后端代码
cd ../server && cp -rf ../fe/dist http_server


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail_linux_amd64  main.go

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail_windows_amd64  main.go

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail_mac_amd64  main.go

CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail_mac_arm64  main.go

# 整理输出文件
cd ..
rm -rf output
mkdir output
cd output
mv ../server/pmail* .
mkdir config
cp -r ../server/config/dkim config/
cp -r ../server/config/ssl config/
cp -r ../server/config/config.json config/
cp ../README.md .