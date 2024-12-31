package main

import (
	"reflect"
	"testing"

	tu "master.private/bstd.git/testutil"
)

func Test_decodeInRoutes(t *testing.T) {
	payload := "7b226261644368616e73223a5b5d2c226261644e6f646573223a5b5d2c2266726f6d223a5b22414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141225d2c22736174223a313030302c22746f223a22424242424242424242424242424242424242424242424242424242424242424242424242424242424242424242424242424242424242424242424242424242424242227d"

	expected := inRoutes{
		Sat:      1000,
		BadNodes: []string{},
		BadChans: []int64{},
		From: []string{
			"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		},
		To: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	}

	r, err := decodeInRoutes([]byte(payload))
	tu.Must(t, err)

	if !reflect.DeepEqual(r, expected) {
		t.Fatalf("unexpected: %+v", r)
	}
}

func Test_shortChannelIdToInt(t *testing.T) {
	payload := "506015x904x1"
	var expected int64 = 556369376388317185

	r := mustShortChannelIdToInt(payload)

	if r != expected {
		t.Fatal("unexpected:", r)
	}
}

func Test_shortChannelIdToString(t *testing.T) {
	var payload int64 = 556369376388317185
	expected := "506015x904x1"

	r := shortChannelIdToString(payload)

	if r != expected {
		t.Fatal("unexpected:", string(r))
	}
}
