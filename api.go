package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"stash-plugin-duplicate-finder/internal/plugin/common"
	"stash-plugin-duplicate-finder/internal/plugin/common/log"
	"stash-plugin-duplicate-finder/internal/plugin/util"

	"github.com/rivo/duplo"
	"github.com/shurcooL/graphql"
)

const spriteSuffix = "_sprite.jpg"

type api struct {
	stopping       bool
	cfg            config
	client         *graphql.Client
	cache          *sceneCache
	duplicateTagID *graphql.ID
}

func main() {
	if len(os.Args) > 1 {
		cmdMain()
		return
	}

	// serves the plugin, providing an object that satisfies the
	// common.RPCRunner interface
	err := common.ServePlugin(&api{})
	if err != nil {
		panic(err)
	}
}

func (a *api) Stop(input struct{}, output *bool) error {
	log.Info("Stopping...")
	a.stopping = true
	*output = true
	return nil
}

// Run is the main work function of the plugin. It interprets the input and
// acts accordingly.
func (a *api) Run(input common.PluginInput, output *common.PluginOutput) error {
	err := a.runImpl(input)

	if err != nil {
		errStr := err.Error()
		*output = common.PluginOutput{
			Error: &errStr,
		}
		return nil
	}

	outputStr := "ok"
	*output = common.PluginOutput{
		Output: &outputStr,
	}

	return nil
}

func (a *api) runImpl(input common.PluginInput) (err error) {
	defer func() {
		// handle panic
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\nstacktrace: %s", r, string(debug.Stack()))
		}
	}()

	pluginDir := input.ServerConnection.PluginDir
	cfg, err := readConfig(filepath.Join(pluginDir, "duplicate-finder.cfg"))
	if err != nil {
		return fmt.Errorf("error reading configuration file: %s", err.Error())
	}

	a.cfg = *cfg
	if !filepath.IsAbs(a.cfg.DBFilename) {
		a.cfg.DBFilename = filepath.Join(pluginDir, a.cfg.DBFilename)
	}

	a.client = util.NewClient(input.ServerConnection)
	a.cache = newSceneCache(a.client)

	if cfg.AddTagName != "" {
		tagID, err := getDuplicateTagId(a.client, cfg.AddTagName)
		if err != nil {
			return err
		}

		if tagID == nil {
			return fmt.Errorf("could not find tag with name %s", cfg.AddTagName)
		}

		a.duplicateTagID = tagID
		log.Debugf("Duplicate tag id = %v", *a.duplicateTagID)
	}

	// find where the generated sprite files are stored
	path, err := getSpriteDir(a.client)
	if err != nil {
		return err
	}

	log.Debugf("Sprite directory is: %s", path)

	log.Info("Processing files for perceptual hashes...")
	m := make(matchInfoMap)
	foundDupes := 0

	hdFunc := func(checksum string, matches duplo.Matches) {
		if len(matches) > 0 {
			foundDupes++
			for _, match := range matches {
				m.add(checksum, match.ID.(string), match.Score)
				a.logDuplicate(checksum, match)
				a.handleDuplicate(m, checksum, true)
			}
		}
	}

	err = a.processFiles(path, hdFunc)
	if err != nil {
		return err
	}

	log.Infof("Found %d duplicate scenes", foundDupes)
	return nil
}

type handleDuplicatesFunc func(checksum string, matches duplo.Matches)

func (a *api) processFiles(path string, hdFunc handleDuplicatesFunc) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	// read the store
	store := duplo.New()
	readDB(store, a.cfg.DBFilename)
	total := len(files)

	for i, f := range files {
		if a.stopping {
			break
		}

		log.Progress(float64(i) / float64(total))

		fn := filepath.Join(path, f.Name())
		if err := a.processFile(fn, store, hdFunc); err != nil {
			log.Errorf("Error processing file %s: %s", f.Name(), err.Error())
		}
	}

	storeDB(store, a.cfg.DBFilename)

	return nil
}

func (a *api) processFile(fn string, store *duplo.Store, hdFunc handleDuplicatesFunc) error {
	if !isSpriteFile(fn) {
		return nil
	}

	checksum := getChecksum(fn)
	existing := store.Has(checksum)
	if existing && a.cfg.NewOnly {
		return nil
	}

	hash, err := getImageHash(fn)
	if err != nil {
		return err
	}

	matches := getHashMatches(store, checksum, *hash, a.cfg.Threshold)

	// remove any matches that no longer exist
	var filteredMatches duplo.Matches
	path := filepath.Dir(fn)
	for _, m := range matches {
		dupeSprite := getSpriteFilename(path, m.ID.(string))
		if _, err := os.Stat(dupeSprite); os.IsNotExist(err) {
			store.Delete(m.ID)
		} else {
			filteredMatches = append(filteredMatches, m)
		}
	}

	hdFunc(checksum, filteredMatches)

	if !existing {
		store.Add(checksum, *hash)
	}

	return nil
}

func (a *api) logDuplicate(checksum string, match *duplo.Match) {
	subject, err := a.cache.get(checksum)
	if err != nil {
		log.Errorf("error getting scene with checksum %s: %s", checksum, err.Error())
		return
	}

	s, err := a.cache.get(match.ID.(string))
	if err != nil {
		log.Errorf("error getting scene with checksum %s: %s", match.ID.(string), err.Error())
		return
	}

	log.Infof("Duplicate: %s - %s (score: %.f)", subject.ID, s.ID, -match.Score)
}

func (a *api) handleDuplicate(m matchInfoMap, checksum string, recurse bool) {
	matches := m[checksum]
	subject, err := a.cache.get(checksum)
	if err != nil {
		log.Errorf("error getting scene with checksum %s: %s", checksum, err.Error())
		return
	}

	newDetails := "=== Duplicate finder plugin ==="
	for _, match := range matches {
		s, err := a.cache.get(match.other)
		if err != nil {
			log.Errorf("error getting scene with checksum %s: %s", match, err.Error())
			continue
		}

		newDetails += fmt.Sprintf("\nDuplicate ID: %s (score: %.f)", s.ID, -match.score)

		if recurse {
			a.handleDuplicate(m, match.other, false)
		}
	}
	newDetails += "\n=== End Duplicate finder plugin ==="

	if a.cfg.AddDetails || a.duplicateTagID != nil {
		details := ""
		if subject.Details != nil {
			details = string(*subject.Details)
		}

		if a.cfg.AddDetails {
			newDetails = addDuplicateDetails(details, newDetails)
		} else {
			newDetails = string(details)
		}

		err = updateScene(a.client, *subject, newDetails, a.duplicateTagID)
		if err != nil {
			log.Errorf("Error updating scene: %s", err.Error())
		}
	}
}

func addDuplicateDetails(origDetails, newDetails string) string {
	re := regexp.MustCompile("(?s)=== Duplicate finder plugin ===.*=== End Duplicate finder plugin ===")
	found := re.FindStringIndex(origDetails)
	if found == nil {
		if len(origDetails) > 0 {
			return origDetails + "\n" + newDetails
		}

		return newDetails
	}

	// replace existing
	return re.ReplaceAllString(origDetails, newDetails)
}

func isSpriteFile(fn string) bool {
	return strings.HasSuffix(fn, spriteSuffix)
}

func getChecksum(fn string) string {
	baseName := filepath.Base(fn)
	return strings.Replace(baseName, spriteSuffix, "", -1)
}

func getSpriteFilename(path, checksum string) string {
	return filepath.Join(path, checksum+spriteSuffix)
}
