package mailx

import (
	"strings"
	"testing"
)

func TestValidateMessageRequiresCoreFields(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
	}{
		{
			name: "missing from",
			msg: Message{
				To:      []string{"to@example.com"},
				Subject: "hello",
				Text:    "content",
			},
		},
		{
			name: "missing to",
			msg: Message{
				From:    "from@example.com",
				Subject: "hello",
				Text:    "content",
			},
		},
		{
			name: "missing subject",
			msg: Message{
				From: "from@example.com",
				To:   []string{"to@example.com"},
				Text: "content",
			},
		},
		{
			name: "missing body",
			msg: Message{
				From:    "from@example.com",
				To:      []string{"to@example.com"},
				Subject: "hello",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateMessage(tt.msg); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateMessageRejectsInvalidAddress(t *testing.T) {
	msg := Message{
		From:    "from@example.com",
		To:      []string{"not-an-email"},
		Subject: "hello",
		Text:    "content",
	}

	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected invalid address error")
	}
}

func TestRecipientsReturnsDedupedEnvelopeRecipients(t *testing.T) {
	msg := Message{
		From:    "from@example.com",
		To:      []string{"A@example.com", "b@example.com"},
		Cc:      []string{"a@example.com"},
		Bcc:     []string{"hidden@example.com", "B@example.com"},
		Subject: "hello",
		Text:    "content",
	}

	recipients, err := msg.Recipients()
	if err != nil {
		t.Fatalf("Recipients returned error: %v", err)
	}

	want := []string{"A@example.com", "b@example.com", "hidden@example.com"}
	if len(recipients) != len(want) {
		t.Fatalf("expected %d recipients, got %d: %v", len(want), len(recipients), recipients)
	}
	for i := range want {
		if recipients[i] != want[i] {
			t.Fatalf("recipient %d mismatch: want %s got %s", i, want[i], recipients[i])
		}
	}
}

func TestBuildMIMEMessageIncludesToAndCcButNotBccHeader(t *testing.T) {
	msg := Message{
		From:    "from@example.com",
		To:      []string{"to1@example.com", "to2@example.com"},
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"hidden@example.com"},
		Subject: "测试邮件",
		Text:    "纯文本内容",
	}

	data, err := buildMIMEMessage(msg)
	if err != nil {
		t.Fatalf("buildMIMEMessage returned error: %v", err)
	}
	text := string(data)

	if !strings.Contains(text, "To: <to1@example.com>, <to2@example.com>") {
		t.Fatalf("expected To header, got:\n%s", text)
	}
	if !strings.Contains(text, "Cc: <cc@example.com>") {
		t.Fatalf("expected Cc header, got:\n%s", text)
	}
	if strings.Contains(text, "Bcc:") || strings.Contains(text, "hidden@example.com") {
		t.Fatalf("Bcc must not be included in MIME headers, got:\n%s", text)
	}
	if !strings.Contains(text, "Subject: =?utf-8?q?") {
		t.Fatalf("expected encoded subject, got:\n%s", text)
	}
}

func TestBuildMIMEMessageSupportsHTMLAndAttachments(t *testing.T) {
	msg := Message{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "hello",
		Text:    "text body",
		HTML:    "<strong>html body</strong>",
		Attachments: []Attachment{
			{
				Filename:    "report.txt",
				ContentType: "text/plain",
				Data:        []byte("attachment content"),
			},
		},
	}

	data, err := buildMIMEMessage(msg)
	if err != nil {
		t.Fatalf("buildMIMEMessage returned error: %v", err)
	}
	text := string(data)

	for _, want := range []string{
		"multipart/mixed",
		"multipart/alternative",
		"text/plain; charset=utf-8",
		"text/html; charset=utf-8",
		`filename="report.txt"`,
		"YXR0YWNobWVudCBjb250ZW50",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected MIME message to contain %q, got:\n%s", want, text)
		}
	}
}

func TestBuildMIMEMessageSupportsHTMLOnly(t *testing.T) {
	msg := Message{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "hello",
		HTML:    "<strong>html body</strong>",
	}

	data, err := buildMIMEMessage(msg)
	if err != nil {
		t.Fatalf("buildMIMEMessage returned error: %v", err)
	}
	text := string(data)

	if !strings.Contains(text, "Content-Type: text/html; charset=utf-8") {
		t.Fatalf("expected html content type, got:\n%s", text)
	}
	if strings.Contains(text, "multipart/alternative") {
		t.Fatalf("did not expect multipart alternative, got:\n%s", text)
	}
}

func TestBuildMIMEMessageUsesDefaultAttachmentContentType(t *testing.T) {
	msg := Message{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "hello",
		Text:    "content",
		Attachments: []Attachment{
			{
				Filename: "data.bin",
				Data:     []byte(strings.Repeat("a", 80)),
			},
		},
	}

	data, err := buildMIMEMessage(msg)
	if err != nil {
		t.Fatalf("buildMIMEMessage returned error: %v", err)
	}
	text := string(data)

	if !strings.Contains(text, "application/octet-stream") {
		t.Fatalf("expected default attachment content type, got:\n%s", text)
	}
	if !strings.Contains(text, "\r\nYWFh") {
		t.Fatalf("expected wrapped base64 attachment data, got:\n%s", text)
	}
}

func TestValidateMessageRejectsInvalidAttachments(t *testing.T) {
	tests := []struct {
		name       string
		attachment Attachment
	}{
		{
			name: "missing filename",
			attachment: Attachment{
				Data: []byte("content"),
			},
		},
		{
			name: "filename injection",
			attachment: Attachment{
				Filename: "report.txt\r\nX-Test: yes",
				Data:     []byte("content"),
			},
		},
		{
			name: "missing data",
			attachment: Attachment{
				Filename: "report.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := validMessage()
			msg.Attachments = []Attachment{tt.attachment}
			if err := ValidateMessage(msg); err == nil {
				t.Fatal("expected attachment validation error")
			}
		})
	}
}

func TestRecipientsRejectsEmptyRecipient(t *testing.T) {
	msg := validMessage()
	msg.Cc = []string{""}

	if _, err := msg.Recipients(); err == nil {
		t.Fatal("expected recipient validation error")
	}
}

func TestValidateMessageRejectsHeaderInjection(t *testing.T) {
	msg := Message{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "hello\r\nBcc: hidden@example.com",
		Text:    "content",
	}

	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected header injection error")
	}
}
