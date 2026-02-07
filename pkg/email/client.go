package email

import "context"

// Client 邮件发送客户端
type Client interface {
	Send(ctx context.Context, to, subject, body string) error
}
