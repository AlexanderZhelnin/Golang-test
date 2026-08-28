export GOARCH=amd64
export GOEXPERIMENT=simd

go build -ldflags="-s -w" -o go-test-simd
