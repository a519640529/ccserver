cd protocol_lua
for %%i in (*.proto) do ( 
	"protoc.exe" --plugin=protoc-gen-lua="R:/gocode/trunk/bin/protoc-gen-lua.bat" --lua_out=. %%i
)	
move /Y *.lua R:\quick-jxjyhj/trunk/jxjyqp/src/protocol

if exist "R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\pbdata_pb.lua" (del R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\pbdata_pb.lua)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\server_pb.lua" (del R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\server_pb.lua)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\transmit_pb.lua" (del R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\transmit_pb.lua)

echo end
cd ..
pause