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

type InMemConnector struct {
	dict         map[string]string
	lock         sync.Mutex
	codeValidity time.Duration
}

func NewCodeConnector() common.CodeConnector {
	return &InMemConnector{
		dict:         make(map[string]string),
		codeValidity: time.Second * 10,
	}

}

func (conn *InMemConnector) codeExpiry() {

}

func (conn *InMemConnector) NewCode(uid string) (string, error) {
	code, err := newUserCode()
	if err != nil {
		return "", err
	}
	conn.lock.Lock()
	conn.dict[code] = uid
	conn.lock.Unlock()
	go func() {
		select {
		case <-time.After(conn.codeValidity):
			conn.lock.Lock()
			delete(conn.dict, code)
			conn.lock.Unlock()
			log.Info("removed unused code: ", code)
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
func (conn *InMemConnector) ConsumeCode(code string) (string, error) {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	if uid, ok := conn.dict[code]; ok {
		delete(conn.dict, code)
		return uid, nil
	}
	return "", errors.New("code not found")
}
