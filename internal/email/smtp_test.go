package email

import (
	"bytes"
	"fmt"
	"net/mail"
	"os"
	"testing"
)

func TestParseEmptyAddress(t *testing.T) {
	addreses := trimAddresses(", email@domain.com , blah@blah, ")
	to, err := mail.ParseAddressList(addreses)
	if err != nil {
		t.Error(err)
	}
	if len(to) > 2 {
		t.Error("more than 2")
	}
	t.Log(to)
}

func TestAttachments(t *testing.T) {

	buf := bytes.Buffer{}
	splittingEncoder := &SplittingWritter{
		innerWriter:    &buf,
		maxLineLength:  1,
		lineTerminator: "!\n",
	}
	_, err := fmt.Fprintf(splittingEncoder, "123456")
	if err != nil {
		t.Error(err)
	}
	expected := `1!
2!
3!
4!
5!
6!
`
	result := buf.String()
	if result != expected {
		t.Error("doesn't match", result)
	}
}
func TestRead(t *testing.T) {
	t.Skip("TODO: fake the sending")

	file, _ := os.Open("test.txt")
	sender := Builder{
		To:      "bingobango@mailinator.com",
		From:    "bingo.bongo@gmail.com",
		Subject: "testing",
		Body: `<!DOCTYPE html>
		<html><body><h1>blah</h1></body></html>`,
		// FileName: []string{"sometest.txt"},
		// File:     files,
	}
	sender.AddFile("tst", file, "text/plain")

	err := sender.Send(nil)
	if err != nil && err.Error() != "not configured" {
		t.Error(err)
	}

}
