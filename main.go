package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/go-faster/jx"
	"github.com/valyala/fasthttp"
)

const listenAddress = "127.0.0.1:4000"

// pgoProfileEnv включает тренировочный режим: путь в переменной означает, что
// сервер пишет CPU-профиль для PGO и завершается по строке в stdin.
// В обычном запуске выполняется только один LookupEnv
const pgoProfileEnv = "API_TEST_PGO_PROFILE"

// Общие аннотации для всего API
// @title Тест REST Golang
// @version 1.0

var ls []Legend
var rect = Rect{Left: 1200, Bottom: 50, Right: 4000, Top: 2850}
var pr = DrawPr{LeftTop: []float64{rect.Left, rect.Top}, Scale: 0.37037037037037035, Mashtab: 100}

var STR1 = "asrgfsadf12421"
var STR2 = "asrgfsadf12321"

var STR1R = []rune("asrgfsadf12421")
var STR2R = []rune("asrgfsadf12321")

var cpuWorkPermits = make(chan struct{}, runtime.GOMAXPROCS(0))

const cpuWorkQueueTimeout = 2 * time.Second

var errOverloaded = errors.New("server overloaded, try again")

var float64ArenaPool = newArenaPool[float64]()
var obrazArenaPool = newArenaPool[Obraz]()
var layerArenaPool = newArenaPool[Layer]()
var charArenaPool = newArenaPool[rune]()

func runCPUWork[T any](ctx context.Context, work func() T) (T, error) {
	var zero T

	timeoutCtx, cancel := context.WithTimeout(ctx, cpuWorkQueueTimeout)
	defer cancel()

	select {
	case cpuWorkPermits <- struct{}{}:
	case <-timeoutCtx.Done():
		return zero, errOverloaded
	}
	defer func() { <-cpuWorkPermits }()

	return work(), nil
}

func parseCoordinate(value []byte) float64 {
	x, _ := strconv.ParseFloat(string(value), 64)
	return x
}

func writeInt(ctx *fasthttp.RequestCtx, value int) {
	var buffer [24]byte
	_, _ = ctx.Write(strconv.AppendInt(buffer[:0], int64(value), 10))
}

// @Summary Получение преобразованных геоданных (тест без реального ответа)
// @Description Выбирает геоданные по области, отсекает приметивы по области, оптимизирует координаты, преобразовывает к экранным
// @Produce str
// @Success 200 {int}
func mapHandler(ctx *fasthttp.RequestCtx) {
	args := ctx.QueryArgs()

	x := parseCoordinate(args.Peek("x")) / 100
	y := parseCoordinate(args.Peek("y")) / 100

	resultLs := build(ls,
		&DrawPr{LeftTop: []float64{pr.LeftTop[0] + x, pr.LeftTop[1] + y}, Scale: pr.Scale, Mashtab: pr.Mashtab},
		&Rect{
			Left:   rect.Left + x,
			Top:    rect.Top + y,
			Right:  rect.Right,
			Bottom: rect.Bottom})

	var count int
	for range resultLs {
		count++
	}

	writeInt(ctx, count)
}

func mapHandlerBlazing(ctx *fasthttp.RequestCtx) {
	args := ctx.QueryArgs()
	x := parseCoordinate(args.Peek("x")) / 100
	y := parseCoordinate(args.Peek("y")) / 100

	n, err := runCPUWork(ctx, func() int {
		arenaFloat64 := float64ArenaPool.Get()
		arenaObraz := obrazArenaPool.Get()
		arenaLayer := layerArenaPool.Get()

		defer float64ArenaPool.Put(arenaFloat64)
		defer obrazArenaPool.Put(arenaObraz)
		defer layerArenaPool.Put(arenaLayer)

		result := buildBlazing(arenaFloat64, arenaObraz, arenaLayer, ls,
			&DrawPr{LeftTop: []float64{pr.LeftTop[0] + x, pr.LeftTop[1] + y}, Scale: pr.Scale, Mashtab: pr.Mashtab},
			&Rect{
				Left:   rect.Left + x,
				Top:    rect.Top + y,
				Right:  rect.Right,
				Bottom: rect.Bottom})
		return len(result)
	})
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
		return
	}

	writeInt(ctx, n)
}

func mapJSONHandler(ctx *fasthttp.RequestCtx) {
	args := ctx.QueryArgs()
	x := parseCoordinate(args.Peek("x")) / 100
	y := parseCoordinate(args.Peek("y")) / 100

	e := jx.GetEncoder()
	defer jx.PutEncoder(e)

	resultLs := build(ls,
		&DrawPr{LeftTop: []float64{pr.LeftTop[0] + x, pr.LeftTop[1] + y}, Scale: pr.Scale, Mashtab: pr.Mashtab},
		&Rect{
			Left:   rect.Left + x,
			Top:    rect.Top + y,
			Right:  rect.Right,
			Bottom: rect.Bottom})

	i := 0
	result := make([]Layer, 5)
	for l := range resultLs {
		if i < 5 {
			result[i] = l
		}
		i++
	}

	encodeLayers(e, result)
	ctx.SetContentType("application/json")
	ctx.Write(e.Bytes())
}

