package commands

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/lzjluzijie/secfiles/core"
	"github.com/urfave/cli"
)

func Add(ctx *cli.Context) (err error) {
	seedS := new(core.SeedS)

	data, err := ioutil.ReadFile("seeds.json")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	err = json.Unmarshal(data, seedS)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	for _, p := range ctx.Args() {
		fi, err := os.Stat(p)
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}

		//Is dir
		if fi.IsDir() {
			log.Printf("Add dir %s", fi.Name())
			err = seedS.AddSeeds(wd + string(os.PathSeparator) + fi.Name())
			if err != nil {
				log.Fatalln(err.Error())
				continue
			}
			continue
		}

		//Is file
		s, err := core.OpenSeed(p)
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}

		log.Printf("Add file %s", s.Path)

		err = seedS.AddSeed(s)
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}

		log.Printf("Added file %s", s.Name)
		continue
	}

	seedS.UpdatedAt = time.Now()

	data, err = json.MarshalIndent(seedS, "", "    ")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	err = ioutil.WriteFile("seeds.json", data, 0600)
	return
}
