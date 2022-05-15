cd backend
go build -o ..\dist\parken.exe .\cli
cd ..\dist
.\parken.exe %*
