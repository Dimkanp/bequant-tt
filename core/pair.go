package core

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Pair struct {
	ID      uuid.UUID
	Fsym    string
	Tsym    string
	Created time.Time
	Raw     string
	Display string
}

type Compare struct {
	Fsym  string
	Tsyms map[string]*CompareData
}

type CompareData struct {
	Raw     string
	Display string
}
