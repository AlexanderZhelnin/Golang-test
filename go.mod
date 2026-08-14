module test

go 1.26.0

require (
	github.com/segmentio/encoding v0.5.4
	github.com/valyala/fasthttp v1.73.0
)

require (
	github.com/andybalholm/brotli v1.2.2 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

tool (
	test/tools/bench
	test/tools/build
)
