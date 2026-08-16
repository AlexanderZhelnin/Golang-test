@rem Экспериментальная SIMD-сборка: включает simd/archsimd (Go 1.25+)
set GOARCH=amd64
set GOEXPERIMENT=simd
go build -ldflags="-s -w" -o test-simd.exe
