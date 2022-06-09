cd backend
go build -o ..\dist\parken.exe .\cmd
cd ..\dist
.\parken.exe %*
