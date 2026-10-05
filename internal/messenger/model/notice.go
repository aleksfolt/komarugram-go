// SPDX-License-Identifier: Unlicense OR MIT

package model

// MessageNotice is a message that just came to the account, for a desktop
// notification.
type MessageNotice struct {
	Chat Chat
	// Sender is who wrote it in a group; empty elsewhere. Text is one line,
	// as the chat list shows it.
	Sender, Text string
	// Silent is set when the sender sent it without sound.
	Silent bool
}
