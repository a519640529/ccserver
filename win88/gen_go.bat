
@echo off
set work_path=%cd%
set proto_path=%work_path%\protocol
set protoc=%work_path%\bin\protoc-3.19.4-win64\bin\protoc.exe
set protoc-gen-go-plugin-path="%work_path%\bin\protoc-gen-go.exe"

echo %protoc3%
cd %proto_path%
for /d %%s in (,*) do (
	if %%s NEQ webapi (
		cd %%s
 		for %%b in (,*.proto) do (
 	    	echo %%b
	    	%protoc% --plugin=protoc-gen-go=%protoc-gen-go-plugin-path% --go_out=. %%b
		)
		cd ..
	)
)
pause