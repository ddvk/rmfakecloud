package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/ddvk/rmfakecloud/internal/model"
)

const (
	userDir     = "users"
	profileName = ".userprofile"
)

// GetUser blah
func (fs *Storage) GetUser(id string) (response *model.User, err error) {
	dataDir := fs.Cfg.DataDir
	fullPath := path.Join(dataDir, userDir, id, profileName)

	var f *os.File
	f, err = os.Open(fullPath)
	if err != nil {
		return
	}
	defer f.Close()

	var content []byte
	content, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}

	response = &model.User{}
	err = json.Unmarshal(content, response)
	if err != nil {
		return
	}

	return
}

// GetUsers blah
func (fs *Storage) GetUsers() (users []*model.User, err error) {
	dataDir := path.Join(fs.Cfg.DataDir, userDir)

	entries, err := ioutil.ReadDir(dataDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if user, err := fs.GetUser(entry.Name()); err == nil {
				users = append(users, user)
			}
		}
	}
	return
}

// RegisterUser blah
func (fs *Storage) RegisterUser(u *model.User) (err error) {
	userDir := path.Join(fs.Cfg.DataDir, userDir, u.Id)
	profilePath := path.Join(userDir, profileName)

	// Create the user's directory
	err = os.MkdirAll(userDir, 0700)
	if err != nil {
		return
	}

	// Create the profile file
	var js []byte
	js, err = json.Marshal(u)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(profilePath, os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(js)
	if err != nil {
		return err
	}

	return
}

func (fs *Storage) UpdateUser(u *model.User) (err error) {
	userDir := path.Join(fs.Cfg.DataDir, u.Id)
	profilePath := path.Join(userDir, profileName)

	// Erase the profile file
	var js []byte
	js, err = json.Marshal(u)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(profilePath, js, 0600)

	return
}
