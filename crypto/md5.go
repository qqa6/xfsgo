package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5Str(src []byte) string {
	h := md5.New()
	h.Write(src)
	return hex.EncodeToString(h.Sum(nil))
}