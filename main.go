package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/go-faster/jx"
	"github.com/valyala/fasthttp"
)

var ls []Legend
var rect = Rect{Left: 1200, Bottom: 50, Right: 4000, Top: 2850}
var pr = DrawPr{LeftTop: []float64{rect.Left, rect.Top}, Scale: 0.37037037037037035, Mashtab: 100}

var STR1 = "asrgfsadf12421"
var STR2 = "asrgfsadf12321"

var cpuWorkPermits = make(chan struct{}, runtime.GOMAXPROCS(0))

const cpuWorkQueueTimeout = 2 * time.Second

var errOverloaded = errors.New("server overloaded, try again")

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

func mapHandler(ctx *fasthttp.RequestCtx) {
	args := ctx.QueryArgs()
	x := parseCoordinate(args.Peek("x")) / 100
	y := parseCoordinate(args.Peek("y")) / 100

	n, err := runCPUWork(ctx, func() int {
		arena := getArena()
		defer putArena(arena)

		result := build(arena, ls,
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

	_, err := runCPUWork(ctx, func() struct{} {
		arena := getArena()
		defer putArena(arena)

		resultLs := build(arena, ls,
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
	plan, _ := os.ReadFile("primitives.json")

	err := json.Unmarshal(plan, &ls)
	if err != nil {
		fmt.Print("Не могу прочитать json ", err)
	}

	handler := func(ctx *fasthttp.RequestCtx) {
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
		Handler:         handler,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		WriteBufferSize: 512 * 1024,
	}

	fmt.Println("ListenAndServe [localhost:4000]")
	httpErr := server.ListenAndServe("localhost:4000")

	if httpErr != nil {
		panic(httpErr)
	}
}
