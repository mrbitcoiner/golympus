package main

import (
	"reflect"
	"testing"
)

func Test_clnDataToPaymentRoute(t *testing.T) {
	tData := clnDataToPaymentRoutePayloadData()

	r := clnDataToPaymentRoute(tData.route, tData.chans)

	if !reflect.DeepEqual(tData.expected, r) {
		t.Fatalf("expecting: %+v\ngot: %+v\n", tData.expected, r)
	}
}

type testPayload struct {
	route    clnRoute
	chans    map[int64][]clnChan
	expected PaymentRoute
}

func clnDataToPaymentRoutePayloadData() testPayload {

	routePayload := []clnHop{
		{
			NodeId:         "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ShortChannelId: "877236x1111x0",
			AmountMsat:     1003451,
			Delay:          253,
		},
		{
			NodeId:         "03cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			ShortChannelId: "877236x1112x0",
			AmountMsat:     1003000,
			Delay:          219,
		},
		{
			NodeId:         "03dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
			ShortChannelId: "877236x1113x0",
			AmountMsat:     1000000,
			Delay:          9,
		},
	}
	for i, v := range routePayload {
		routePayload[i].ShortChannelIdInt = mustShortChannelIdToInt(v.ShortChannelId)
	}

	channelsPayload := [][]clnChan{
		{
			{
				Source:          "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Destination:     "02bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				ShortChannelId:  "877236x1111x0",
				BaseFeeMsat:     0,
				FeePerMillionth: 450,
				Delay:           34,
				HtlcMinMsat:     "1msat",
				HtlcMaxMsat:     "2070588000msat",
			},
			{
				Source:          "02bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				Destination:     "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				ShortChannelId:  "877236x1111x0",
				BaseFeeMsat:     1000,
				FeePerMillionth: 100,
				Delay:           34,
				HtlcMinMsat:     "1msat",
				HtlcMaxMsat:     "2070588000msat",
			},
		},
		{
			{
				Source:          "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Destination:     "03cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				ShortChannelId:  "877236x1112x0",
				BaseFeeMsat:     0,
				FeePerMillionth: 450,
				Delay:           34,
				HtlcMinMsat:     "1msat",
				HtlcMaxMsat:     "6930000000msat",
			},
			{
				Source:          "03cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				Destination:     "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				ShortChannelId:  "877236x1112x0",
				BaseFeeMsat:     0,
				FeePerMillionth: 3000,
				Delay:           210,
				HtlcMinMsat:     "1000msat",
				HtlcMaxMsat:     "21049000msat",
			},
		},
		{
			{
				Source:          "03dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				Destination:     "03cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				ShortChannelId:  "877236x1113x0",
				BaseFeeMsat:     1000,
				FeePerMillionth: 1,
				Delay:           80,
				HtlcMinMsat:     "1000msat",
				HtlcMaxMsat:     "9900000000msat",
			},
			{
				Source:          "03cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				Destination:     "03dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				ShortChannelId:  "877236x1113x0",
				BaseFeeMsat:     0,
				FeePerMillionth: 3000,
				Delay:           210,
				HtlcMinMsat:     "1000msat",
				HtlcMaxMsat:     "8117000msat",
			},
		},
	}

	channelsMapPayload := map[int64][]clnChan{}
	for _, v := range channelsPayload {
		scid := mustShortChannelIdToInt(v[0].ShortChannelId)
		channelsMapPayload[scid] = v
	}

	expected := []Hop{
		{
			NodeId:                    "02bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			ShortChannelId:            964531182376517632,
			CltvExpiryDelta:           34,
			HtlcMinimumMsat:           1,
			FeeBaseMsat:               1000,
			FeeProportionalMillionths: 100,
		},
		{
			NodeId:                    "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ShortChannelId:            964531182376583168,
			CltvExpiryDelta:           34,
			HtlcMinimumMsat:           1,
			FeeBaseMsat:               0,
			FeeProportionalMillionths: 450,
		},
		{
			NodeId:                    "03cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			ShortChannelId:            964531182376648704,
			CltvExpiryDelta:           210,
			HtlcMinimumMsat:           1000,
			FeeBaseMsat:               0,
			FeeProportionalMillionths: 3000,
		},
	}

	return testPayload{
		route:    clnRoute{routePayload},
		chans:    channelsMapPayload,
		expected: expected,
	}
}
