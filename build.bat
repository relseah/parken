rmdir dist /s /q
cd backend
go build %* -o ..\dist\parken.exe .\cmd
cd ..
robocopy . dist config.json dummy.json
robocopy frontend dist\frontend /s
:: /ndl /nfl
if errorlevel 8 (exit /b 1) else (exit /b 0)
