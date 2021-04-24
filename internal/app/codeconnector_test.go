package app

import (
	"testing"
)

func TestGenerateCode(t *testing.T) {
	u := NewCodeConnector()

	code, err := u.NewCode("test")

	if err != nil {
		t.Error(err)
	}

	uid, err := u.ConsumeCode(code)
	if err != nil {
		t.Error(err)
	}

	if uid != "test" {
		t.Fail()
	}

}
