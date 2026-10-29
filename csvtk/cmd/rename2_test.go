package cmd

import "testing"

// TestRename2LiteralKV preserves dollars when mapped values replace header names.
func TestRename2LiteralKV(t *testing.T) { testLiteralKV(t, "rename2") }
