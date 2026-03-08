package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/hkdf"
)

// EncryptCredentials encrypts plaintext using a per-user AES-256-GCM key derived
// from the current master key version via HKDF. The returned blob is versioned:
// "v<N>.<base64(nonce)>.<base64(ciphertext)>".
func EncryptCredentials(plaintext string, userID int64) (string, error) {
	keys, current, err := loadEncryptionKeys()
	if err != nil {
		return "", err
	}
	key, err := deriveUserKey(keys[current], userID)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := aesGCM.Seal(nil, nonce, []byte(plaintext), nil)

	return fmt.Sprintf("v%d.%s.%s",
		current,
		base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(ciphertext),
	), nil
}

// DecryptCredentials decrypts a versioned blob produced by EncryptCredentials.
// It selects the correct master key by version and derives the per-user subkey.
func DecryptCredentials(encoded string, userID int64) (string, error) {
	version, rest, err := parseBlobVersion(encoded)
	if err != nil {
		return "", err
	}
	keys, _, err := loadEncryptionKeys()
	if err != nil {
		return "", err
	}
	masterKey, ok := keys[version]
	if !ok {
		return "", fmt.Errorf("no encryption key configured for version %d", version)
	}
	key, err := deriveUserKey(masterKey, userID)
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(rest, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted credentials format")
	}
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credentials: %w", err)
	}
	return string(plaintext), nil
}

// deriveUserKey derives a 32-byte AES key from masterKey and userID using HKDF-SHA256.
// Each user gets a unique subkey so a single compromised derived key does not expose others.
func deriveUserKey(masterKey []byte, userID int64) ([]byte, error) {
	salt := []byte(strconv.FormatInt(userID, 10))
	r := hkdf.New(sha256.New, masterKey, salt, []byte("tg-gmail-parser-bot credentials"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("failed to derive user key: %w", err)
	}
	return key, nil
}

// parseBlobVersion extracts the version integer and the remainder of a versioned blob.
// Legacy blobs without a "v<N>." prefix are treated as version 1.
func parseBlobVersion(encoded string) (version int, rest string, err error) {
	if !strings.HasPrefix(encoded, "v") {
		return 1, encoded, nil
	}
	dot := strings.Index(encoded, ".")
	if dot < 2 {
		return 0, "", fmt.Errorf("invalid encrypted credentials format")
	}
	version, err = strconv.Atoi(encoded[1:dot])
	if err != nil {
		return 0, "", fmt.Errorf("invalid version in encrypted credentials: %w", err)
	}
	return version, encoded[dot+1:], nil
}

// loadEncryptionKeys reads versioned keys from the environment.
//
// Preferred form (supports rotation):
//
//	TOKEN_ENCRYPTION_KEY_1=<base64>   # oldest still-needed key
//	TOKEN_ENCRYPTION_KEY_2=<base64>   # newer key
//	TOKEN_ENCRYPTION_KEY_CURRENT=2    # which version to use for new encryptions
//
// Legacy fallback (single-key, no rotation):
//
//	TOKEN_ENCRYPTION_KEY=<base64>     # treated as version 1
func loadEncryptionKeys() (keys map[int][]byte, current int, err error) {
	// Try versioned form first.
	currentStr := os.Getenv("TOKEN_ENCRYPTION_KEY_CURRENT")
	if currentStr != "" {
		current, err = strconv.Atoi(currentStr)
		if err != nil || current < 1 {
			return nil, 0, fmt.Errorf("TOKEN_ENCRYPTION_KEY_CURRENT must be a positive integer")
		}
		keys = make(map[int][]byte)
		for _, env := range os.Environ() {
			pair := strings.SplitN(env, "=", 2)
			if len(pair) != 2 {
				continue
			}
			name, val := pair[0], pair[1]
			if !strings.HasPrefix(name, "TOKEN_ENCRYPTION_KEY_") {
				continue
			}
			suffix := strings.TrimPrefix(name, "TOKEN_ENCRYPTION_KEY_")
			n, convErr := strconv.Atoi(suffix)
			if convErr != nil {
				continue // skip CURRENT, PREVIOUS, etc.
			}
			k, decErr := parseRawKey(val, name)
			if decErr != nil {
				return nil, 0, decErr
			}
			keys[n] = k
		}
		if _, ok := keys[current]; !ok {
			return nil, 0, fmt.Errorf("TOKEN_ENCRYPTION_KEY_%d not set but is current version", current)
		}
		return keys, current, nil
	}

	// Legacy fallback: TOKEN_ENCRYPTION_KEY → version 1.
	raw := os.Getenv("TOKEN_ENCRYPTION_KEY")
	if raw == "" {
		return nil, 0, fmt.Errorf("TOKEN_ENCRYPTION_KEY_CURRENT or TOKEN_ENCRYPTION_KEY is required")
	}
	k, err := parseRawKey(raw, "TOKEN_ENCRYPTION_KEY")
	if err != nil {
		return nil, 0, err
	}
	return map[int][]byte{1: k}, 1, nil
}

func parseRawKey(raw, name string) ([]byte, error) {
	k, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be base64-encoded: %w", name, err)
	}
	if len(k) != 32 {
		return nil, fmt.Errorf("%s must decode to exactly 32 bytes, got %d", name, len(k))
	}
	return k, nil
}
