rmdir dist /s /q
cd backend
go build -o ..\dist\parken.exe
cd ..
robocopy . dist config.json parkings.json 
robocopy frontend dist\frontend /s
