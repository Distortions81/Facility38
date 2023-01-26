
#sudo apt install osslsigncode
GOGC=10 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o GameTest-win64-raw.exe
osslsigncode sign -certs ~/code-cert/cert-20230123-193538.crt -key ~/code-cert/key-20230123-193538.pem -t http://timestamp.digicert.com -h sha256 -in GameTest-win64-raw.exe -out GameTest-win64.exe 
zip GameTest-win64.zip GameTest-win64.exe data/gfx/* data/gfx/*/*.png
rm GameTest-win64-raw.exe GameTest-win64.exe