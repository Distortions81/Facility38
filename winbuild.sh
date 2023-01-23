rm *.exe *.zip *.upx

GOGC=10 GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o Million-Benchmark.exe
upx -9 --ultra-brute Million-Benchmark.exe
zip -9 Million-Benchmark-win64.zip Million-Benchmark.exe data/gfx/* data/gfx/*/*.png
