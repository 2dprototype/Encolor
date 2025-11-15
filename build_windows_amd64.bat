mkdir build
set GOARCH=amd64
set GOOS=windows
go build -o build/encolor.exe  -ldflags "-s -w" main.go
pause