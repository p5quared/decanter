package main

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"
)

type FileTokenStore struct {
	filename string
	dir      string
}

func NewFileTokenStore(filename string) *FileTokenStore {
	return &FileTokenStore{
		filename: filename,
		dir:      os.TempDir(), // Wrong.
	}
}

func (f FileTokenStore) Load() (oauth2.Token, error) {
	return LoadAuthFromFile(f.filename)
}

func (f FileTokenStore) Save(r oauth2.Token) error {
	return SaveAuthToFile(r, f.filename)
}

func SaveAuthToFile(r oauth2.Token, f string) error {
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	os.WriteFile(f, bytes, 0644)
	return nil
}

func LoadAuthFromFile(f string) (oauth2.Token, error) {
	bytes, err := os.ReadFile(f)
	if err != nil {
		return oauth2.Token{}, err
	}

	var r oauth2.Token
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return oauth2.Token{}, err
	}

	return r, nil
}
