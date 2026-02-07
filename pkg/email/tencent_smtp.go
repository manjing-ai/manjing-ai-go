package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// SMTPConfig SMTP 配置
type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	UseSSL      bool
	UseStartTLS bool
	FromName    string
	FromAddr    string
}

// SMTPClient SMTP 客户端
type SMTPClient struct {
	cfg SMTPConfig
}

// NewSMTPClient 创建 SMTP 客户端
func NewSMTPClient(cfg SMTPConfig) *SMTPClient {
	return &SMTPClient{cfg: cfg}
}

// Send 发送邮件
func (c *SMTPClient) Send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	var client *smtp.Client
	var err error

	if c.cfg.UseSSL {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: c.cfg.Host})
		if err != nil {
			return err
		}
		client, err = smtp.NewClient(conn, c.cfg.Host)
		if err != nil {
			return err
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return err
		}
	}
	defer client.Quit()

	if c.cfg.UseStartTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: c.cfg.Host}); err != nil {
				return err
			}
		}
	}

	if c.cfg.Username != "" && c.cfg.Password != "" {
		auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	from := c.cfg.FromAddr
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	msg := buildMessage(c.cfg.FromName, c.cfg.FromAddr, to, subject, body)
	_, err = w.Write([]byte(msg))
	return err
}

func buildMessage(fromName, fromAddr, to, subject, body string) string {
	from := fromAddr
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, fromAddr)
	}
	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}
	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body
	return msg
}

var _ Client = (*SMTPClient)(nil)
