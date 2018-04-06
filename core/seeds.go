package core

import "time"

type SeedS struct {
	Name  string
	BDUSS string

	Key    []byte `json:"-"`
	B36Key string `json:"Key"`

	Seeds []*Seed

	CreatedAt time.Time
	UpdatedAt time.Time
}
