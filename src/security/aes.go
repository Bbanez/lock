package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltSize  = 16
	nonceSize = 12
	keySize   = 32 // AES-256
	iter      = 100_000
)

var lockFlag = []byte{0x6C, 0x6F, 0x63, 0x6B} // "lock"

func Encrypt(password string, inData []byte, checkIfEnc bool) ([]byte, error) {
	if checkIfEnc {
		_, err := Decrypt(password, inData, false)
		if err == nil {
			return nil, fmt.Errorf("data is already encrypted")
		}
	}
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	key := pbkdf2.Key([]byte(password), salt, iter, keySize, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	data := append(lockFlag, inData...)
	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	result := append(salt, nonce...)
	result = append(result, ciphertext...)
	return result, nil
}

func Decrypt(password string, enc []byte, checkIfDec bool) ([]byte, error) {
	if len(enc) < saltSize+nonceSize {
		return nil, fmt.Errorf("data len is too short to be valid")
	}
	salt := enc[:saltSize]
	nonce := enc[saltSize : saltSize+nonceSize]
	ciphertext := enc[saltSize+nonceSize:]
	key := pbkdf2.Key([]byte(password), salt, iter, keySize, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	lockCheck := plaintext[0:len(lockFlag)]
	if string(lockCheck) != string(lockFlag) {
		return nil, fmt.Errorf("data is not encrypted or password is incorrect")
	}
	plaintext = plaintext[len(lockFlag):]
	return plaintext, nil
}
