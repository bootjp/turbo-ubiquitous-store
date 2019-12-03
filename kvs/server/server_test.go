package main

import (
	"testing"
)

func TestCommand(t *testing.T) {
	if v, err := commandParser([]byte("GET AAAAAAA")); err != nil || v != "GET" {
		t.Errorf("invalid command %s", v)
	}

	if v, err := commandParser([]byte("SET AAAAAAA")); err != nil || v != "GET" {
		t.Errorf("invalid command %s", v)
	}

	if v, err := commandParser([]byte("SET VVVV ")); err != nil || v != "GET" {
		t.Errorf("invalid command %s", v)
	}

}
