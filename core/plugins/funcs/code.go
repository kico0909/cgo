package funcs

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
)

func MD5(str string) string {
	tmp := md5.New()
	tmp.Write([]byte(str))
	MD5Str := tmp.Sum(nil)
	return hex.EncodeToString(MD5Str)
}

func SHA1(str string) string {
	t := sha1.New()
	t.Write([]byte(str))
	sha1Str := t.Sum(nil)
	return hex.EncodeToString(sha1Str)
}

func HMAC_SHA1(str, key string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func BASE64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
