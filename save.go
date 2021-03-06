package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {
	gob.Register(potion(0))
	gob.Register(projectile(0))
	gob.Register(&simpleEvent{})
	gob.Register(&monsterEvent{})
	gob.Register(&cloudEvent{})
	gob.Register(armour(0))
	gob.Register(weapon(0))
	gob.Register(shield(0))
}

func (g *game) DataDir() (string, error) {
	var xdg string
	if os.Getenv("GOOS") == "windows" {
		xdg = os.Getenv("LOCALAPPDATA")
	} else {
		xdg = os.Getenv("XDG_DATA_HOME")
	}
	if xdg == "" {
		xdg = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	dataDir := filepath.Join(xdg, "boohu")
	_, err := os.Stat(dataDir)
	if err != nil {
		err = os.MkdirAll(dataDir, 0755)
		if err != nil {
			return "", fmt.Errorf("%v\n", err)
		}
	}
	return dataDir, nil
}

func (g *game) Save() {
	dataDir, err := g.DataDir()
	if err != nil {
		g.Print(err.Error())
		return
	}
	saveFile := filepath.Join(dataDir, "save.gob")
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err = enc.Encode(g)
	if err != nil {
		g.Print(err.Error())
		return
	}
	err = ioutil.WriteFile(saveFile, data.Bytes(), 0644)
	if err != nil {
		g.Print(err.Error())
	}
}

func (g *game) RemoveSaveFile() {
	dataDir, err := g.DataDir()
	if err != nil {
		g.Print(err.Error())
		return
	}
	saveFile := filepath.Join(dataDir, "save.gob")
	_, err = os.Stat(saveFile)
	if err == nil {
		err := os.Remove(saveFile)
		if err != nil {
			fmt.Fprint(os.Stderr, "Error removing old save file")
		}
	}
}

func (g *game) Load() (bool, error) {
	dataDir, err := g.DataDir()
	if err != nil {
		return false, err
	}
	saveFile := filepath.Join(dataDir, "save.gob")
	_, err = os.Stat(saveFile)
	if err != nil {
		// no save file, new game
		return false, err
	}
	data, err := ioutil.ReadFile(saveFile)
	if err != nil {
		return true, err
	}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var lg game
	err = dec.Decode(&lg)
	if err != nil {
		return true, err
	}
	*g = lg
	return true, nil
}
