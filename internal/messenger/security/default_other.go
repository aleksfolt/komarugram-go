// SPDX-License-Identifier: Unlicense OR MIT

//go:build !darwin

package security

// defaultTPM is the hardware the security configuration is sealed by.
func defaultTPM() TPM { return hardwareTPM{} }
