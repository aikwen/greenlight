package data

import "time"

type Movie struct {
	ID       int64     `json:"id"`
	CreateAt time.Time `json:"create_at"`
	Title    string    `json:"title"`
	Year     int32     `json:"year"`
	Runtime  Runtime   `json:"runtime"`
	Genres   []string  `json:"genres"`
	Version  int32     `json:"version"`
}
