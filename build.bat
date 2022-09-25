del /q dist\*
for /d %%d in (dist\*) do if not %%~nd == tiles rmdir /s /q "%%~d"
cd backend
go build %* -o ..\dist\parken.exe .\cmd
cd ..
robocopy . dist config.json dummy.json
robocopy frontend dist\frontend /s /xf *.js
:: debug
:: robocopy frontend dist\frontend /s 
call uglifyjs frontend\leaflet\leaflet.js frontend\*.js -c -m --toplevel -o dist\frontend\parken.js
for /r dist\frontend %%f in (*.html *.css *.js *.ttf) do 7z a %%f.gz %%f
if errorlevel 8 (exit /b 1) else (exit /b 0)
