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

	cur, ok := u.CurrentCode("test")
	if !ok || cur != code {
		t.Fatalf("CurrentCode = %q,%v want %q,true", cur, ok, code)
	}

	uid, err := u.ConsumeCode(code)
	if err != nil {
		t.Error(err)
	}

	if uid != "test" {
		t.Fail()
	}

	if _, ok := u.CurrentCode("test"); ok {
		t.Fatal("CurrentCode should be empty after consume")
	}
}
