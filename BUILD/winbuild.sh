
#!/bin/bash
#sudo apt install osslsigncode
curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

# Check if an argument was passed in
if [ $# -eq 1 ]; then
  # Overwrite curTime with the argument
  curTime=$1
fi


GOGC=100 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug=true" -o BUILD/builds/Facility38-win64-raw.exe
osslsigncode sign -certs ~/code-cert/cert-20230123-193538.crt -key ~/code-cert/key-20230123-193538.pem -t http://timestamp.digicert.com -h sha256 -in BUILD/builds/Facility38-win64-raw.exe -out BUILD/builds/Facility38-$curTime-win64.exe 
zip BUILD/builds/Facility38-$curTime-win64.zip BUILD/builds/Facility38-$curTime-win64.exe
rm BUILD/builds/Facility38-win64-raw.exe BUILD/builds/Facility38-$curTime-win64.exe