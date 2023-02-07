
#!/bin/bash
#sudo apt install osslsigncode
curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=100 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug=true -X main.UPSBench=true -X main.LoadTest=true" -o BUILD/builds/GameTest-win64-raw.exe
osslsigncode sign -certs ~/code-cert/cert-20230123-193538.crt -key ~/code-cert/key-20230123-193538.pem -t http://timestamp.digicert.com -h sha256 -in BUILD/builds/GameTest-win64-raw.exe -out BUILD/builds/GameTest-$curTime-win64-bench.exe 
zip BUILD/builds/GameTest-$curTime-win64-bench.zip BUILD/builds/GameTest-$curTime-win64-bench.exe
rm BUILD/builds/GameTest-win64-raw.exe BUILD/builds/GameTest-$curTime-win64-bench.exe