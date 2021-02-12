package fs

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/model"
	log "github.com/sirupsen/logrus"
)

const (
	userDir     = "users"
	profileName = ".userprofile"
)

func NewStorage(cfg *config.Config) *Storage {
	fs := &Storage{
		Cfg: cfg,
	}

	usersPath := fs.getUserPath("")
	err := os.MkdirAll(usersPath, 0700)
	if err != nil {
		log.Panic("cannot create the user path " + usersPath)
	}

	return fs

}

func (fs *Storage) GetUser(uid string) (user *model.User, err error) {
	if uid == "" {
		err = errors.New("empty user")
		return
	}
	profilePath := fs.getPathFromUser(uid, profileName)
	_, err = os.Stat(profilePath)
	if err != nil {
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
		log.Error("Cannot read the user profile:", profilePath)
		return
	}

	user, err = model.DeserializeUser(content)
	if err != nil {
		log.Error("Cannot deserialize the user profile", profilePath)
		return
	}

	return
}

// GetUsers blah
func (fs *Storage) GetUsers() (users []*model.User, err error) {
	usersDir := fs.getUserPath("")

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
	if u.Id == "" {
		err = errors.New("empty id")
		return
	}
	userPath := fs.getUserPath(u.Id)

	// Create the user's directory
	err = os.MkdirAll(userPath, 0700)
	if err != nil {
		return
	}

	profilePath := fs.getPathFromUser(u.Id, profileName)
	// Create the profile file
	js, err := u.Serialize()
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
	if u.Id == "" {
		err = errors.New("empty id")
		return
	}

	err = os.MkdirAll(fs.getUserPath(u.Id), 0700)
	if err != nil {
		return
	}

	profilePath := fs.getPathFromUser(u.Id, profileName)
	// Overwrite the profile
	js, err := u.Serialize()
	if err != nil {
		return
	}
	err = ioutil.WriteFile(profilePath, js, 0600)

	return
}
