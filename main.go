package main

import (
	"bufio"
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"

	segmentjson "github.com/segmentio/encoding/json"
	"github.com/valyala/fasthttp"
)

const listenAddress = "127.0.0.1:4000"

// pgoProfileEnv включает тренировочный режим: путь в переменной означает, что
// сервер пишет CPU-профиль для PGO и завершается по строке в stdin.
// В обычном запуске выполняется только один LookupEnv
const pgoProfileEnv = "API_TEST_PGO_PROFILE"

var legends []Legend

var mapRect = Rect{Left: 1200, Bottom: 50, Right: 4000, Top: 2850}

var mapProperties = DrawPr{
	LeftTop: [2]float64{mapRect.Left, mapRect.Top},
	Scale:   0.37037037037037035,
	Mashtab: 100,
}

const (
	sortPrefixFirst  = "asrgfsadf12421"
	sortPrefixSecond = "asrgfsadf12321"
)

// Все CPU-bound эндпоинты проходят через ограничитель ёмкостью GOMAXPROCS
// ожидающие горутины становятся в FIFO-очередь, одновременно считают
// не более GOMAXPROCS запросов. Отдельный пул потоков не нужен:
// горутины уже мультиплексируются планировщиком на
// GOMAXPROCS ядер, ограничивать надо только параллелизм
var cpuWorkPermits = make(chan struct{}, runtime.GOMAXPROCS(0))

var mapJSONBufferPool = sync.Pool{New: func() any {
	buffer := make([]byte, 0, 512*1024)
	return &buffer
}}

const mapJSONFlags = segmentjson.EscapeHTML | segmentjson.SortMapKeys

func runCPUWork[T any](work func() T) T {
	cpuWorkPermits <- struct{}{}
	defer func() { <-cpuWorkPermits }()
	return work()
}

func loadLegends() error {
	data, err := os.ReadFile("primitives.json")
	if err != nil {
		return fmt.Errorf("read primitives.json: %w", err)
	}
	if err := stdjson.Unmarshal(data, &legends); err != nil {
		return fmt.Errorf("decode primitives.json: %w", err)
	}
	return nil
}

func parseCoordinate(value []byte) (float64, error) {
	if len(value) == 0 {
		return 0, nil
	}
	return strconv.ParseFloat(string(value), 64)
}

func drawInput(ctx *fasthttp.RequestCtx) (DrawPr, Rect, error) {
	args := ctx.QueryArgs()
	x, err := parseCoordinate(args.Peek("x"))
	if err != nil {
		return DrawPr{}, Rect{}, fmt.Errorf("invalid x: %w", err)
	}
	y, err := parseCoordinate(args.Peek("y"))
	if err != nil {
		return DrawPr{}, Rect{}, fmt.Errorf("invalid y: %w", err)
	}

	x /= 100
	y /= 100

	properties := DrawPr{
		LeftTop: [2]float64{mapProperties.LeftTop[0] + x, mapProperties.LeftTop[1] + y},
		Scale:   mapProperties.Scale,
		Mashtab: mapProperties.Mashtab,
	}
	clipRect := Rect{
		Left:   mapRect.Left + x,
		Top:    mapRect.Top + y,
		Right:  mapRect.Right,
		Bottom: mapRect.Bottom,
	}

	return properties, clipRect, nil
}

func writeJSONInt(ctx *fasthttp.RequestCtx, value int) {
	ctx.SetContentType("application/json; charset=utf-8")
	var buffer [24]byte
	_, _ = ctx.Write(strconv.AppendInt(buffer[:0], int64(value), 10))
}

func mapHandler(ctx *fasthttp.RequestCtx) {
	properties, rect, err := drawInput(ctx)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	count := runCPUWork(func() int {
		scratch := acquireBuildScratch()
		defer releaseBuildScratch(scratch, mapScratchRetention)

		count := 0
		visitLayers(legends, &properties, &rect, scratch, func(Layer) bool {
			count++
			scratch.result.reset()
			scratch.temporary.reset()
			scratch.obrazes.reset()
			return true
		})
		return count
	})
	writeJSONInt(ctx, count)
}

type encodedJSON struct {
	body   []byte
	buffer *[]byte
	err    error
}

