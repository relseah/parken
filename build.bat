del /q dist\*
for /d %%d in (dist\*) do if not %%~nd == tiles rmdir /s /q "%%~d"
cd backend
go build %* -o ..\dist\parken.exe .\cmd
cd ..
robocopy . dist config.json dummy.json
robocopy frontend dist\frontend /s /xf *.js
uglifyjs frontend\leaflet\leaflet.js frontend\*.js -c -m --toplevel -o dist\frontend\parken.js
:: debug
:: robocopy frontend dist\frontend /s 
if errorlevel 8 (exit /b 1) else (exit /b 0)
