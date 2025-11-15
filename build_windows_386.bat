mkdir build
set GOARCH=386
set GOOS=windows
go build -o build/encolor32.exe  -ldflags "-s -w" main.go
pause