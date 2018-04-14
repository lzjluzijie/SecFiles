package core

import (
	"io/ioutil"
	"log"
	"os"
	"time"
)

type SeedS struct {
	Name  string
	BDUSS string

	Key    []byte `json:"-"`
	B36Key string `json:"Key"`

	Seeds []*Seed

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (seedS *SeedS) AddSeed(seed *Seed) {
	for _, s := range seedS.Seeds {
		if s.B36Hash == seed.B36Hash {
			log.Printf("%s already exist", seed.Name)
		}
	}

	seedS.Seeds = append(seedS.Seeds, seed)
	return
}

func (seedS *SeedS) AddSeeds(p string) (err error) {
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return
	}

	for _, fi := range files {
		if fi.IsDir() {
			err := seedS.AddSeeds(p + string(os.PathSeparator) + fi.Name())
			if err != nil {
				return err
			}
			continue
		}

		s, err := OpenSeed(p + string(os.PathSeparator) + fi.Name())
		if err != nil {
			return err
		}

		seedS.AddSeed(s)
		continue
	}

	return
}
