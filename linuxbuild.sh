rm GameTest *.zip

GOGC=10 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o GameTest
zip -9 GameTest-linux64.zip GameTest data/gfx/* data/gfx/*/*.png
