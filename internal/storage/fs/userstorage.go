package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/ddvk/rmfakecloud/internal/model"
	log "github.com/sirupsen/logrus"
)

const (
	userDir     = "users"
	profileName = ".userprofile"
)

// GetUser returns the user using id/email
func (fs *Storage) GetUser(uid string) (response *model.User, err error) {
	profilePath := fs.getPathFromUser(uid, profileName)

	if _, _err := os.Stat(profilePath); os.IsNotExist(_err) {		
		return
	}

	var f *os.File
	f, err = os.Open(profilePath)
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
	usersDir := path.Join(fs.Cfg.DataDir, userDir)

	entries, err := ioutil.ReadDir(usersDir)
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
	userPath := fs.getUserPath(u.Id)

	// Create the user's directory
	err = os.MkdirAll(userPath, 0700)
	if err != nil {
		return
	}

	profilePath := fs.getPathFromUser(u.Id, profileName)
	// Create the profile file
	var js []byte
	js, err = json.Marshal(u)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(profilePath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		log.Warn("cant open: ", profilePath)
		return err
	}
	defer f.Close()
	_, err = f.Write(js)
	if err != nil {
		log.Warn("could not write ", profilePath)
		return err
	}

	return
}

func (fs *Storage) UpdateUser(u *model.User) (err error) {

	err = os.MkdirAll(fs.getUserPath(u.Id), 0700)
	if err != nil {
		return
	}

	profilePath := fs.getPathFromUser(u.Id, profileName)
	// Overwrite the profile
	var js []byte
	js, err = json.Marshal(u)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(profilePath, js, 0600)

	return
}
