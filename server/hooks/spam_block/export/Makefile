build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o export_linux_amd64  export.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build  -o export_windows_amd64.exe  export.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build  -o export_mac_amd64  export.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build  -o export_mac_arm64  export.go
