// SPDX-License-Identifier: Unlicense OR MIT

package security

import (
	"bytes"
	"testing"
)

// The test uses the login keychain of whoever runs it, with one item of its
// own, which it removes.
func TestKeychainSealUnseal(t *testing.T) {
	k := keychainKey{}
	if err := k.Probe(); err != nil {
		t.Skipf("no keychain: %v", err)
	}
	secret := []byte("0123456789abcdef0123456789abcdef")
	public, private, err := k.Seal(secret, []byte("auth"))
	if err != nil {
		t.Fatal(err)
	}
	defer k.Forget(public)
	got, err := k.Unseal(public, private, []byte("auth"))
	if err != nil || !bytes.Equal(got, secret) {
		t.Fatalf("Unseal = %q, %v", got, err)
	}
	if _, err := k.Unseal(public, private, []byte("wrong")); err == nil {
		t.Fatal("a wrong authorization opened the sealed object")
	}
	k.Forget(public)
	if _, err := k.Unseal(public, private, []byte("auth")); err == nil {
		t.Fatal("the sealed object opened without its keychain item")
	}
}
