package xutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
)

//================================================================================================

//Hash Hash
type Hash struct{ Data string }

//NewHash NewHash
func NewHash(dat string) Hash {
	return Hash{Data: dat}
}

//MD5  MD5
func (h Hash) MD5() string {
	c := md5.New()
	c.Write([]byte(h.Data))
	return hex.EncodeToString(c.Sum(nil))
}

//SHA1  SHA1
func (h Hash) SHA1() string {
	c := sha1.New()
	c.Write([]byte(h.Data))
	return hex.EncodeToString(c.Sum(nil))
}

//================================================================================================
//明文补码算法
func pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//明文减码算法
func pKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//明文补码算法: 用0去填充
func zeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

//明文减码算法：用0去填充
func zeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}

//================================================================================================
// var modes = []string{"ECB", "CBC", "CTR", "OFB", "CFB"}
// var paddings = []string{"PKCS5", "ZERO"}

type Crypto struct {
	Key, IV                  []byte
	Algorithm, Mode, Padding string
}

func NewCrypto(key []byte) Crypto {
	return Crypto{Key: key, IV: key, Algorithm: "AES", Mode: "CBC", Padding: "PKCS5"}
}

func (c *Crypto) padding(data []byte, blockSize int) []byte {
	if c.Padding == "PKCS5" {
		return pKCS5Padding(data, blockSize)
	}
	return zeroPadding(data, blockSize)

}
func (c *Crypto) unpadding(data []byte) []byte {
	if c.Padding == "PKCS5" {
		return pKCS5UnPadding(data)
	}
	return zeroUnPadding(data)
}

//================================================================================================
func (c *Crypto) Encrypt(data []byte) (dst []byte, err error) {
	var block cipher.Block
	if c.Algorithm == "AES" {
		block, err = aes.NewCipher(c.Key)
	} else if c.Algorithm == "DES" {
		block, err = des.NewCipher(c.Key)
	}
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	data = c.padding(data, bs)
	dst = make([]byte, len(data))
	switch c.Mode {
	case "ECB":
		for len(data) > 0 {
			block.Encrypt(dst, data[:bs])
			data = data[bs:]
			dst = dst[bs:]
		}
	case "CBC":
		cipher.NewCBCEncrypter(block, c.IV).CryptBlocks(dst, data)
	case "CTR":
		cipher.NewCTR(block, c.IV).XORKeyStream(dst, data)
	case "OFB":
		cipher.NewOFB(block, c.IV).XORKeyStream(dst, data)
	case "CFB":
		cipher.NewCFBEncrypter(block, c.IV).XORKeyStream(dst, data)
	}
	return dst, nil
}

func (c *Crypto) Decrypt(data []byte) (dst []byte, err error) {
	var block cipher.Block
	if c.Algorithm == "AES" {
		block, err = aes.NewCipher(c.Key)
	} else if c.Algorithm == "DES" {
		block, err = des.NewCipher(c.Key)
	}
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	dst = make([]byte, len(data))
	bs := block.BlockSize()
	switch c.Mode {
	case "ECB":
		for len(data) > 0 {
			block.Decrypt(dst, data[:bs])
			data = data[bs:]
			dst = dst[bs:]
		}
	case "CBC":
		cipher.NewCBCDecrypter(block, c.IV).CryptBlocks(dst, data)
	case "CTR":
		cipher.NewCTR(block, c.IV).XORKeyStream(dst, data)
	case "OFB":
		cipher.NewOFB(block, c.IV).XORKeyStream(dst, data)
	case "CFB":
		cipher.NewCFBDecrypter(block, c.IV).XORKeyStream(dst, data)
	}
	return c.unpadding(dst), nil
}
