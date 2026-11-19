package app

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type codeEntry struct {
	uid       string
	expiresAt time.Time
}

type inMemoryCodeConnector struct {
	dict         map[string]codeEntry
	uids         map[string]string
	lock         sync.Mutex
	codeValidity time.Duration
}

// CodeConnector matches a code to users
type CodeConnector interface {
	// NewCode generates one time code for a user
	NewCode(uid string) (code string, err error)
	// ConsumeCode consumes a code and returns the uid if found
	ConsumeCode(code string) (uid string, err error)
	// CodeStatus returns expiration and whether the user's code is still valid
	CodeStatus(uid string) (expiresAt time.Time, valid bool)
}

// NewCodeConnector constructor
func NewCodeConnector() CodeConnector {
	return &inMemoryCodeConnector{
		dict:         make(map[string]codeEntry),
		uids:         make(map[string]string),
		codeValidity: time.Minute * 5,
	}
}

func (conn *inMemoryCodeConnector) NewCode(uid string) (string, error) {
	code, err := newUserCode()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(conn.codeValidity)
	conn.lock.Lock()
	if oldcode, ok := conn.uids[uid]; ok {
		delete(conn.dict, oldcode)
	}
	conn.dict[code] = codeEntry{uid: uid, expiresAt: expiresAt}
	conn.uids[uid] = code
	conn.lock.Unlock()
	go func() {
		<-time.After(conn.codeValidity)
		if _, err := conn.ConsumeCode(code); err == nil {
			log.Infof("removed unused code: %s for uid: %s ", code, uid)
		}
	}()
	return code, nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) (string, error) {
	b := make([]rune, n)
	for i := range b {
		ri, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[int(ri.Int64())]
	}
	return string(b), nil
}
func newUserCode() (code string, err error) {
	return randSeq(8)
	// b := make([]byte, 5)

	// if _, err = rand.Read(b); err != nil {
	// 	return
	// }

	// code = base32.StdEncoding.EncodeToString(b)

	// return code, nil
}

// ConsumeCode returns the userId matching the code
func (conn *inMemoryCodeConnector) ConsumeCode(code string) (string, error) {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	if ent, ok := conn.dict[code]; ok {
		delete(conn.dict, code)
		delete(conn.uids, ent.uid)
		return ent.uid, nil
	}
	return "", errors.New("code not found")
}

// CodeStatus returns the expiration time and whether the user's code is still valid
func (conn *inMemoryCodeConnector) CodeStatus(uid string) (expiresAt time.Time, valid bool) {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	code, ok := conn.uids[uid]
	if !ok {
		return time.Time{}, false
	}
	ent, ok := conn.dict[code]
	if !ok {
		return time.Time{}, false
	}
	return ent.expiresAt, true
}
