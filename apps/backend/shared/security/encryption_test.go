package security

import (
	"testing"
)

func TestAESGCM_EncryptDecrypt(t *testing.T) {
	secret := "acis_secret_otp_encryption_key_32bytes!!"
	plainText := "948201"

	cipherText, err := EncryptAESGCM(plainText, secret)
	if err != nil {
		t.Fatalf("EncryptAESGCM failed: %v", err)
	}

	if cipherText == plainText {
		t.Fatal("ciphertext should not equal plaintext")
	}

	decrypted, err := DecryptAESGCM(cipherText, secret)
	if err != nil {
		t.Fatalf("DecryptAESGCM failed: %v", err)
	}

	if decrypted != plainText {
		t.Errorf("expected decrypted text %s, got %s", plainText, decrypted)
	}
}

func TestAESGCM_WrongKeyFails(t *testing.T) {
	secret1 := "secret_key_one_32_bytes_length!"
	secret2 := "secret_key_two_32_bytes_length!"
	plainText := "123456"

	cipherText, err := EncryptAESGCM(plainText, secret1)
	if err != nil {
		t.Fatalf("EncryptAESGCM failed: %v", err)
	}

	_, err = DecryptAESGCM(cipherText, secret2)
	if err == nil {
		t.Fatal("expected decryption failure with wrong key, got success")
	}
}

func TestAESGCM_EmptyKeyFails(t *testing.T) {
	_, err := EncryptAESGCM("123456", "")
	if err == nil {
		t.Fatal("expected error on empty key for encryption, got nil")
	}

	_, err = DecryptAESGCM("someciphertext", "")
	if err == nil {
		t.Fatal("expected error on empty key for decryption, got nil")
	}
}
