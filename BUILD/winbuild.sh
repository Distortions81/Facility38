
#!/bin/bash
#sudo apt install osslsigncode
curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=100 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.buildTime=$curTime" -o GameTest-win64-raw.exe
osslsigncode sign -certs ~/code-cert/cert-20230123-193538.crt -key ~/code-cert/key-20230123-193538.pem -t http://timestamp.digicert.com -h sha256 -in GameTest-win64-raw.exe -out GameTest-$curTime-win64.exe 
zip GameTest-$curTime-win64.zip GameTest-$curTime-win64.exe
rm GameTest-win64-raw.exe GameTest-$curTime-win64.exe