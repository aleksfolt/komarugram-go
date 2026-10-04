// SPDX-License-Identifier: Unlicense OR MIT

package security

/*
#cgo LDFLAGS: -framework Security -framework CoreFoundation
#include <Security/Security.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

#define KC_SERVICE "komarugram-go"

static CFMutableDictionaryRef kcQuery(const char *account) {
	CFMutableDictionaryRef q = CFDictionaryCreateMutable(NULL, 0, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	CFStringRef service = CFStringCreateWithCString(NULL, KC_SERVICE, kCFStringEncodingUTF8);
	CFStringRef acct = CFStringCreateWithCString(NULL, account, kCFStringEncodingUTF8);
	CFDictionarySetValue(q, kSecClass, kSecClassGenericPassword);
	CFDictionarySetValue(q, kSecAttrService, service);
	CFDictionarySetValue(q, kSecAttrAccount, acct);
	CFRelease(service);
	CFRelease(acct);
	return q;
}

static int kcAdd(const char *account, const unsigned char *data, int n) {
	CFMutableDictionaryRef q = kcQuery(account);
	CFDataRef value = CFDataCreate(NULL, data, n);
	CFDictionarySetValue(q, kSecValueData, value);
	OSStatus st = SecItemAdd(q, NULL);
	CFRelease(value);
	CFRelease(q);
	return (int)st;
}

// kcGet copies the item into out, which holds cap bytes, and its length into n.
static int kcGet(const char *account, unsigned char *out, int cap, int *n) {
	CFMutableDictionaryRef q = kcQuery(account);
	CFDictionarySetValue(q, kSecReturnData, kCFBooleanTrue);
	CFDictionarySetValue(q, kSecMatchLimit, kSecMatchLimitOne);
	CFTypeRef res = NULL;
	OSStatus st = SecItemCopyMatching(q, &res);
	CFRelease(q);
	if (st != errSecSuccess) {
		return (int)st;
	}
	CFIndex len = CFDataGetLength((CFDataRef)res);
	if (len > cap) {
		CFRelease(res);
		return errSecAllocate;
	}
	CFDataGetBytes((CFDataRef)res, CFRangeMake(0, len), out);
	*n = (int)len;
	CFRelease(res);
	return errSecSuccess;
}

static int kcDelete(const char *account) {
	CFMutableDictionaryRef q = kcQuery(account);
	OSStatus st = SecItemDelete(q);
	CFRelease(q);
	return (int)st;
}

static int kcProbe(void) {
	unsigned char b[1];
	int n = 0;
	int st = kcGet("probe", b, sizeof b, &n);
	return st == errSecItemNotFound ? errSecSuccess : st;
}

static char *kcMessage(int st) {
	CFStringRef s = SecCopyErrorMessageString((OSStatus)st, NULL);
	if (s == NULL) {
		return NULL;
	}
	CFIndex size = CFStringGetMaximumSizeForEncoding(CFStringGetLength(s), kCFStringEncodingUTF8) + 1;
	char *buf = malloc(size);
	if (!CFStringGetCString(s, buf, size, kCFStringEncodingUTF8)) {
		free(buf);
		buf = NULL;
	}
	CFRelease(s);
	return buf;
}
*/
import "C"

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/crypto/chacha20poly1305"
)

// keychainKey is the TPM of a Mac: a random key kept in the login keychain,
// which the system releases to this program only, and only while the user is
// logged in. The sealed object is the secret wrapped by that key and by the
// authorization, so the password is still needed to open it.
//
// The keychain ties an item to the code signature of the program which made
// it: a rebuilt, unsigned binary is asked about once by the system.
type keychainKey struct{}

// defaultTPM is the hardware the security configuration is sealed by.
func defaultTPM() TPM { return keychainKey{} }

func keychainError(what string, st C.int) error {
	msg := ""
	if m := C.kcMessage(st); m != nil {
		msg = C.GoString(m)
		C.free(unsafe.Pointer(m))
	}
	return fmt.Errorf("keychain %s: %s (%d)", what, msg, int(st))
}

func (keychainKey) Probe() error {
	if st := C.kcProbe(); st != 0 {
		return fmt.Errorf("%w: %w", ErrUnavailable, keychainError("probe", st))
	}
	return nil
}

func wrapKey(item, authorization []byte) []byte {
	mac := hmac.New(sha256.New, item)
	mac.Write(authorization)
	return mac.Sum(nil)
}

func (keychainKey) Seal(secret, authorization []byte) (public, private []byte, err error) {
	item := make([]byte, keySize)
	id := make([]byte, 16)
	if _, err := rand.Read(item); err != nil {
		return nil, nil, err
	}
	if _, err := rand.Read(id); err != nil {
		return nil, nil, err
	}
	aead, err := chacha20poly1305.NewX(wrapKey(item, authorization))
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	private = aead.Seal(nonce, nonce, secret, id)
	account := C.CString(hex.EncodeToString(id))
	defer C.free(unsafe.Pointer(account))
	if st := C.kcAdd(account, (*C.uchar)(unsafe.Pointer(&item[0])), C.int(len(item))); st != 0 {
		return nil, nil, keychainError("add", st)
	}
	return id, private, nil
}

func (keychainKey) Unseal(public, private, authorization []byte) ([]byte, error) {
	account := C.CString(hex.EncodeToString(public))
	defer C.free(unsafe.Pointer(account))
	item := make([]byte, 256)
	var n C.int
	if st := C.kcGet(account, (*C.uchar)(unsafe.Pointer(&item[0])), C.int(len(item)), &n); st != 0 {
		return nil, keychainError("read", st)
	}
	aead, err := chacha20poly1305.NewX(wrapKey(item[:n], authorization))
	if err != nil {
		return nil, err
	}
	if len(private) < aead.NonceSize() {
		return nil, errors.New("keychain: sealed object is too short")
	}
	return aead.Open(nil, private[:aead.NonceSize()], private[aead.NonceSize():], public)
}

// forget removes the keychain item of a sealed object.
func (keychainKey) Forget(public []byte) {
	account := C.CString(hex.EncodeToString(public))
	defer C.free(unsafe.Pointer(account))
	C.kcDelete(account)
}
