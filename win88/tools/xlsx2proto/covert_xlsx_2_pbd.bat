@echo off
set path=%path%;E:\GO\go1.15.15\bin
set work_path="E:\gocode\trunk\src\games.yol.com\win88"
set protoc=%work_path%\bin\protoc-3.19.4-win64\bin\protoc.exe
set protoc-gen-go-plugin-path="%work_path%\bin\protoc-gen-go.exe"
if not exist xlsx2proto.exe (
    go build
)
xlsx2proto.exe
cd ../../protocol/server
%protoc% --plugin=protoc-gen-go=%protoc-gen-go-plugin-path% --go_out=. pbdata.proto
cd ../../tools/xlsx2binary
go fmt
go build
xlsx2binary.exe
rem cd ../../
rem copydata.bat
pause
