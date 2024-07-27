build: build_fe build_server telegram_push web_push wechat_push package

clean:
	rm -rf output


build_fe:
	cd fe && yarn && yarn build
	rm -rf server/http_server/dist
	cd server && cp -rf ../fe/dist http_server

build_server:
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail_linux_amd64  main.go
	cd server && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail_windows_amd64.exe  main.go
	cd server && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail_mac_amd64  main.go
	cd server && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail_mac_arm64  main.go

telegram_push:
	cd server/hooks/telegram_push && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o output/telegram_push_linux_amd64  telegram_push.go
	cd server/hooks/telegram_push && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o output/telegram_push_windows_amd64.exe  telegram_push.go
	cd server/hooks/telegram_push && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o output/telegram_push_mac_amd64  telegram_push.go
	cd server/hooks/telegram_push && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o output/telegram_push_mac_arm64  telegram_push.go


wechat_push:
	cd server/hooks/wechat_push && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o output/wechat_push_linux_amd64  wechat_push.go
	cd server/hooks/wechat_push && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o output/wechat_push_windows_amd64.exe  wechat_push.go
	cd server/hooks/wechat_push && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o output/wechat_push_mac_amd64  wechat_push.go
	cd server/hooks/wechat_push && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o output/wechat_push_mac_arm64  wechat_push.go

spam_block:
	cd server/hooks/spam_block && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o output/spam_block_linux_amd64  spam_block.go
	cd server/hooks/spam_block && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o output/spam_block_windows_amd64.exe  spam_block.go
	cd server/hooks/spam_block && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o output/spam_block_mac_amd64  spam_block.go
	cd server/hooks/spam_block && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o output/spam_block_mac_arm64  spam_block.go



plugin: telegram_push wechat_push


package: clean
	mkdir output
	mv server/pmail* output/
	mkdir output/config
	mkdir output/plugins
	cp -r server/config/dkim output/config/
	cp -r server/config/ssl output/config/
	cp -r server/config/config.json output/config/
	mv server/hooks/telegram_push/output/* output/plugins
	mv server/hooks/wechat_push/output/* output/plugins
	cp README.md output/

test:
	export setup_port=17888 && cd server && go test -v ./...

test_mysql:
	export setup_port=17888 && cd server && go test -args "mysql" -v ./...

test_postgres:
	export setup_port=17888 && cd server && go test -args "postgres" -v ./...