func mapJSONBlazingHandler(ctx *fasthttp.RequestCtx) {
	args := ctx.QueryArgs()
	x := parseCoordinate(args.Peek("x")) / 100
	y := parseCoordinate(args.Peek("y")) / 100

	e := jx.GetEncoder()
	defer jx.PutEncoder(e)

	_, err := runCPUWork(ctx, func() struct{} {

		arenaFloat64 := float64ArenaPool.Get()
		arenaObraz := obrazArenaPool.Get()
		arenaLayer := layerArenaPool.Get()

		defer float64ArenaPool.Put(arenaFloat64)
		defer obrazArenaPool.Put(arenaObraz)
		defer layerArenaPool.Put(arenaLayer)

		resultLs := buildBlazing(arenaFloat64, arenaObraz, arenaLayer, ls,
			&DrawPr{LeftTop: []float64{pr.LeftTop[0] + x, pr.LeftTop[1] + y}, Scale: pr.Scale, Mashtab: pr.Mashtab},
			&Rect{
				Left:   rect.Left + x,
				Top:    rect.Top + y,
				Right:  rect.Right,
				Bottom: rect.Bottom})

		if len(resultLs) > 5 {
			resultLs = resultLs[:5]
		}
		encodeLayers(e, resultLs)
		return struct{}{}
	})
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
		return
	}

	ctx.SetContentType("application/json")
	ctx.Write(e.Bytes())
}

func naturalSortHandler(ctx *fasthttp.RequestCtx) {
	result := 0
	for i := range 10000 {
		result += Compare(STR1+strconv.Itoa(i), STR2+strconv.Itoa(i))
	}

	writeInt(ctx, result)
}

func appendInt(mas []rune, num int) {

	digitsValue := num
	digits := 1

	for digitsValue >= 100_000 {
		digitsValue /= 100_000
		digits += 5
	}

	if digitsValue >= 100_000 {
		digitsValue /= 100_000
		digits += 5
	}
	if digitsValue < 10 {
	} else if digitsValue < 100 {
		digits = digits + 1
	} else if digitsValue < 1000 {
		digits = digits + 2
	} else if digitsValue < 10_000 {
		digits = digits + 3
	} else {
		digits = 4
	}

	if digits == 1 {
		mas[0] = rune(num%10 + '0')
		return
	}

	index := digits - 1

	for num > 0 {

		mas[index] = rune(num%10 + '0')
		index--
		num /= 10
	}
}

func naturalSortBlazingHandler(ctx *fasthttp.RequestCtx) {

	result := 0

	arena := charArenaPool.Get()
	defer charArenaPool.Put(arena)

	for i := range 10000 {

		s1 := arena.allocFull(len(STR1) + 5)
		copy(s1, STR1R)

		appendInt(s1[len(STR1):], i)

		s2 := arena.allocFull(len(STR2) + 5)
		copy(s2, STR2R)
		appendInt(s2[len(STR2):], i)

		result += CompareRunes(s1, s2)
	}

	writeInt(ctx, result)
}

func naturalSortHackHandler(ctx *fasthttp.RequestCtx) {
	buf1 := append(make([]byte, 0, len(STR1)+5), STR1...)
	buf2 := append(make([]byte, 0, len(STR2)+5), STR2...)

	result := 0
	for i := range 10000 {
		s1 := string(strconv.AppendInt(buf1[:len(STR1)], int64(i), 10))
		s2 := string(strconv.AppendInt(buf2[:len(STR2)], int64(i), 10))
		result += Compare(s1, s2)
	}

	writeInt(ctx, result)
}

func rootHandler(ctx *fasthttp.RequestCtx) {
	_, _ = ctx.WriteString("Hello World golang!")
}

func main() {
	fmt.Println("Тестовый сервер Golang")
	plan, _ := os.ReadFile("primitives.json")

	err := json.Unmarshal(plan, &ls)
	if err != nil {
		fmt.Print("Не могу прочитать json ", err)
	}

	// r := http.NewServeMux()

	// r.HandleFunc("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:4000/swagger/doc.json")))

	handler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/map":
			mapHandler(ctx)
		case "/mapBlazing":
			mapHandlerBlazing(ctx)
		case "/mapJSON":
			mapJSONHandler(ctx)
		case "/mapJSONBlazing":
			mapJSONBlazingHandler(ctx)
		case "/naturalsort":
			naturalSortHandler(ctx)
		case "/naturalsortBlazing":
			naturalSortBlazingHandler(ctx)
		case "/":
			rootHandler(ctx)
		default:
			ctx.Error("404 page not found", fasthttp.StatusNotFound)
		}
	}

	server := &fasthttp.Server{
		Handler:         handler,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		WriteBufferSize: 512 * 1024,
	}

	fmt.Println("ListenAndServe [" + listenAddress + "]")

	if profilePath, training := os.LookupEnv(pgoProfileEnv); training {
		if err := serveWithProfile(server, profilePath); err != nil {
			panic(err)
		}
		return
	}

	if httpErr := server.ListenAndServe(listenAddress); httpErr != nil {
		panic(httpErr)
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
