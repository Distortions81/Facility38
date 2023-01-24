
#sudo apt install osslsigncode
rm *.exe *.zip *.upx

GOGC=10 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o GameTest.exe
#upx -9 GameTest.exe
osslsigncode sign -certs ~/code-cert/cert-20230123-193538.crt -key ~/code-cert/key-20230123-193538.pem -t http://timestamp.digicert.com -h sha256 -in GameTest.exe -out GameTest-signed.exe
zip GameTest-win64-signed.zip GameTest-signed.exe data/gfx/* data/gfx/*/*.png
