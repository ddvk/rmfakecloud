package app

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type inMemoryCodeConnector struct {
	dict         map[string]string
	uids         map[string]string
	lock         sync.Mutex
	codeValidity time.Duration
}

// CodeConnector matches a code to users
type CodeConnector interface {
	//NewCode generates one time code for a user
	NewCode(uid string) (code string, err error)

	//CurrentCode returns the outstanding pairing code for a user, if any.
	CurrentCode(uid string) (code string, ok bool)

	//ConsumeCode a code and returns the uid if ofound
	ConsumeCode(code string) (uid string, err error)
}

// NewCodeConnector constructor
func NewCodeConnector() CodeConnector {
	return &inMemoryCodeConnector{
		dict:         make(map[string]string),
		uids:         make(map[string]string),
		codeValidity: time.Minute * 5,
	}

}

func (conn *inMemoryCodeConnector) NewCode(uid string) (string, error) {
	code, err := newUserCode()
	if err != nil {
		return "", err
	}
	conn.lock.Lock()
	conn.dict[code] = uid
	if oldcode, ok := conn.uids[uid]; ok {
		delete(conn.dict, oldcode)
	}
	conn.uids[uid] = code
	conn.lock.Unlock()
	go func() {
		<-time.After(conn.codeValidity)
		conn.lock.Lock()
		defer conn.lock.Unlock()
		if cur, ok := conn.uids[uid]; ok && cur == code {
			delete(conn.dict, code)
			delete(conn.uids, uid)
			log.Infof("removed unused code: %s for uid: %s ", code, uid)
		}
	}()
	return code, nil
}

func (conn *inMemoryCodeConnector) CurrentCode(uid string) (string, bool) {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	code, ok := conn.uids[uid]
	return code, ok
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

// ConsumeCode return the userId matching the
func (conn *inMemoryCodeConnector) ConsumeCode(code string) (string, error) {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	if uid, ok := conn.dict[code]; ok {
		delete(conn.dict, code)
		delete(conn.uids, uid)
		return uid, nil
	}
	return "", errors.New("code not found")
}
