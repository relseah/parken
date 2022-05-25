rmdir dist /s /q
cd backend
go build %* -o ..\dist\parken.exe .\cli
cd ..
robocopy . dist config.json payload.json
robocopy frontend dist\frontend /s
if errorlevel 8 (exit /b 1) else (exit /b 0)
