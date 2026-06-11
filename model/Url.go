package model

import "fmt"

type Url struct {
	Url      string `json:"url"`
	IsUp     bool   `json:"isUp"`
	Previous bool   `json:"-"`
}

func (u Url) String() string {
	return fmt.Sprintf("Url{Url: %s, UpDown: %t}", u.Url, u.IsUp)
}
