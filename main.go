package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-faster/jx"
)

var ls []Legend
var rect = Rect{Left: 1200, Bottom: 50, Right: 4000, Top: 2850}
var pr = DrawPr{LeftTop: []float64{rect.Left, rect.Top}, Scale: 0.37037037037037035, Mashtab: 100}

var STR1 = "asrgfsadf12421"
var STR2 = "asrgfsadf12321"

func main() {
	fmt.Println("Тестовый сервер Golang")
	plan, _ := os.ReadFile("primitives.json")

	err := json.Unmarshal(plan, &ls)
	if err != nil {
		fmt.Print("Не могу прочитать json ", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		xStr := query.Get("x")
		yStr := query.Get("y")

		x, _ := strconv.ParseFloat(xStr, 64)
		y, _ := strconv.ParseFloat(yStr, 64)

		x /= 100
		y /= 100

		arena := getArena()
		defer putArena(arena)

		n := build(arena, ls,
			&DrawPr{LeftTop: []float64{pr.LeftTop[0] + x, pr.LeftTop[1] + y}, Scale: pr.Scale, Mashtab: pr.Mashtab},
			&Rect{
				Left:   rect.Left + x,
				Top:    rect.Top + y,
				Right:  rect.Right,
				Bottom: rect.Bottom})

		fmt.Fprint(w, len(n))
	})

	mux.HandleFunc("/mapJSON", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		xStr := query.Get("x")
		yStr := query.Get("y")

		x, _ := strconv.ParseFloat(xStr, 64)
		y, _ := strconv.ParseFloat(yStr, 64)

		x /= 100
		y /= 100

		arena := getArena()
		defer putArena(arena)

		resultLs := build(arena, ls,
			&DrawPr{LeftTop: []float64{pr.LeftTop[0] + x, pr.LeftTop[1] + y}, Scale: pr.Scale, Mashtab: pr.Mashtab},
			&Rect{
				Left:   rect.Left + x,
				Top:    rect.Top + y,
				Right:  rect.Right,
				Bottom: rect.Bottom})

		e := jx.GetEncoder()
		defer jx.PutEncoder(e)

		if len(resultLs) > 5 {
			resultLs = resultLs[:5]
		}
		encodeLayers(e, resultLs)

		w.Header().Set("Content-Type", "application/json")
		w.Write(e.Bytes())
	})

	mux.HandleFunc("/naturalsort", func(w http.ResponseWriter, r *http.Request) {
		buf1 := append(make([]byte, 0, len(STR1)+5), STR1...)
		buf2 := append(make([]byte, 0, len(STR2)+5), STR2...)

		result := 0
		for i := range 10000 {
			s1 := string(strconv.AppendInt(buf1[:len(STR1)], int64(i), 10))
			s2 := string(strconv.AppendInt(buf2[:len(STR2)], int64(i), 10))
			result += Compare(s1, s2)
		}

		fmt.Fprint(w, result)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World golang!")
	})

	server := &http.Server{
		Addr:              "localhost:4000",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Println("ListenAndServe [localhost:4000]")
	httpErr := server.ListenAndServe()

	if httpErr != nil {
		panic(httpErr)
	}
}
