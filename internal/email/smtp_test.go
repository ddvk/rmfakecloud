package email

import (
	"bytes"
	"fmt"
	"net/mail"
	"os"
	"testing"
)

func TestParseEmptyAddress(t *testing.T) {
	addreses := TrimAddresses(", email@domain.com , blah@blah, ")
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
func TestSendMail(t *testing.T) {
	sendTo := os.Getenv("RM_SMTP_TO")
	if sendTo == "" {
		t.Skip("manual test")
	}

	cfg := &SMTPConfig{
		Server:   os.Getenv("RM_SMTP_SERVER"),
		Username: os.Getenv("RM_SMTP_USERNAME"),
		Password: os.Getenv("RM_SMTP_PASSWORD"),
	}

	file, _ := os.Open("test.txt")
	sender := Builder{
		To:      []*mail.Address{{Address: sendTo}},
		From:    &mail.Address{Address: "from@test.com"},
		Subject: "subj test 鬼 тест",
		Body: `<!DOCTYPE html>
		<html><body><h1>blah</h1></body></html>`,
	}
	sender.AddFile("test 鬼 тест ", file, "text/plain")

	err := sender.Send(cfg)
	if err != nil && err.Error() != "not configured" {
		t.Error(err)
	}
}
