mkdir build
set GOARCH=arm64
set GOOS=android
go build -o build/encolor_android_arm64 -ldflags "-s -w" main.go
pause