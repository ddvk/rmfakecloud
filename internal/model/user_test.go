package model

import (
	"testing"
)

func TestGenerateCode(t *testing.T) {
	u := User{}
	code, err := u.NewUserCode()

	if err != nil {
		t.Error(err)
	}
	t.Log(code)

}
