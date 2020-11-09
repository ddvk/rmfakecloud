package email

import (
	"io/ioutil"
	"net/mail"
	"testing"
)

func TestParseEmptyAddress(t *testing.T) {
	addreses := ", email@domain.com , "
	to, err := mail.ParseAddressList(addreses)
	if err != nil {
		t.Error(err)
	}
	if len(to) > 1 {
		t.Error("more than 2")
	}
	t.Log(to)
}

func TestRead(t *testing.T) {
	t.Skip("TODO: fake the sending")

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
	if err != nil && err.Error() != "not configured" {
		t.Error(err)
	}

}
