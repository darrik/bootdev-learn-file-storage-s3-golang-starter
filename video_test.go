package main

import (
	"testing"
)

func TestGetVideoAspectRatio(t *testing.T) {
	t.Parallel()

	a, err := getVideoAspectRatio("samples/boots-video-horizontal.mp4")
	t.Logf("a: %v\n", a)

	if a != "16:9" && a != "9:16" && a != "other" {
		t.Fatalf("horizontal har fel aspect ratio: %v\n", err)
	}
}
