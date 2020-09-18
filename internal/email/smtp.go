package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SSL/TLS Email Example
var servername, username, password, fromOverride string

func init() {
	//TODO: remove env dependency
	servername = os.Getenv("RM_SMTP_SERVER")
	username = os.Getenv("RM_SMTP_USERNAME")
	password = os.Getenv("RM_SMTP_PASSWORD")
	if servername == "" {
		log.Warnln("smtp not configured, no emails will be sent")
	}
	fromOverride = os.Getenv("RM_STMTP_FROM")
}

type EmailBuilder struct {
	From      string
	To        string
	ReplyTo   string
	Body      string
	Subject   string
	fileNames []string
	files     [][]byte
}

/// remove remarkable ads
func Strip(msg string) string {
	br := "<br>--<br>"
	i := strings.Index(msg, br)
	if i > 0 {
		return msg[:i]
	}
	return msg
}

func (b *EmailBuilder) AddFile(name string, data []byte) {
	if b.fileNames == nil || b.files == nil {
		b.fileNames = []string{name}
		b.files = [][]byte{data}
		return
	}
	b.fileNames = append(b.fileNames, name)
	b.files = append(b.files, data)
}

func (b *EmailBuilder) Send() (err error) {
	if servername == "" {
		return fmt.Errorf("not configured")
	}
	log.Println("smtp client")
	frm := b.From
	if fromOverride != "" {
		frm = fromOverride
	}
	//if not defined
	from, err := mail.ParseAddress(frm)
	if err != nil {
		return err
	}
	to, err := mail.ParseAddressList(b.To)
	if err != nil {
		return err
	}

	log.Println("from:", from)
	log.Println("to:", to)

	host, _, _ := net.SplitHostPort(servername)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if username != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	if err = c.Mail(from.Address); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr.Address); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	delimeter := "**=myohmy689407924327898338383"
	//basic email headers
	msg := fmt.Sprintf("From: %s\r\n", b.From)
	msg += fmt.Sprintf("To: %s\r\n", b.To)
	msg += fmt.Sprintf("Subject: %s\r\n", b.Subject)
	// msg += fmt.Sprintf("ReplyTo: %s\r\n", b.ReplyTo)

	msg += "MIME-Version: 1.0\r\n"
	msg += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter)

	msg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	msg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	msg += "Content-Transfer-Encoding: quoted-printable\r\n"
	msg += "Content-Disposition: inline\r\n"
	msg += "\r\n"
	msg += Strip(b.Body)

	log.Println("mime msg", msg)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	//Add attachments
	for i, f := range b.fileNames {
		log.Printf("File attachment: %s\n", f)

		file := fmt.Sprintf("\r\n--%s\r\n", delimeter)
		file += "Content-Type: text/plain; charset=\"utf-8\"\r\n"
		file += "Content-Transfer-Encoding: base64\r\n"
		file += "Content-Disposition: attachment;filename=\"" + f + "\"\r\n\r\n"
		_, err = w.Write([]byte(file))
		if err != nil {
			return err
		}

		encoder := base64.NewEncoder(base64.StdEncoding, w)
		defer encoder.Close()
		_, err := encoder.Write(b.files[i])
		if err != nil {
			return err
		}
	}
	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	return nil
}
