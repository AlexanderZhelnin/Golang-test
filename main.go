package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		xStr := query["x"][0]
		yStr := query["y"][0]

		x, _ := strconv.ParseFloat(xStr, 64)
		y, _ := strconv.ParseFloat(yStr, 64)

		x /= 100
		y /= 100

		resultLs := build(ls,
			&DrawPr{LeftTop: []float64{pr.LeftTop[0] + x, pr.LeftTop[1] + y}, Scale: pr.Scale, Mashtab: pr.Mashtab},
			&Rect{
				Left:   rect.Left + x,
				Top:    rect.Top + y,
				Right:  rect.Right,
				Bottom: rect.Bottom})

		fmt.Fprint(w, len(resultLs))
	})

	http.HandleFunc("/mapJSON", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		xStr := query["x"][0]
		yStr := query["y"][0]

		x, _ := strconv.ParseFloat(xStr, 64)
		y, _ := strconv.ParseFloat(yStr, 64)

		x /= 100
		y /= 100

		resultLs := build(ls,
			&DrawPr{LeftTop: []float64{pr.LeftTop[0] + x, pr.LeftTop[1] + y}, Scale: pr.Scale, Mashtab: pr.Mashtab},
			&Rect{
				Left:   rect.Left + x,
				Top:    rect.Top + y,
				Right:  rect.Right,
				Bottom: rect.Bottom})

		b, err := json.Marshal(resultLs[:5])

		if err != nil {
			fmt.Print("Не могу сохранить json ", err)
		}

		w.Header().Set("Content-Type", "application/json")

		w.Write(b)
	})

	http.HandleFunc("/naturalsort", func(w http.ResponseWriter, r *http.Request) {
		result := 0
		for i := range 10000 {
			result += Compare(STR1+strconv.Itoa(i), STR2+strconv.Itoa(i))
		}

		fmt.Fprint(w, result)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World golang!")
	})

	fmt.Println("ListenAndServe [localhost:4000]")
	httpErr := http.ListenAndServe("localhost:4000", nil)

	if httpErr != nil {
		panic(httpErr)
	}
}
