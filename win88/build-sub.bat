@echo off
echo Build %1 task!
cd %1
go fmt
go vet
go build -v
echo errorlevel:%errorlevel%
if "%errorlevel%"=="0" exit
if not "%errorlevel%"=="0" echo %1 build failed!
pause
exit