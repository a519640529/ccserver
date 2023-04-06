set GODEBUG=gctrace=1
cd gatesrv
start gatesrv.exe
cd ../dbproxy
start dbproxy.exe
cd ../mgrsrv
start mgrsrv.exe
cd ../worldsrv
start worldsrv.exe
cd ../gamesrv
start gamesrv.exe
rem cd ../minigame
rem start minigame.exe