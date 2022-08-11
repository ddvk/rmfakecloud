package fs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zgs225/rmfakecloud/internal/config"
	"github.com/zgs225/rmfakecloud/internal/model"
	log "github.com/sirupsen/logrus"
)

const (
	userDir     = "users"
	profileName = ".userprofile"
)

// NewStorage new file system storage
func NewStorage(cfg *config.Config) *FileSystemStorage {
	fs := &FileSystemStorage{
		Cfg: cfg,
	}

	usersPath := fs.getUserPath("")
	err := os.MkdirAll(usersPath, 0700)
	if err != nil {
		log.Fatal("cannot create the user path " + usersPath)
	}

	return fs
}

// GetUser retrieves a user from the storage
func (fs *FileSystemStorage) GetUser(uid string) (user *model.User, err error) {
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
		return nil, fmt.Errorf("cannot read profile: %s, %w", profilePath, err)
	}

	user, err = model.DeserializeUser(content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse profile: %s, %w", profilePath, err)
	}

	return
}

// GetUsers gets all users
func (fs *FileSystemStorage) GetUsers() (users []*model.User, err error) {
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
func (fs *FileSystemStorage) RegisterUser(u *model.User) (err error) {
	if u.ID == "" {
		err = errors.New("empty id")
		return
	}
	userBlobPath := fs.getUserBlobPath(u.ID)

	// Create the user's directory
	err = os.MkdirAll(userBlobPath, 0700)
	if err != nil {
		return
	}

	profilePath := fs.getPathFromUser(u.ID, profileName)
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

// UpdateUser updates the user
func (fs *FileSystemStorage) UpdateUser(u *model.User) (err error) {
	if u.ID == "" {
		err = errors.New("empty id")
		return
	}

	userSyncPath := fs.getUserBlobPath(u.ID)
	err = os.MkdirAll(userSyncPath, 0700)
	if err != nil {
		return
	}

	profilePath := fs.getPathFromUser(u.ID, profileName)
	// Overwrite the profile
	js, err := u.Serialize()
	if err != nil {
		return
	}
	err = ioutil.WriteFile(profilePath, js, 0600)

	return
}

// RemoveUser remove the user and their data
func (fs *FileSystemStorage) RemoveUser(uid string) (err error) {
	if uid == "" {
		err = errors.New("empty id")
		return
	}

	userSyncPath := fs.getUserPath(uid)
	err = os.RemoveAll(userSyncPath)
	if err != nil {
		return
	}

	return
}
