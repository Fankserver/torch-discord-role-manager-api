package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	storageFile = flag.String("storage", "storage.json", "location of the storage file")
)

type storage struct {
	store     map[int64]string
	storeLock sync.Mutex
	saveLock  sync.Mutex
}

func newStorage() (*storage, error) {
	s := storage{}
	if _, err := os.Stat(*storageFile); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(*storageFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &s.store)
		if err != nil {
			return nil, err
		}
	} else {
		s.store = make(map[int64]string)
	}
	return &s, nil
}

func (s *storage) Add(steamID int64, discordTag string) {
	s.storeLock.Lock()
	s.store[steamID] = discordTag
	s.storeLock.Unlock()

	go func() {
		s.saveLock.Lock()
		defer s.saveLock.Unlock()

		s.storeLock.Lock()
		b, err := json.Marshal(s.store)
		s.storeLock.Unlock()
		if err != nil {
			logrus.Errorln(err)
			return
		}

		err = ioutil.WriteFile(*storageFile, b, 0666)
		if err != nil {
			logrus.Errorln(err)
			return
		}
	}()
}

func (s *storage) GetDiscordTag(steamID int64) string {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()
	return s.store[steamID]
}
