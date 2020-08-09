package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/shurcooL/graphql"
)

type Tag struct {
	ID   graphql.ID     `graphql:"id"`
	Name graphql.String `graphql:"name"`
}

type Scene struct {
	ID      graphql.ID
	Title   *graphql.String
	Path    graphql.String
	Details *graphql.String
	Tags    []Tag
}

func (s Scene) getTagIds() []graphql.ID {
	ret := []graphql.ID{}

	for _, t := range s.Tags {
		ret = append(ret, t.ID)
	}

	return ret
}

type ConfigGeneralResult struct {
	GeneratedPath graphql.String `graphql:"generatedPath"`
}

type ConfigResult struct {
	General ConfigGeneralResult `graphql:"general"`
}

func getSpriteDir(client *graphql.Client) (string, error) {
	var m struct {
		Configuration *ConfigResult `graphql:"configuration"`
	}

	err := client.Query(context.Background(), &m, nil)
	if err != nil {
		return "", fmt.Errorf("Error getting sprite directory from configuration: %s", err.Error())
	}

	ret := filepath.Join(string(m.Configuration.General.GeneratedPath), "vtt")
	return ret, nil
}

func addTagId(tagIds []graphql.ID, tagId graphql.ID) []graphql.ID {
	for _, t := range tagIds {
		if t == tagId {
			return tagIds
		}
	}

	tagIds = append(tagIds, tagId)
	return tagIds
}

func findSceneFromChecksum(client *graphql.Client, checksum string) (*Scene, error) {
	var m struct {
		FindScene *Scene `graphql:"findScene(checksum: $c)"`
	}

	vars := map[string]interface{}{
		"c": graphql.String(checksum),
	}

	err := client.Query(context.Background(), &m, vars)
	if err != nil {
		return nil, err
	}

	return m.FindScene, nil
}

type SceneHashInput struct {
	Oshash *graphql.String `graphql:"oshash" json:"oshash"`
}

func findSceneFromOshash(client *graphql.Client, oshash string) (*Scene, error) {
	var m struct {
		FindScene *Scene `graphql:"findSceneByHash(input: $i)"`
	}

	input := SceneHashInput{
		Oshash: graphql.NewString(graphql.String(oshash)),
	}

	vars := map[string]interface{}{
		"i": input,
	}

	err := client.Query(context.Background(), &m, vars)
	if err != nil {
		return nil, err
	}

	return m.FindScene, nil
}

type SceneUpdate struct {
	ID graphql.ID `graphql:"id"`
}

type BulkUpdateIds struct {
	IDs  []graphql.ID   `graphql:"ids" json:"ids"`
	Mode graphql.String `graphql:"mode" json:"mode"`
}

type BulkSceneUpdateInput struct {
	IDs     []graphql.ID   `graphql:"ids" json:"ids"`
	Details graphql.String `graphql:"details" json:"details"`
	TagIds  *BulkUpdateIds `graphql:"tag_ids" json:"tag_ids"`
}

func updateScene(client *graphql.Client, s Scene, details string, duplicateTagID *graphql.ID) error {
	// use BulkSceneUpdateInput since sceneUpdate requires performers, etc.
	var m struct {
		SceneUpdate []SceneUpdate `graphql:"bulkSceneUpdate(input: $s)"`
	}

	input := BulkSceneUpdateInput{
		IDs:     []graphql.ID{s.ID},
		Details: graphql.String(details),
	}

	if duplicateTagID != nil {
		tagIds := &BulkUpdateIds{
			IDs:  s.getTagIds(),
			Mode: "ADD",
		}
		tagIds.IDs = addTagId(tagIds.IDs, *duplicateTagID)
		input.TagIds = tagIds
	}

	vars := map[string]interface{}{
		"s": input,
	}

	err := client.Mutate(context.Background(), &m, vars)
	if err != nil {
		return err
	}

	return nil
}

func getDuplicateTagId(client *graphql.Client, tagName string) (*graphql.ID, error) {
	var m struct {
		AllTags []Tag `graphql:"allTags"`
	}

	err := client.Query(context.Background(), &m, nil)
	if err != nil {
		fmt.Printf("Error getting tags: %s\n", err.Error())
		return nil, err
	}

	for _, t := range m.AllTags {
		if string(t.Name) == tagName {
			id := t.ID
			return &id, nil
		}
	}

	return nil, err
}
