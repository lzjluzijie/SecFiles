package core

import (
	"errors"
	"fmt"
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

func (seedS *SeedS) AddSeed(seed *Seed) error {
	for _, s := range seedS.Seeds {
		if s.B36Hash == seed.B36Hash {
			return errors.New(fmt.Sprintf("%s has been already added", seed.Name))
		}
	}

	seedS.Seeds = append(seedS.Seeds, seed)
	return nil
}

func (seedS *SeedS) AddSeeds(p string) (err error) {
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return
	}

	for _, fi := range files {
		if fi.IsDir() {
			log.Printf("Add dir %s", fi.Name())
			err := seedS.AddSeeds(p + string(os.PathSeparator) + fi.Name())
			if err != nil {
				return err
			}
			log.Printf("Added dir %s", fi.Name())
			continue
		}

		s, err := OpenSeed(p + string(os.PathSeparator) + fi.Name())
		if err != nil {
			return err
		}

		log.Printf("Add file %s", s.Path)

		err = seedS.AddSeed(s)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		log.Printf("Added %s", s.Name)
		continue
	}

	return
}
