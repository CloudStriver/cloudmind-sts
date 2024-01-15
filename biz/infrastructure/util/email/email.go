package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util"
	"github.com/CloudStriver/go-pkg/utils/pconvertor"
	"github.com/zeromicro/go-zero/core/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

const (
	contentType = "text/html; charset=UTF-8"
	body        = "<body><div class=\"container\"><p>你好，</p><p>你此次重置密码的验证码如下，请在 5 分钟内输入验证码进行下一步操作。如非你本人操作，请忽略此邮件。</p><p><strong>验证码：</strong>{{.code}}</p></div></body><style>body{font-family:Arial,sans-serif;background-color:#f0f0f0;margin:0;padding:0;}.container{max-width:600px;margin:0 auto;padding:20px;background-color:#ffffff;border-radius:5px;box-shadow:0 0 10px rgba(0,0,0,.1);}p{font-size:16px;line-height:1.6;color:#333333;}strong{font-weight:bold;}</style>\n"
)

func SendEmail(ctx context.Context, EmailConf config.EmailConf, toEmail, subject string) (string, error) {
	_, span := trace.TracerFromContext(ctx).Start(ctx, "auth/SendEmail", oteltrace.WithTimestamp(time.Now()), oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer func() {
		span.End(oteltrace.WithTimestamp(time.Now()))
	}()
	header := make(map[string]string)
	header["From"] = "CloudMind " + "<" + EmailConf.Email + ">"
	header["To"] = toEmail
	header["Subject"] = subject
	header["Content-Type"] = contentType

	Code := util.GenerateCode()
	message := buildMessage(header, strings.Replace(body, "{{.code}}", Code, 1))

	auth := smtp.PlainAuth("", EmailConf.Email, EmailConf.Password, EmailConf.Host)

	return Code, SendMailWithTLS(fmt.Sprintf("%s:%d", EmailConf.Host, EmailConf.Port), auth, EmailConf.Email, []string{toEmail}, pconvertor.String2Bytes(message))
}

func buildMessage(header map[string]string, body string) string {
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	return message
}

func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("tls.Dial Error:", err)
		return nil, err
	}

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func SendMailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	c, err := Dial(addr)
	if err != nil {
		log.Println("Create smtp client error:", err)
		return err
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
				return err
			}
		}
	}

	if err := c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return c.Quit()
}
