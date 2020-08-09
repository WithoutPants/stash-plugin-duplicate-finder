package main

import (
	"fmt"

	"github.com/shurcooL/graphql"
)

type sceneCache struct {
	scenes map[string]*Scene
	client *graphql.Client
}

func newSceneCache(client *graphql.Client) *sceneCache {
	return &sceneCache{
		scenes: make(map[string]*Scene),
		client: client,
	}
}

func (c *sceneCache) get(hash string) (*Scene, error) {
	if c.scenes[hash] != nil {
		return c.scenes[hash], nil
	}

	var ret *Scene
	var err error
	if len(hash) == 32 {
		ret, err = findSceneFromChecksum(c.client, hash)
		if err != nil {
			return nil, err
		}
	} else if len(hash) == 16 {
		ret, err = findSceneFromOshash(c.client, hash)
		if err != nil {
			return nil, err
		}
	}

	if ret == nil {
		return nil, fmt.Errorf("scene with hash %s is nil", hash)
	}

	c.scenes[hash] = ret
	return ret, nil
}
