// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tool

import (
	"os"
	"testing"
)

func TestRunWASM(t *testing.T) {
	wasmf, err := os.Open("testdata/echo.wasm")
	if err != nil {
		t.Fatal(err)
	}
	defer wasmf.Close()

	resources := []*Resource{{
		reader: wasmf,
	}}
	args := Arguments{"name": "echo"}
	r, err := newWASMRunner(t.Context(), args, resources)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Run(t.Context(), []byte("yo!"))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "yo!" {
		t.Fatalf("expected 'yo!', got %q", string(out))
	}
}
