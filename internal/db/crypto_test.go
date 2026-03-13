package db

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// testKey is a base64-encoded 32-byte key used only in tests.
const testKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

func setTestEncryptionEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TOKEN_ENCRYPTION_KEY_1", testKey)
	t.Setenv("TOKEN_ENCRYPTION_KEY_CURRENT", "1")
	t.Setenv("TOKEN_ENCRYPTION_KEY", "") // suppress legacy fallback
}

// ---------- parseBlobVersion ----------

func Test_parseBlobVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		encoded     string
		wantVersion int
		wantRest    string
		wantErr     bool
	}{
		{
			name:        "versioned v1",
			encoded:     "v1.nonce.cipher",
			wantVersion: 1,
			wantRest:    "nonce.cipher",
		},
		{
			name:        "versioned v2",
			encoded:     "v2.abc.def",
			wantVersion: 2,
			wantRest:    "abc.def",
		},
		{
			name:        "legacy no prefix treated as v1",
			encoded:     "nonce.cipher",
			wantVersion: 1,
			wantRest:    "nonce.cipher",
		},
		{
			name:    "invalid — v with no dot",
			encoded: "v",
			wantErr: true,
		},
		{
			name:    "invalid — v with dot too early",
			encoded: "v.",
			wantErr: true,
		},
		{
			name:    "invalid — v with non-numeric version",
			encoded: "vX.rest",
			wantErr: true,
		},
		{
			name:        "version with multiple dots in rest",
			encoded:     "v3.a.b.c",
			wantVersion: 3,
			wantRest:    "a.b.c",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			version, rest, err := parseBlobVersion(tc.encoded)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != tc.wantVersion {
				t.Errorf("version = %d, want %d", version, tc.wantVersion)
			}
			if rest != tc.wantRest {
				t.Errorf("rest = %q, want %q", rest, tc.wantRest)
			}
		})
	}
}

// ---------- parseRawKey ----------

func Test_parseRawKey(t *testing.T) {
	t.Parallel()

	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{
			name: "valid 32-byte key",
			raw:  validKey,
		},
		{
			name:    "not base64",
			raw:     "not-valid-base64!@#",
			wantErr: true,
		},
		{
			name:    "too short — 16 bytes",
			raw:     base64.StdEncoding.EncodeToString(make([]byte, 16)),
			wantErr: true,
		},
		{
			name:    "too long — 64 bytes",
			raw:     base64.StdEncoding.EncodeToString(make([]byte, 64)),
			wantErr: true,
		},
		{
			name:    "empty string",
			raw:     "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			key, err := parseRawKey(tc.raw, "TEST_KEY")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil; key len=%d", len(key))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(key) != 32 {
				t.Errorf("key length = %d, want 32", len(key))
			}
		})
	}
}

// ---------- deriveUserKey ----------

func Test_deriveUserKey(t *testing.T) {
	t.Parallel()

	masterKey := make([]byte, 32)

	tests := []struct {
		name   string
		userID int64
	}{
		{name: "user 1", userID: 1},
		{name: "user 0", userID: 0},
		{name: "large user ID", userID: 999_999_999},
		{name: "negative user ID", userID: -1},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			key, err := deriveUserKey(masterKey, tc.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(key) != 32 {
				t.Errorf("derived key length = %d, want 32", len(key))
			}
		})
	}

	// Different users must produce different keys.
	t.Run("per-user isolation", func(t *testing.T) {
		t.Parallel()
		key1, _ := deriveUserKey(masterKey, 1)
		key2, _ := deriveUserKey(masterKey, 2)
		if string(key1) == string(key2) {
			t.Error("different users derived the same key — isolation violated")
		}
	})

	// Same input must be deterministic.
	t.Run("deterministic", func(t *testing.T) {
		t.Parallel()
		key1, _ := deriveUserKey(masterKey, 42)
		key2, _ := deriveUserKey(masterKey, 42)
		if string(key1) != string(key2) {
			t.Error("same inputs produced different keys — must be deterministic")
		}
	})
}

// ---------- EncryptCredentials / DecryptCredentials roundtrip ----------

func TestEncryptDecryptRoundtrip(t *testing.T) {
	setTestEncryptionEnv(t)

	tests := []struct {
		name      string
		plaintext string
		userID    int64
	}{
		{name: "simple JSON", plaintext: `{"client_id":"id","client_secret":"sec","refresh_token":"tok"}`, userID: 1},
		{name: "empty string", plaintext: "", userID: 1},
		{name: "unicode content", plaintext: "привет мир 🌍", userID: 2},
		{name: "different user", plaintext: "same plaintext", userID: 99},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			blob, err := EncryptCredentials(tc.plaintext, tc.userID)
			if err != nil {
				t.Fatalf("EncryptCredentials: %v", err)
			}

			got, err := DecryptCredentials(blob, tc.userID)
			if err != nil {
				t.Fatalf("DecryptCredentials: %v", err)
			}
			if got != tc.plaintext {
				t.Errorf("roundtrip = %q, want %q", got, tc.plaintext)
			}
		})
	}
}

