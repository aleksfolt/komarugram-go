// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !cgo

package deviceinfo

// libcVersion is unknown without cgo.
func libcVersion() (name, version string) { return "", "" }