func mapJSONHandler(ctx *fasthttp.RequestCtx) {
	properties, rect, err := drawInput(ctx)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	encoded := runCPUWork(func() encodedJSON {
		scratch := acquireBuildScratch()
		defer releaseBuildScratch(scratch, mapJSONScratchRetention)

		// Как C# Drawer.Build(...).Take(5): сначала строятся все слои
		result := build(legends, &properties, &rect, scratch)
		if len(result) > 5 {
			result = result[:5]
		}
		buffer := mapJSONBufferPool.Get().(*[]byte)
		body, err := segmentjson.Append((*buffer)[:0], result, mapJSONFlags)
		return encodedJSON{body: body, buffer: buffer, err: err}
	})
	if encoded.err != nil {
		releasePooledJSONBuffer(encoded.body, encoded.buffer)
		ctx.Error("encode JSON", fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetContentType("application/json; charset=utf-8")
	body := &pooledJSONBody{buffer: encoded.buffer, body: encoded.body}
	body.reader.Reset(encoded.body)
	ctx.SetBodyStream(body, len(encoded.body))
}

type pooledJSONBody struct {
	reader bytes.Reader
	buffer *[]byte
	body   []byte
}

func (p *pooledJSONBody) Read(destination []byte) (int, error) {
	return p.reader.Read(destination)
}

func (p *pooledJSONBody) Close() error {
	releasePooledJSONBuffer(p.body, p.buffer)
	return nil
}

func releasePooledJSONBuffer(body []byte, buffer *[]byte) {
	if cap(body) <= 1024*1024 {
		*buffer = body[:0]
	} else {
		*buffer = make([]byte, 0, 512*1024)
	}
	mapJSONBufferPool.Put(buffer)
}

var (
	sortPrefixFirstUnits  = widenASCII(sortPrefixFirst)
	sortPrefixSecondUnits = widenASCII(sortPrefixSecond)
)

const sortValueCapacity = 32

func makeSortValue(prefix []uint16, digits []byte) []uint16 {
	value := make([]uint16, 0, sortValueCapacity)
	value = append(value, prefix...)
	for _, digit := range digits {
		value = append(value, uint16(digit))
	}
	return value
}

func naturalSortResult() int {
	result := 0
	var digits [20]byte
	for i := range 10_000 {
		// Число форматируется отдельно для каждой строки, как в C#
		first := makeSortValue(sortPrefixFirstUnits, strconv.AppendInt(digits[:0], int64(i), 10))
		second := makeSortValue(sortPrefixSecondUnits, strconv.AppendInt(digits[:0], int64(i), 10))
		result += compareUTF16(first, second)
	}
	return result
}

func naturalSortHandler(ctx *fasthttp.RequestCtx) {
	writeJSONInt(ctx, runCPUWork(naturalSortResult))
}

func rootHandler(ctx *fasthttp.RequestCtx) {
	_, _ = ctx.WriteString("Hello World golang!")
}

func main() {
	if err := loadLegends(); err != nil {
		panic(err)
	}

	handler := func(ctx *fasthttp.RequestCtx) {
		if !ctx.IsGet() {
			ctx.Error("Method Not Allowed", fasthttp.StatusMethodNotAllowed)
			return
		}
		switch string(ctx.Path()) {
		case "/map":
			mapHandler(ctx)
		case "/mapJSON":
			mapJSONHandler(ctx)
		case "/naturalsort":
			naturalSortHandler(ctx)
		case "/":
			rootHandler(ctx)
		default:
			ctx.Error("404 page not found", fasthttp.StatusNotFound)
		}
	}

	server := &fasthttp.Server{
		Handler: handler,
	}

	fmt.Println("Тестовый сервер Golang")
	fmt.Println("ListenAndServe [" + listenAddress + "]")

	if profilePath, training := os.LookupEnv(pgoProfileEnv); training {
		if err := serveWithProfile(server, profilePath); err != nil {
			panic(err)
		}
		return
	}

	if err := server.ListenAndServe(listenAddress); err != nil {
		panic(err)
	}
}

// обслуживает запросы, записывая CPU-профиль для PGO,
// и завершается по строке в stdin
func serveWithProfile(server *fasthttp.Server, profilePath string) error {
	file, err := os.Create(profilePath)
	if err != nil {
		return err
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return err
	}

	failure := make(chan error, 1)
	go func() { failure <- server.ListenAndServe(listenAddress) }()

	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')

	shutdownErr := server.Shutdown()
	pprof.StopCPUProfile()
	if err := file.Close(); err != nil {
		return err
	}
	if shutdownErr != nil {
		return shutdownErr
	}
	return <-failure
}
