package model

import "fmt"

type Site struct {
	Url      string `json:"url"`
	IsUp     bool   `json:"isUp"`
	Previous bool   `json:"-"`
}

func (s Site) String() string {

	return fmt.Sprintf("Url{Url: %s, UpDown: %t}", s.Url, s.IsUp)
}
