// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package dataurl

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// decodeDATAURL decodes data URL to data and mime type.
func Decode(dataURL string) (data []byte, mimeType string, err error) {
	parts := strings.Split(dataURL, ",")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid data URL: %s", dataURL)
	}
	if !strings.HasPrefix(parts[0], "data:") {
		return nil, "", fmt.Errorf("invalid data URL: %s", dataURL)
	}
	data, err = base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", fmt.Errorf("base64 decode failed: %w", err)
	}
	mimeType = parts[0]
	return
}

// encodeDATAURL encodes data to data URL.
func Encode(mimeType string, data []byte) string {
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
}
