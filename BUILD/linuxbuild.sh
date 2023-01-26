GOGC=10 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o GameTest-linux64
zip -9 GameTest-linux64.zip GameTest-linux64 data/gfx/* data/gfx/*/*.png
rm GameTest-linux64
