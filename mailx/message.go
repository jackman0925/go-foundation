package mailx

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"strings"
	"time"
)

const crlf = "\r\n"

// Message 表示一封待发送邮件。
type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Text        string
	HTML        string
	Attachments []Attachment
}

// Attachment 表示邮件附件。
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Sender 定义邮件发送器接口，便于业务项目替换实现或单元测试 mock。
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// ValidateMessage 校验邮件基础字段和地址格式。
func ValidateMessage(msg Message) error {
	if _, err := parseSingleAddress("from", msg.From); err != nil {
		return err
	}
	if len(msg.To) == 0 {
		return errors.New("mail to is required")
	}
	if err := validateAddressList("to", msg.To); err != nil {
		return err
	}
	if err := validateAddressList("cc", msg.Cc); err != nil {
		return err
	}
	if err := validateAddressList("bcc", msg.Bcc); err != nil {
		return err
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return errors.New("mail subject is required")
	}
	if containsLineBreak(msg.Subject) {
		return errors.New("mail subject contains invalid line break")
	}
	if strings.TrimSpace(msg.Text) == "" && strings.TrimSpace(msg.HTML) == "" && len(msg.Attachments) == 0 {
		return errors.New("mail body or attachment is required")
	}
	for i, attachment := range msg.Attachments {
		if strings.TrimSpace(attachment.Filename) == "" {
			return fmt.Errorf("mail attachment %d filename is required", i)
		}
		if containsLineBreak(attachment.Filename) {
			return fmt.Errorf("mail attachment %d filename contains invalid line break", i)
		}
		if len(attachment.Data) == 0 {
			return fmt.Errorf("mail attachment %d data is required", i)
		}
	}
	return nil
}

// Recipients 返回 SMTP envelope 使用的去重收件人列表。
func (m Message) Recipients() ([]string, error) {
	groups := [][]string{m.To, m.Cc, m.Bcc}
	seen := make(map[string]struct{})
	recipients := make([]string, 0, len(m.To)+len(m.Cc)+len(m.Bcc))

	for _, group := range groups {
		for _, raw := range group {
			address, err := parseSingleAddress("recipient", raw)
			if err != nil {
				return nil, err
			}
			key := strings.ToLower(address.Address)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			recipients = append(recipients, address.Address)
		}
	}

	if len(recipients) == 0 {
		return nil, errors.New("mail recipient is required")
	}
	return recipients, nil
}

func buildMIMEMessage(msg Message) ([]byte, error) {
	if err := ValidateMessage(msg); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	writeHeader(&buffer, "From", mustAddressString(msg.From))
	writeHeader(&buffer, "To", mustAddressListString(msg.To))
	if len(msg.Cc) > 0 {
		writeHeader(&buffer, "Cc", mustAddressListString(msg.Cc))
	}
	writeHeader(&buffer, "Subject", mime.QEncoding.Encode("utf-8", msg.Subject))
	writeHeader(&buffer, "Date", time.Now().Format(time.RFC1123Z))
	writeHeader(&buffer, "MIME-Version", "1.0")

	if len(msg.Attachments) == 0 {
		writeBodyPart(&buffer, msg.Text, msg.HTML)
		return buffer.Bytes(), nil
	}

	mixedBoundary := makeBoundary("mixed", msg.Subject)
	writeHeader(&buffer, "Content-Type", fmt.Sprintf(`multipart/mixed; boundary="%s"`, mixedBoundary))
	buffer.WriteString(crlf)

	buffer.WriteString("--" + mixedBoundary + crlf)
	writeBodyPart(&buffer, msg.Text, msg.HTML)

	for _, attachment := range msg.Attachments {
		buffer.WriteString(crlf)
		buffer.WriteString("--" + mixedBoundary + crlf)
		writeAttachment(&buffer, attachment)
	}
	buffer.WriteString(crlf)
	buffer.WriteString("--" + mixedBoundary + "--" + crlf)

	return buffer.Bytes(), nil
}

func writeBodyPart(buffer *bytes.Buffer, text string, html string) {
	if strings.TrimSpace(text) != "" && strings.TrimSpace(html) != "" {
		boundary := makeBoundary("alternative", text+html)
		writeHeader(buffer, "Content-Type", fmt.Sprintf(`multipart/alternative; boundary="%s"`, boundary))
		buffer.WriteString(crlf)
		buffer.WriteString("--" + boundary + crlf)
		writeTextPart(buffer, "text/plain", text)
		buffer.WriteString(crlf)
		buffer.WriteString("--" + boundary + crlf)
		writeTextPart(buffer, "text/html", html)
		buffer.WriteString(crlf)
		buffer.WriteString("--" + boundary + "--" + crlf)
		return
	}
	if strings.TrimSpace(html) != "" {
		writeTextPart(buffer, "text/html", html)
		return
	}
	writeTextPart(buffer, "text/plain", text)
}

func writeTextPart(buffer *bytes.Buffer, contentType string, body string) {
	writeHeader(buffer, "Content-Type", contentType+"; charset=utf-8")
	writeHeader(buffer, "Content-Transfer-Encoding", "base64")
	buffer.WriteString(crlf)
	writeBase64(buffer, []byte(body))
}

func writeAttachment(buffer *bytes.Buffer, attachment Attachment) {
	contentType := strings.TrimSpace(attachment.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	encodedFilename := mime.QEncoding.Encode("utf-8", attachment.Filename)

	writeHeader(buffer, "Content-Type", fmt.Sprintf(`%s; name="%s"`, contentType, encodedFilename))
	writeHeader(buffer, "Content-Transfer-Encoding", "base64")
	writeHeader(buffer, "Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, encodedFilename))
	buffer.WriteString(crlf)
	writeBase64(buffer, attachment.Data)
}

func writeBase64(buffer *bytes.Buffer, data []byte) {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	for len(encoded) > 76 {
		buffer.Write(encoded[:76])
		buffer.WriteString(crlf)
		encoded = encoded[76:]
	}
	buffer.Write(encoded)
	buffer.WriteString(crlf)
}

func writeHeader(buffer *bytes.Buffer, key string, value string) {
	buffer.WriteString(key)
	buffer.WriteString(": ")
	buffer.WriteString(value)
	buffer.WriteString(crlf)
}

func validateAddressList(field string, values []string) error {
	for _, value := range values {
		if _, err := parseSingleAddress(field, value); err != nil {
			return err
		}
	}
	return nil
}

func parseSingleAddress(field string, value string) (*mail.Address, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("mail %s is required", field)
	}
	if containsLineBreak(value) {
		return nil, fmt.Errorf("mail %s contains invalid line break", field)
	}
	address, err := mail.ParseAddress(value)
	if err != nil {
		return nil, fmt.Errorf("mail %s is invalid: %w", field, err)
	}
	if !strings.Contains(address.Address, "@") {
		return nil, fmt.Errorf("mail %s is invalid", field)
	}
	return address, nil
}

func mustAddressString(value string) string {
	address, _ := mail.ParseAddress(value)
	return address.String()
}

func mustAddressListString(values []string) string {
	addresses := make([]string, 0, len(values))
	for _, value := range values {
		address, _ := mail.ParseAddress(value)
		addresses = append(addresses, address.String())
	}
	return strings.Join(addresses, ", ")
}

func containsLineBreak(value string) bool {
	return strings.ContainsAny(value, "\r\n")
}

func makeBoundary(prefix string, seed string) string {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(seed))
	if len(encoded) > 24 {
		encoded = encoded[:24]
	}
	return "go-foundation-" + prefix + "-" + encoded
}
