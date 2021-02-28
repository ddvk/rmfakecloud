package app

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"sync"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	log "github.com/sirupsen/logrus"
)

type inMemoryCodeConnector struct {
	dict         map[string]string
	uids         map[string]string
	lock         sync.Mutex
	codeValidity time.Duration
}

// NewCodeConnector constructor
func NewCodeConnector() common.CodeConnector {
	return &inMemoryCodeConnector{
		dict:         make(map[string]string),
		uids:         make(map[string]string),
		codeValidity: time.Minute * 5,
	}

}

func (conn *inMemoryCodeConnector) codeExpiry() {

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
		select {
		case <-time.After(conn.codeValidity):
			if _, err := conn.ConsumeCode(code); err == nil {
				log.Infof("removed unused code: %s for uid: %s ", code, uid)
			}
		}

	}()
	return code, nil
}

func newUserCode() (code string, err error) {
	b := make([]byte, 5)

	if _, err = rand.Read(b); err != nil {
		return
	}

	code = base32.StdEncoding.EncodeToString(b)

	return code, nil
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
