rmdir dist /s /q
cd backend
go build -o ..\dist\parken.exe .\cli
cd ..
robocopy . dist parkings.json
robocopy frontend dist\frontend /s
