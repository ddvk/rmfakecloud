package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// MaxLineLength is the maximum line length per RFC 2045
	MaxLineLength = 76
	delimeter     = "**=myohmy689407924327898338383"
)

// SMTPConfig smtp configuration
type SMTPConfig struct {
	Server       string
	Username     string
	Password     string
	FromOverride *mail.Address
	Helo         string
	InsecureTLS  bool
}

// Builder builds emails
type Builder struct {
	From    string
	To      string
	ReplyTo string
	Body    string
	Subject string

	attachments []emailAttachment
}

type emailAttachment struct {
	filename    string
	contentType string
	data        io.Reader
}

func sanitizeAttachmentName(name string) string {
	return filepath.Base(name)
}

// trimAddresses workaround for go < 1.15
func trimAddresses(address string) string {
	return strings.Trim(strings.Trim(address, " "), ",")
}

// AddFile adds a file attachment
func (b *Builder) AddFile(name string, data io.Reader, contentType string) {
	log.Debugln("Adding file: ", name, " contentType: ", contentType)
	if contentType == "" {
		log.Warnln("no contentType, setting to binary")
		contentType = "application/octet-stream"
	}
	attachment := emailAttachment{
		contentType: contentType,
		filename:    sanitizeAttachmentName(name),
		data:        data,
	}
	b.attachments = append(b.attachments, attachment)
}

// WriteAttachments streams the attachments
func (b *Builder) WriteAttachments(w io.Writer) (err error) {
	for _, attachment := range b.attachments {
		log.Debugln("File attachment: ", attachment.filename)

		fileHeader := fmt.Sprintf("\r\n--%s\r\n", delimeter)
		fileHeader += "Content-Type: " + attachment.contentType + "; charset=\"utf-8\"\r\n"
		fileHeader += "Content-Transfer-Encoding: base64\r\n"
		fileHeader += "Content-Disposition: attachment;filename*=utf-8''" + url.PathEscape(attachment.filename) + "\r\n\r\n"
		_, err = w.Write([]byte(fileHeader))
		if err != nil {
			return err
		}

		splittingEncoder := &SplittingWritter{
			innerWriter:    w,
			maxLineLength:  MaxLineLength,
			lineTerminator: "\r\n",
		}
		base64Encoder := base64.NewEncoder(base64.StdEncoding, splittingEncoder)
		_, err := io.Copy(base64Encoder, attachment.data)

		if err != nil {
			return err
		}
		base64Encoder.Close()
	}
	return nil
}

func utf8encode(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}

// Send sends the email
func (b *Builder) Send(cfg *SMTPConfig) (err error) {
	if cfg == nil {
		return fmt.Errorf("no smtp config")
	}
	var from *mail.Address
	if cfg.FromOverride != nil {
		from = cfg.FromOverride
	} else {
		from, err = mail.ParseAddress(b.From)
		if err != nil {
			log.Error("Invalid From address: ", b.From)
		}
	}
	//if not defined
	to, err := mail.ParseAddressList(trimAddresses(b.To))
	if err != nil {
		return err
	}

	log.Debug("from:", from)
	log.Debug("to:", to)

	host, _, _ := net.SplitHostPort(cfg.Server)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureTLS,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", cfg.Server, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if cfg.Helo != "" {
		err = c.Hello(cfg.Helo)
		if err != nil {
			return err
		}
	}

	if cfg.Username != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, host)
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
	//basic email headers
	msg := fmt.Sprintf("From: %s\r\n", utf8encode(from.String()))
	msg += fmt.Sprintf("To: %s\r\n", utf8encode(b.To))
	msg += fmt.Sprintf("Subject: %s\r\n", utf8encode(b.Subject))
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

	err = b.WriteAttachments(w)
	if err != nil {
		return err
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

// SplittingWritter writes a stream and inserts a terminator
type SplittingWritter struct {
	innerWriter       io.Writer
	currentLineLength int
	maxLineLength     int
	lineTerminator    string
}

func (w *SplittingWritter) Write(p []byte) (n int, err error) {
	length := len(p)
	total := 0
	for to, from := 0, 0; from < length; from = to {
		delta := w.maxLineLength - w.currentLineLength

		to = from + delta
		if to > length {
			to = length
			delta = length - from
		}

		n, err = w.innerWriter.Write(p[from:to])
		total += n
		if err != nil {
			return total, err
		}

		w.currentLineLength += delta

		if w.currentLineLength == w.maxLineLength {
			n, err = w.innerWriter.Write([]byte(w.lineTerminator))
			total += n
			if err != nil {
				return total, err
			}
			w.currentLineLength = 0
		}
	}

	return total, nil
}
