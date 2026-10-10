package infrastructure

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"errors"
	"crypto/subtle"
	"golang.org/x/crypto/argon2"
)

const (
	memory uint32 = 64 * 1024
	iterations uint32 = 3
	parallelism uint8 = 2
	keyLength uint32 = 32
	saltLength uint32 = 16
)
 func HashPassword(password string) (string, error){
 	salt := make([]byte, saltLength)
  	if _, err := rand.Read(salt); err != nil {
   		return "", err
   }
   hash := argon2.IDKey(
   		[]byte(password),
     	salt,
      	iterations,
       	memory,
        parallelism,
       	keyLength,
   	)
   encoder := base64.RawStdEncoding
   return fmt.Sprintf(
   "$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memory,iterations, parallelism, encoder.EncodeToString(salt), encoder.EncodeToString(hash)), nil
}

func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")

	if len(parts) != 6 ||
		parts[1] != "argon2id" ||
		parts[2] != "v=19" ||
		parts[3] != "m=65536,t=3,p=2" {
		return false
	}

	decoder := base64.RawStdEncoding

	salt, err := decoder.DecodeString(parts[4])
	if err != nil || len(salt) != int(saltLength) {
		return false
	}

	expected, err := decoder.DecodeString(parts[5])
	if err != nil || len(expected) != int(keyLength) {
		return false
	}

	actual := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	return subtle.ConstantTimeCompare(actual, expected) == 1
}

var ErrInvalidCredentials = errors.New("invalid credentials")
