GOGC=10 GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o GameTest.exe
upx -9 -f GameTest.exe
#zip GameTest.zip GameTest.exe data/gfx/* data/gfx/*/*.png
zip GameTest-win64.zip GameTest.exe