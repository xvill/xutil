package xutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
)

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
	return Crypto{Key: key, IV: key, Mode: "CBC", Algorithm: "AES", Padding: "PKCS5"}
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
func (c *Crypto) Encrypt(data []byte) (text []byte, err error) {
	key, iv := c.Key, c.IV
	var block cipher.Block
	if c.Algorithm == "AES" {
		block, err = aes.NewCipher(key)
	} else if c.Algorithm == "DES" {
		block, err = des.NewCipher(key)
	}
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	data = c.padding(data, bs)
	text = make([]byte, len(data))
	switch c.Mode {
	case "ECB":
		// dst := text
		for len(data) > 0 {
			block.Encrypt(text, data[:bs])
			data = data[bs:]
			text = text[bs:]
		}
	case "CBC":
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(text, data)
	case "CTR":
		cipher.NewCTR(block, iv).XORKeyStream(text, data)
	case "OFB":
		cipher.NewOFB(block, iv).XORKeyStream(text, data)
	case "CFB":
		cipher.NewCFBEncrypter(block, iv).XORKeyStream(text, data)
	}
	return text, nil
}

func (c *Crypto) Decrypt(data []byte) (text []byte, err error) {
	key, iv := c.Key, c.IV
	var block cipher.Block
	if c.Algorithm == "AES" {
		block, err = aes.NewCipher(key)
	} else if c.Algorithm == "DES" {
		block, err = des.NewCipher(key)
	}
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	text = make([]byte, len(data))
	bs := block.BlockSize()
	switch c.Mode {
	case "ECB":
		dst := text
		for len(data) > 0 {
			block.Decrypt(dst, data[:bs])
			data = data[bs:]
			dst = dst[bs:]
		}
	case "CBC":
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(text, data)
	case "CTR":
		cipher.NewCTR(block, iv).XORKeyStream(text, data)
	case "OFB":
		cipher.NewOFB(block, iv).XORKeyStream(text, data)
	case "CFB":
		cipher.NewCFBDecrypter(block, iv).XORKeyStream(text, data)
	}
	return c.unpadding(text), nil
}
