set GODEBUG=gctrace=1
cd gatesrv_tcp
::start gatesrv_tcp.exe    不需要启动
cd ../mgrsrv
start mgrsrv.exe
cd ../worldsrv
start worldsrv.exe
cd ../gamesrv
start gamesrv.exe
cd ../gatesrv_ws
start gatesrv_ws.exe
cd ../dbproxy
start dbproxy.exe

