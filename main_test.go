package main

import (
	"testing"
)

func TestPrepareDirectory(t *testing.T) {
	dir := "docs"
	if prepareDirectory(dir) == "docs" {
		t.FailNow()
	}
}