func TestEncryptCredentials_BlobFormat(t *testing.T) {
	setTestEncryptionEnv(t)

	blob, err := EncryptCredentials("test", 1)
	if err != nil {
		t.Fatalf("EncryptCredentials: %v", err)
	}

	// Blob must start with "v1."
	if !strings.HasPrefix(blob, "v1.") {
		t.Errorf("blob %q does not start with v1.", blob)
	}

	// Blob must have exactly 3 dot-separated parts: v<N>, nonce, ciphertext.
	parts := strings.SplitN(blob, ".", 3)
	if len(parts) != 3 {
		t.Fatalf("blob has %d parts, want 3", len(parts))
	}

	// Nonce and ciphertext must be valid base64.
	for i, part := range parts[1:] {
		if _, err := base64.StdEncoding.DecodeString(part); err != nil {
			t.Errorf("parts[%d] is not valid base64: %v", i+1, err)
		}
	}
}

func TestDecryptCredentials_WrongUser(t *testing.T) {
	setTestEncryptionEnv(t)

	blob, err := EncryptCredentials("secret data", 1)
	if err != nil {
		t.Fatalf("EncryptCredentials: %v", err)
	}

	// Decrypting with a different user ID must fail (different derived key).
	_, err = DecryptCredentials(blob, 2)
	if err == nil {
		t.Error("expected decryption to fail for wrong user, but it succeeded")
	}
}

func TestDecryptCredentials_Tampered(t *testing.T) {
	setTestEncryptionEnv(t)

	blob, err := EncryptCredentials("sensitive", 1)
	if err != nil {
		t.Fatalf("EncryptCredentials: %v", err)
	}

	// Flip a byte in the ciphertext portion.
	parts := strings.SplitN(blob, ".", 3)
	cipher, _ := base64.StdEncoding.DecodeString(parts[2])
	cipher[0] ^= 0xFF
	parts[2] = base64.StdEncoding.EncodeToString(cipher)
	tampered := strings.Join(parts, ".")

	_, err = DecryptCredentials(tampered, 1)
	if err == nil {
		t.Error("expected decryption to fail on tampered ciphertext, but it succeeded")
	}
}

func TestDecryptCredentials_MissingKeyVersion(t *testing.T) {
	// Blob version 2 when only key 1 is configured.
	t.Setenv("TOKEN_ENCRYPTION_KEY_1", testKey)
	t.Setenv("TOKEN_ENCRYPTION_KEY_CURRENT", "1")
	t.Setenv("TOKEN_ENCRYPTION_KEY", "")

	_, err := DecryptCredentials("v2.nonce.ciphertext", 1)
	if err == nil {
		t.Error("expected error for unknown key version, got nil")
	}
}

// ---------- Benchmarks ----------

func BenchmarkEncryptCredentials(b *testing.B) {
	b.Setenv("TOKEN_ENCRYPTION_KEY_1", testKey)
	b.Setenv("TOKEN_ENCRYPTION_KEY_CURRENT", "1")
	b.Setenv("TOKEN_ENCRYPTION_KEY", "")

	plaintext := fmt.Sprintf(`{"client_id":"test_client","client_secret":"test_secret","refresh_token":"%s"}`,
		strings.Repeat("x", 100))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = EncryptCredentials(plaintext, 12345)
	}
}

func BenchmarkDecryptCredentials(b *testing.B) {
	b.Setenv("TOKEN_ENCRYPTION_KEY_1", testKey)
	b.Setenv("TOKEN_ENCRYPTION_KEY_CURRENT", "1")
	b.Setenv("TOKEN_ENCRYPTION_KEY", "")

	plaintext := `{"client_id":"id","client_secret":"sec","refresh_token":"tok"}`
	blob, err := EncryptCredentials(plaintext, 12345)
	if err != nil {
		b.Fatalf("setup: %v", err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = DecryptCredentials(blob, 12345)
	}
}

// ---------- Fuzz ----------

func FuzzParseBlobVersion(f *testing.F) {
	// Seed with representative inputs.
	f.Add("v1.nonce.cipher")
	f.Add("v2.abc.def")
	f.Add("nonce.cipher")
	f.Add("v")
	f.Add("")
	f.Add("v0.x.y")

	f.Fuzz(func(t *testing.T, s string) {
		// Must not panic; error is acceptable.
		_, _, _ = parseBlobVersion(s)
	})
}
