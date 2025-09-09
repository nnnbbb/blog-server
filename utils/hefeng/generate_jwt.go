package hefeng

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"
)

// 生成 EdDSA (Ed25519) JWT。ttl 建议前端短、服务端长，但<=24h。
func GenerateJWT(kid, sub string, privPEM []byte, ttl time.Duration) (string, error) {
	if kid == "" || sub == "" {
		return "", errors.New("kid/sub 不能为空")
	}
	// 最长 24 小时
	if ttl <= 0 || ttl > 24*time.Hour {
		ttl = 24 * time.Hour
	}

	// 解析 PKCS#8 Ed25519 私钥
	priv, err := parseEd25519PrivateKeyFromPEM(privPEM)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %w", err)
	}

	// Header & Payload
	header := map[string]any{
		"alg": "EdDSA",
		"kid": kid,
	}
	now := time.Now().Unix()
	payload := map[string]any{
		"sub": sub,
		"iat": now - 30,                     // 防止时间误差
		"exp": now + int64(ttl/time.Second), // 过期时间
	}

	// JSON 编码（紧凑）
	hb, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("编码 header 失败: %w", err)
	}
	pb, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("编码 payload 失败: %w", err)
	}

	// Base64URL(无填充)
	b64 := base64.RawURLEncoding
	hEnc := b64.EncodeToString(hb)
	pEnc := b64.EncodeToString(pb)

	// 签名
	signingInput := hEnc + "." + pEnc
	sig := ed25519.Sign(priv, []byte(signingInput))
	sEnc := b64.EncodeToString(sig)

	return signingInput + "." + sEnc, nil
}

func parseEd25519PrivateKeyFromPEM(pemBytes []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("PEM 解码失败")
	}
	// 和风的示例是 PKCS#8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("PKCS#8 解析失败: %w", err)
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("不是 Ed25519 私钥")
	}
	return priv, nil
}

func GenerateToken() (string, error) {
	const kid = "TN5849VDFJ"
	const sub = "2FKRUP4CBG"

	pemPath := os.Getenv("HEFENG_PEM_PATH")
	privateKeyPEM, err := os.ReadFile(pemPath)

	if err != nil {
		panic(err) // 或者 return err
	}
	// 这里示例设置 10 分钟有效期；可按需调整（上限 24h）
	token, err := GenerateJWT(kid, sub, []byte(privateKeyPEM), 1*time.Minute)
	if err != nil {
		panic(err)
	}
	// fmt.Println(token)
	return token, nil
}
