del dist
cd backend
go build -o ..\dist\parken.exe .\cli
cd ..
robocopy frontend dist\frontend