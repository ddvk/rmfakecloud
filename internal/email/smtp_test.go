package email

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	file, _ := ioutil.ReadFile("test.txt")
	sender := EmailBuilder{
		To:      "bingobango@mailinator.com",
		From:    "bingo.bongo@gmail.com",
		Subject: "testing",
		Body: `<!DOCTYPE html>
		<html><body><h1>blah</h1></body></html>`,
		// FileName: []string{"sometest.txt"},
		// File:     files,
	}
	sender.AddFile("tst", file)

	err := sender.Send()
	assert.NoError(t, err)

}
