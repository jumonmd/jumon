// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package subject

import (
	"encoding/base64"
)

// Escape escapes NATS subject name.
func Escape(subject string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(subject))
}

// Unescape unescapes NATS subject name.
func Unescape(escaped string) string {
	decoded, err := base64.RawURLEncoding.DecodeString(escaped)
	if err != nil {
		return ""
	}
	return string(decoded)
}
