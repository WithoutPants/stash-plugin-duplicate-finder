package main

import (
	"image/jpeg"
	"io/ioutil"
	"os"
	"sort"

	"stash-plugin-duplicate-finder/internal/plugin/common/log"

	"github.com/rivo/duplo"
)

func storeDB(store *duplo.Store, filename string) error {
	data, err := store.GobEncode()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func readDB(store *duplo.Store, filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		// assume no file
		log.Info("Assuming no existing db file. Starting from scratch...")
	}

	err = store.GobDecode(data)
	if err != nil {
		return err
	}

	log.Infof("Read store from file: %d hashes loaded", store.Size())
	return nil
}

func getImageHash(fn string) (*duplo.Hash, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, err
	}

	hash, _ := duplo.CreateHash(img)
	return &hash, nil
}

func getHashMatches(store *duplo.Store, checksum string, hash duplo.Hash, threshold int) duplo.Matches {
	ret := duplo.Matches{}

	matches := store.Query(hash)
	sort.Sort(matches)

	for _, m := range matches {
		// exclude same id
		if checksum == m.ID {
			continue
		}

		if m.Score <= float64(-threshold) {
			ret = append(ret, m)
		}
	}

	return ret
}
