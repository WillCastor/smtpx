// Simple smtp client to send email(support Chinese).
// Work well with qq, 163, mac mail app, google mail.
package smtpx

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"net/url"
	"strings"
	"sync"
)

type Attachment struct {
	Filename string
	Data     []byte
}

type Sender struct {
	Host     string
	Port     int
	Address  string
	Password string
	Name     string
}

type Letter struct {
	address     string
	name        string
	Subject     string
	content     string
	receivers   []string
	carbonCopy  []string
	attachments []Attachment
	body        []byte
	mutex       sync.Mutex
}

const boundary = "xxxxxxxx"

func NewSender(host string, port int, name, address, password string) *Sender {
	return &Sender{Host: host, Port: port, Address: address, Password: password, Name: name}
}

func NewLetter() *Letter {
	return &Letter{address: s.Address, name: s.Name}
}

func NewAttachment(filename string, data []byte) Attachment{
	return Attachment{Filename:filename, Data:data}
}

func (l *Letter) AddReceivers(addresses ...string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, address := range addresses {
		l.receivers = append(l.receivers, address)
	}
}

func (l *Letter) AddCarbonCopy(addresses ...string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, address := range addresses {
		l.carbonCopy = append(l.carbonCopy, address)
	}
}

func (l *Letter) SetSubject(subject string) {
	l.Subject = subject
}

func (l *Letter) SetContent(content string) {
	l.content = content
}

func (l *Letter) AddAttachments(attachments ...Attachment) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, item := range attachments {
		l.attachments = append(l.attachments, item)
	}
}

func (l *Letter) AddAttachment(filename string, data []byte) {
	l.AddAttachments(Attachment{Filename:filename, Data:data})
}

func (s *Sender) Send(l *Letter) error {
	l.build()
	auth := smtp.PlainAuth(
		"",
		s.Address,
		s.Password,
		s.Host,
	)
	return s.sendMailUsingTLS(
		fmt.Sprintf("%s:%d", s.Host, s.Port),
		auth,
		s.Address,
		l.receivers,
		l.carbonCopy,
		l.body)
}

func (l *Letter) Dump() {
	if len(l.body) == 0 {
		l.build()
	}
	fmt.Println(string(l.body))
}

func (l *Letter) build() {
	l.body = []byte{}
	l.buildHeader()
	l.buildContent()
	l.buildAttachments()
	l.buildEnd()
}

func (s *Sender) dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func (s *Sender) sendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, cc []string, msg []byte) (err error) {
	client, err := s.dial(addr)
	if err != nil {
		log.Println("Create smpt client error:", err)
		return err
	}
	defer client.Close()
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
				return err
			}
		}
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}
	for _, addr := range cc {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return client.Quit()
}

func (l *Letter) buildHeader() {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(
		fmt.Sprintf("From: %s<%s>\nTo: %s\nCC: %s\nSubject: %s\nMIME-Version: 1.0\n",
			l.name, l.address, strings.Join(l.receivers, ","), strings.Join(l.carbonCopy, ","), l.Subject))
	buffer.WriteString(
		fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n\n", boundary))
	buffer.WriteString(fmt.Sprintf("--%s\n", boundary))
	l.body = buffer.Bytes()
}

func (l *Letter) buildContent() {
	buffer := bytes.NewBuffer(l.body)
	buffer.WriteString("Content-Type: text/plain; charset=UTF-8\n")
	buffer.WriteString("Content-Transfer-Encoding: quoted-printable\n\n")
	buffer.WriteString(l.content)
	buffer.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
	l.body = buffer.Bytes()
}

func (l *Letter) buildAttachments() {
	buffer := bytes.NewBuffer(l.body)
	for i, item := range l.attachments {
		buffer.WriteString("Content-Type: application/octet-stream; charset=UTF-8;\n 	name=\"?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(item.Filename)) + "?=\"\n")
		buffer.WriteString("Content-Transfer-Encoding: base64\n")
		buffer.WriteString("Content-Disposition: attachment; filename*=utf-8''" + url.PathEscape(item.Filename) + "\n\n")
		encodeBuffer := make([]byte, base64.StdEncoding.EncodedLen(len(item.Data)))
		base64.StdEncoding.Encode(encodeBuffer, item.Data)
		buffer.Write(encodeBuffer)

		if i != len(l.attachments)-1 {
			buffer.WriteString(fmt.Sprintf("\n--%s\n", boundary))
		}
	}
	l.body = buffer.Bytes()
}

func (l *Letter) buildEnd() {
	buffer := bytes.NewBuffer(l.body)
	buffer.WriteString("\n--" + boundary + "--\n")
	l.body = buffer.Bytes()
}
