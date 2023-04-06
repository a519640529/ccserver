cd ../../

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
cd ./mgrsrv
go build
cd ../gatesrv
go build
cd ../worldsrv
go build
cd ../gamesrv
go build
cd ../minigame
go build
cd ../robot
go build
cd ../dbproxy
go build

cd ..
cd ./tools/upload

WinSCP.exe /console /script=upload.txt /log=upload.log