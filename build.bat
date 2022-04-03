cd backend
go build -o ..\dist\
cd ../frontend
call npm run build -- --mode=%1
cd ..
