package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"master.private/bstd.git/stackerr"
)

func decodeInRoutes(hexStr []byte) (inRoutes, error) {
	var r inRoutes
	res := make([]byte, len(hexStr)/2+1)
	n, err := hex.Decode(res, hexStr)
	if err != nil {
		return r, stackerr.Wrap(err)
	}
	res = res[:n]
	err = json.Unmarshal(res, &r)
	if err != nil {
		return r, stackerr.Wrap(err)
	}
	return r, nil
}

func shortChannelIdToInt(scid string) (int64, error) {
	const nParts = 3
	var ints [nParts]int64
	nums := strings.Split(scid, "x")
	if n := len(nums); n != nParts {
		return 0, fmt.Errorf("expecting %d separators, got %d", nParts, n)
	}
	for i := range nParts {
		asInt, err := strconv.ParseInt(string(nums[i]), 10, 64)
		if err != nil {
			return 0, stackerr.Wrap(err)
		}
		ints[i] = asInt
	}

	// Block Height (24 bits) (<< 40)
	// Transaction Index (24 bits) (<< 16)
	// Output Index (16 bits) (<< 0)
	result := ints[0]<<40 | ints[1]<<16 | ints[2]
	return result, nil
}

func shortChannelIdToString(scid int64) string {
	ints := [...]int64{
		scid >> 40,
		scid >> 16 & 0xFFFFFF,
		scid & 0xFFFF,
	}
	return fmt.Sprintf("%dx%dx%d", ints[0], ints[1], ints[2])
}
