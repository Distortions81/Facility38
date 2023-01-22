GOGC=10 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o GameTest.exe
upx --brute GameTest.exe
#zip GameTest.zip GameTest.exe data/gfx/* data/gfx/*/*.png
zip GameTest.zip GameTest.exe