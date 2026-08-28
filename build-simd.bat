set GOARCH=amd64
set GOEXPERIMENT=simd

go build -ldflags="-s -w" -o go-test-simd
