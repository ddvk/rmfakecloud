package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SSL/TLS Email Example
var servername, username, password, fromOverride, helo string

func init() {
	//TODO: remove env dependency
	servername = os.Getenv("RM_SMTP_SERVER")
	username = os.Getenv("RM_SMTP_USERNAME")
	password = os.Getenv("RM_SMTP_PASSWORD")
	if servername == "" {
		log.Warnln("smtp not configured, no emails will be sent")
	}
	helo = os.Getenv("RM_SMTP_HELO")
	fromOverride = os.Getenv("RM_SMTP_FROM")
}

type EmailBuilder struct {
	From    string
	To      string
	ReplyTo string
	Body    string
	Subject string

	attachments []Attachment
}
type Attachment struct {
	filename    string
	contentType string
	data        []byte
}

func sanitizeAttachmentName(name string) string {
	return filepath.Base(name)
}

// workaround for go < 1.15
func TrimAddresses(address string) string {
	return strings.Trim(strings.Trim(address, " "), ",")
}

func (b *EmailBuilder) AddFile(name string, data []byte, contentType string) {
	log.Debugln("Adding file: ", name, " contentType: ", contentType)
	if contentType == "" {
		log.Warnln("no contentType, setting to binary")
		contentType = "application/octet-stream"
	}
	attachment := Attachment{
		contentType: contentType,
		filename:    sanitizeAttachmentName(name),
		data:        data,
	}
	b.attachments = append(b.attachments, attachment)
}

func (b *EmailBuilder) Send() (err error) {
	if servername == "" {
		return fmt.Errorf("not configured")
	}
	frm := b.From
	if fromOverride != "" {
		frm = fromOverride
	}
	//if not defined
	from, err := mail.ParseAddress(frm)
	if err != nil {
		return err
	}
	to, err := mail.ParseAddressList(TrimAddresses(b.To))
	if err != nil {
		return err
	}

	log.Debug("from:", from)
	log.Debug("to:", to)

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

	if helo != "" {
		err = c.Hello(helo)
		if err != nil {
			return err
		}
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
	msg := fmt.Sprintf("From: %s\r\n", from)
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
	msg += b.Body

	log.Debug("mime msg", msg)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	//Add attachments
	for _, attachment := range b.attachments {
		log.Debugln("File attachment: ", attachment.filename)

		file := fmt.Sprintf("\r\n--%s\r\n", delimeter)
		file += "Content-Type: " + attachment.contentType + "; charset=\"utf-8\"\r\n"
		file += "Content-Transfer-Encoding: base64\r\n"
		file += "Content-Disposition: attachment;filename=\"" + attachment.filename + "\"\r\n\r\n"
		_, err = w.Write([]byte(file))
		if err != nil {
			return err
		}

		encoder := base64.NewEncoder(base64.StdEncoding, w)
		defer encoder.Close()
		_, err := encoder.Write(attachment.data)
		if err != nil {
			return err
		}
	}

	// Add last boundary delimeter, with trailing -- according to RFC 1341
	lastBoundary := fmt.Sprintf("\r\n--%s--\r\n", delimeter)
	_, err = w.Write([]byte(lastBoundary))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	log.Info("Message sent")
	return nil
}
