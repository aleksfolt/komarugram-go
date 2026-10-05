// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && cgo

package deviceinfo

/*
#include <stddef.h>
#include <features.h>
#ifdef __GLIBC__
#include <gnu/libc-version.h>
static const char *deviceinfo_libc(void) { return "glibc"; }
static const char *deviceinfo_libc_version(void) { return gnu_get_libc_version(); }
#else
static const char *deviceinfo_libc(void) { return NULL; }
static const char *deviceinfo_libc_version(void) { return NULL; }
#endif
*/
import "C"

// libcVersion is the C library the process runs with; only glibc tells.
func libcVersion() (name, version string) {
	if p := C.deviceinfo_libc(); p != nil {
		name = C.GoString(p)
	}
	if p := C.deviceinfo_libc_version(); p != nil {
		version = C.GoString(p)
	}
	return name, version
}
