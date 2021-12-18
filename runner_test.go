package main

import (
	"runtime"
	"testing"
)

func TestA(t *testing.T) {
	t.Log("OS:", runtime.GOOS)
}

func TestB(t *testing.T) {
}
