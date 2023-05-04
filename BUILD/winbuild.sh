
#!/bin/bash
#sudo apt install osslsigncode
path="BUILD/builds"
curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`


# Check if an argument was passed in
if [ $# -eq 1 ]; then
  # Overwrite curTime with the argument
  curTime=$1
fi

go build -ldflags="-X main.buildTime=$curTime"
versionString=`./Facility38 --version`

GOGC=100 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug=true" -o $path/Facility38-win64-raw.exe

cd $path
osslsigncode sign -certs ~/code-cert/cert-20230123-193538.crt -key ~/code-cert/key-20230123-193538.pem -t http://timestamp.digicert.com -h sha256 -in Facility38-win64-raw.exe -out Facility38-$versionString-win64.exe 
zip Facility38-$versionString-win64.zip Facility38-$versionString-win64.exe
scp -P 5313 Facility38-$versionString-win64.zip dist@m45sci.xyz:~/F38Auth/www/dl/
rm Facility38-win64-raw.exe Facility38-$versionString-win64.exe