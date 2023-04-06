set path=%path%;E:\GO\go1.15.15\bin
if not exist xlsx2proto.exe (
    go build
)
xlsx2proto.exe
cd ../../protocol/pbdata
protoc --go_out=. pbdata.proto
cd ../../tools/xlsx2binary
go fmt
go build
xlsx2binary.exe
cd ../../

pause
