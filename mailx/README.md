# mailx

`mailx` 提供基础邮件发送能力，第一版聚焦 SMTP 同步发送和并发控制。

## 导入

```go
import "github.com/jackman0925/go-foundation/mailx"
```

## 推荐用法

```go
sender, err := mailx.NewSMTPClient(mailx.SMTPClientOptions{
    SMTP: mailx.SMTPOptions{
        Host:       cfg.Mail.Host,
        Port:       cfg.Mail.Port,
        Username:   cfg.Mail.Username,
        Password:   cfg.Mail.Password,
        From:       cfg.Mail.From,
        Encryption: mailx.EncryptionSTARTTLS,
        Timeout:    10 * time.Second,
    },
    LimitOptions: mailx.LimitOptions{
        MaxConcurrent: 5,
        MaxWaiting:    100,
    },
})
if err != nil {
    return err
}

err = sender.Send(ctx, mailx.Message{
    To:      []string{"user@example.com"},
    Cc:      []string{"manager@example.com"},
    Bcc:     []string{"audit@example.com"},
    Subject: "通知",
    Text:    "这是一封文本邮件",
    HTML:    "<p>这是一封 HTML 邮件</p>",
})
```

`NewSMTPClient` 默认会包装 `LimitedSender`。如果不传限制配置，默认同时最多发送 5 封，最多 100 个调用等待。

## SMTP 裸发送器

明确由业务系统或 worker 自己控制并发时，可以直接使用 `NewSMTPSender`。

```go
sender, err := mailx.NewSMTPSender(mailx.SMTPOptions{
    Host:       cfg.Mail.Host,
    Port:       cfg.Mail.Port,
    Username:   cfg.Mail.Username,
    Password:   cfg.Mail.Password,
    From:       cfg.Mail.From,
    Encryption: mailx.EncryptionSTARTTLS,
    Timeout:    10 * time.Second,
})
```

## 并发限制

频繁发送邮件时，建议使用 `LimitedSender` 控制同时发送数量和等待数量，避免业务系统把 SMTP 服务打满，也避免无限阻塞业务 goroutine。

```go
limited, err := mailx.NewLimitedSender(sender, mailx.LimitOptions{
    MaxConcurrent: 5,
    MaxWaiting:    100,
})
if err != nil {
    return err
}

err = limited.Send(ctx, msg)
if errors.Is(err, mailx.ErrMailBusy) {
    // 邮件没有发送，也没有进入队列，业务系统应写任务表、稍后重试或降级处理。
}
```

## 附件

```go
err = sender.Send(ctx, mailx.Message{
    To:      []string{"user@example.com"},
    Subject: "报表",
    Text:    "请查看附件",
    Attachments: []mailx.Attachment{
        {
            Filename:    "report.csv",
            ContentType: "text/csv",
            Data:        data,
        },
    },
})
```

## 注意事项

- `To` 必须至少有一个收件人；
- `Cc` 和 `Bcc` 可选；
- SMTP envelope 会包含 `To`、`Cc`、`Bcc`，邮件 MIME 头只包含 `To` 和 `Cc`；
- `Bcc` 不会写入邮件头，避免密送地址泄露；
- `LimitedSender` 在并发未满时立即发送，并发已满但等待队列未满时排队等待；
- `LimitedSender` 在并发和等待队列都满时返回 `ErrMailBusy`，不会静默丢弃邮件；
- `ErrMailBusy` 表示邮件没有发送，也没有进入队列，调用方必须显式处理；
- 默认最大邮件大小为 10MB，可通过 `SMTPOptions.MaxMessageBytes` 调整；
- 本包不提供业务模板、验证码、队列、持久化重试和发送记录落库；
- 高频或必须补偿的邮件场景，建议业务系统使用任务表、MQ 或独立邮件服务组合 `Sender`。
