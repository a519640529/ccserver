
@echo off
cd tools/proto2js
go fmt
go build
proto2js.exe
pause