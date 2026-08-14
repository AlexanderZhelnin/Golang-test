package main

type MashtabRange struct {
	Min float64
	Max float64
}

type Block struct {
	Id     int64
	Size   float64
	Scaled bool
}

type Fill struct {
	Color1         string
	Color2         string
	Scaled         bool
	Style          int
	FillHatchStyle int
	GradientStyle  int
	Block          Block
}

type Border struct {
	Color      string
	Style      int
	DashStyle  int
	StartCap   int
	EndCap     int
	DashCap    int
	Orientated int
	Scaled     bool
	Size       float64
}

type Font struct {
	Family string
	Size   float64
	Style  int
}

type Text struct {
	Color        string
	MashtabRange MashtabRange
	MashtabBase  float64
	Scaled       bool
	Position     int
	IsAnalyze    bool
	Font         Font
}

type Rect struct {
	Left   float64
	Right  float64
	Top    float64
	Bottom float64
}

type Primitive struct {
	Name       string
	Coords     []float64
	TextCoordX float64
	TextCoordY float64
	TextAngle  float64
	Rect       Rect
}

type Legend struct {
	Id           int64
	Type         uint
	MashtabRange MashtabRange
	Priority     int
	Block        Block
	Fill         Fill
	Border       Border
	Text         Text
	Primitives   []Primitive
}

type DrawPr struct {
	LeftTop [2]float64
	Scale   float64
	Mashtab float64
}

type Layer struct {
	LegendId int64   `json:"legendId"`
	Obrazes  []Obraz `json:"obrazes"`
}

type ResultClip struct {
	Coords []float64
}

type Obraz struct {
	Name   string    `json:"name"`
	Coords []float64 `json:"coords"`
}
