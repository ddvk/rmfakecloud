package email

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"

	"github.com/zgs225/rmfakecloud/internal/common"
	log "github.com/sirupsen/logrus"
)

const (
	// MaxLineLength is the maximum line length per RFC 2045
	MaxLineLength = 76
	delimeter     = "**=myohmy689407924327898338383"
	smtpLog       = "[smtp] "
)

// SMTPConfig smtp configuration
type SMTPConfig struct {
	Server       string
	Username     string
	Password     string
	FromOverride *mail.Address
	Helo         string
	InsecureTLS  bool
	NoTLS        bool
	StartTLS     bool
}

// Builder builds emails
type Builder struct {
	From    *mail.Address
	To      []*mail.Address
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

// TrimAddresses workaround for go < 1.15
func TrimAddresses(address string) string {
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
		filename:    common.Sanitize(name),
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
		return errors.New("no smtp config")
	}

	host, _, err := net.SplitHostPort(cfg.Server)
	if err != nil {
		return err
	}

	var conn net.Conn

	tlsconfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureTLS,
		ServerName:         host,
	}

	if cfg.NoTLS {
		conn, err = net.Dial("tcp", cfg.Server)
	} else {
		conn, err = tls.Dial("tcp", cfg.Server, tlsconfig)
	}

	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if cfg.StartTLS {
		err = c.StartTLS(tlsconfig)
		if err != nil {
			return err
		}
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

	if err = c.Mail(b.From.Address); err != nil {
		return err
	}

	for _, addr := range b.To {
		if err = c.Rcpt(addr.Address); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	toList := make([]string, 0)
	for _, toAddr := range b.To {
		toList = append(toList, toAddr.String())
	}
	to := strings.Join(toList, ", ")

	msgBuilder := strings.Builder{}
	//basic email headers
	msgBuilder.WriteString(fmt.Sprintf("From: %s\r\n", utf8encode(b.From.String())))
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", utf8encode(to)))
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", utf8encode(b.Subject)))
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter))
	msgBuilder.WriteString(fmt.Sprintf("\r\n--%s\r\n", delimeter))
	msgBuilder.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msgBuilder.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	msgBuilder.WriteString("Content-Disposition: inline\r\n")
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(b.Body)

	msg := msgBuilder.String()

	log.Debug("mime msg:\n", msg)

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

	err = c.Quit()
	return err
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
