set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go fmt
go build
@echo "complete"
pause
