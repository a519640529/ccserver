@echo "go fmt common..."
cd common
go fmt
@echo "go fmt api3th..."
cd ../api3th
go fmt
@echo "go fmt model..."
cd ../model
go fmt
@echo "go fmt webapi..."
cd ../webapi
go fmt
@echo.
@echo.
@echo.
@echo "go fmt lib complete!"
cd ..
start build-sub.bat gatesrv
start build-sub.bat mgrsrv
start build-sub.bat worldsrv
start build-sub.bat gamesrv
rem start build-sub.bat minigame
start build-sub.bat robot
start build-sub.bat dbproxy
@echo "Wait all build task complete!"
pause
