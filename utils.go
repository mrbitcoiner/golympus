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

func mustShortChannelIdToInt(scid string) int64 {
	const nParts = 3
	var ints [nParts]int64
	nums := strings.Split(scid, "x")
	if n := len(nums); n != nParts {
		panic(stackerr.Wrap(fmt.Errorf("expecting %d separators, got %d", nParts, n)))
	}
	for i := range nParts {
		asInt, err := strconv.ParseInt(string(nums[i]), 10, 64)
		if err != nil {
			panic(stackerr.Wrap(err))
		}
		ints[i] = asInt
	}

	// Block Height (24 bits) (<< 40)
	// Transaction Index (24 bits) (<< 16)
	// Output Index (16 bits) (<< 0)
	result := ints[0]<<40 | ints[1]<<16 | ints[2]
	return result
}

func shortChannelIdToString(scid int64) string {
	ints := [...]int64{
		scid >> 40,
		scid >> 16 & 0xFFFFFF,
		scid & 0xFFFF,
	}
	return fmt.Sprintf("%dx%dx%d", ints[0], ints[1], ints[2])
}

func serializeRoutes(routes []PaymentRoute) [][]string {
	var serializedRoutes [][]string
	var hopSwap [118]byte

	for _, route := range routes {
		var serializedRoute []string
		for _, hop := range route {
			hopBytes := serializeHop(hop)
			hex.Encode(hopSwap[:], hopBytes[:])
			serializedRoute = append(serializedRoute, string(hopSwap[:]))
		}
		serializedRoutes = append(serializedRoutes, serializedRoute)
	}

	return serializedRoutes
}

func serializeHop(hop Hop) [59]byte {
	r := [59]byte{}

	// pubkey (33 bytes BE)
	_, err := hex.Decode(r[:33], []byte(hop.NodeId))
	must(err)

	// short channel id (8 bytes BE)
	r[33] = byte(hop.ShortChannelId >> (7 * 8))
	r[34] = byte(hop.ShortChannelId >> (6 * 8))
	r[35] = byte(hop.ShortChannelId >> (5 * 8))
	r[36] = byte(hop.ShortChannelId >> (4 * 8))
	r[37] = byte(hop.ShortChannelId >> (3 * 8))
	r[38] = byte(hop.ShortChannelId >> (2 * 8))
	r[39] = byte(hop.ShortChannelId >> (1 * 8))
	r[40] = byte(hop.ShortChannelId >> (0 * 8))

	// cltv delta (2 bytes BE)
	r[41] = byte(hop.CltvExpiryDelta >> (1 * 8))
	r[42] = byte(hop.CltvExpiryDelta >> (0 * 8))

	// htlc min msat (8 bytes BE)
	r[43] = byte(hop.HtlcMinimumMsat >> (7 * 8))
	r[44] = byte(hop.HtlcMinimumMsat >> (6 * 8))
	r[45] = byte(hop.HtlcMinimumMsat >> (5 * 8))
	r[46] = byte(hop.HtlcMinimumMsat >> (4 * 8))
	r[47] = byte(hop.HtlcMinimumMsat >> (3 * 8))
	r[48] = byte(hop.HtlcMinimumMsat >> (2 * 8))
	r[49] = byte(hop.HtlcMinimumMsat >> (1 * 8))
	r[50] = byte(hop.HtlcMinimumMsat >> (0 * 8))

	// feebase (4 bytes BE)
	r[51] = byte(hop.FeeBaseMsat >> (3 * 8))
	r[52] = byte(hop.FeeBaseMsat >> (2 * 8))
	r[53] = byte(hop.FeeBaseMsat >> (1 * 8))
	r[54] = byte(hop.FeeBaseMsat >> (0 * 8))

	// feeProportional (4 bytes BE)
	r[55] = byte(hop.FeeProportionalMillionths >> (3 * 8))
	r[56] = byte(hop.FeeProportionalMillionths >> (2 * 8))
	r[57] = byte(hop.FeeProportionalMillionths >> (1 * 8))
	r[58] = byte(hop.FeeProportionalMillionths >> (0 * 8))

	return r
}
