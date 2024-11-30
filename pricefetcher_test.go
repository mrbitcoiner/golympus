package main

import (
	"testing"

	"master.private/bstd.git/testutil"
)

func TestIntegrationFetchSymbols(t *testing.T) {
	pf := NewPriceFetcher()

	res, err := pf.Fetch(USD, EUR, JPY, CNY, BRL)
	testutil.Must(t, err)
	if len(res) != 5 {
		t.Fatal("unexpected len", len(res))
	}
	for k, v := range res {
		if v <= 0 {
			t.Fatal("unexpected price to symbol", k)
		}
	}

	res1, err := pf.Fetch(USD, EUR, JPY, CNY, BRL)
	testutil.Must(t, err)
	if len(res1) != 5 {
		t.Fatal("unexpected len", len(res))
	}
	for k, v := range res1 {
		if v <= 0 {
			t.Fatal("unexpected price to symbol", k)
		}
	}

	if pf.fetchCount != 1 {
		t.Fatal("unexpected fetch count", pf.fetchCount)
	}
}
