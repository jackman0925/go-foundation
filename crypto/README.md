# crypto

`crypto` 提供常用哈希、签名和随机数字工具。

## 导入

```go
import foundationCrypto "github.com/jackman0925/go-foundation/crypto"
```

## 基础用法

```go
md5Text := foundationCrypto.MD5Hex("hello")
shaText := foundationCrypto.SHA256Hex("hello")
sign := foundationCrypto.HMACSHA256Hex("secret", "hello")
code, err := foundationCrypto.RandomDigits(6)
```

## 注意事项

- 本包不提供支付签名、License 加密、RSA 私钥管理或证书管理；
- 随机数字使用 `crypto/rand`。
