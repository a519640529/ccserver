set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
cd gatesrv
go fmt
go build
cd ../mgrsrv
go fmt
go build
cd ../worldsrv
go fmt
go build
cd ../gamesrv
go fmt
go build
rem cd ../minigame
rem go fmt
rem go build
cd ../robot
go fmt
go build
cd ../dbproxy
go fmt
go build
cd ..
@echo "complete"
pause
