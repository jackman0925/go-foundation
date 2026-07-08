# crypto

`crypto` 提供常用哈希、签名、随机数字、密码哈希和 checksum 工具。

## 导入

```go
import foundationCrypto "github.com/jackman0925/go-foundation/crypto"
```

## 基础用法

```go
md5Text := foundationCrypto.MD5Hex("hello")
md5Short := foundationCrypto.MD5Hex16("hello")
shaText := foundationCrypto.SHA256Hex("hello")
sign := foundationCrypto.HMACSHA256Hex("secret", "hello")
code, err := foundationCrypto.RandomDigits(6)
hash, err := foundationCrypto.HashPassword("secret")
ok := foundationCrypto.CheckPassword("secret", hash)
checksum := foundationCrypto.MapChecksumMD5(map[string]any{"name": "demo"})
```

## 注意事项

- 本包不提供支付签名、License 加密、RSA 私钥管理或证书管理；
- 随机数字使用 `crypto/rand`；
- 密码哈希使用 bcrypt；
- `MapChecksumMD5` 会按 key 排序，保证 map 插入顺序不影响结果。
