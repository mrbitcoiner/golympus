package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"master.private/bstd.git/jsonrpc"
	"master.private/bstd.git/stackerr"
)

type lnRouter struct {
	client *jsonrpc.Client
}

func NewLnRouter(network, address string) *lnRouter {
	conn, err := net.Dial(network, address)
	if err != nil {
		panic(stackerr.Wrap(err))
	}
	return &lnRouter{
		client: jsonrpc.NewClient(conn),
	}
}

func (lr *lnRouter) FindRoutes(
	fromPubkeys []string, toPubkey string, msat int64,
) ([]PaymentRoute, error) {

	clnRoute, err := lr.getRoute(toPubkey, msat)
	if err != nil {
		return nil, stackerr.Wrap(err)
	}

	channelsData := make([]clnChan, 0, len(clnRoute.Hops))
	for _, v := range clnRoute.Hops {
		cd, err := lr.getChan(v.ShortChannelId, v.NodeId)
		if err != nil {
			return nil, stackerr.Wrap(err)
		}
		channelsData = append(channelsData, cd)
	}

	paymentRoute := clnDataToPaymentRoute(clnRoute, channelsData)

	return []PaymentRoute{paymentRoute}, nil
}

func (lr *lnRouter) Close() error {
	err := lr.client.Close()
	if err != nil {
		return stackerr.Wrap(err)
	}
	return nil
}

func (lr *lnRouter) getRoute(toPubkey string, msat int64) (clnRoute, error) {
	const (
		riskFactor = 0
		maxHops    = 3
	)
	var r clnRoute
	params := struct {
		ToPubkey   string `json:"id"`
		AmountMsat int64  `json:"msatoshi"`
		RiskFactor int64  `json:"riskfactor"`
		MaxHops    int64  `json:"maxhops"`
	}{
		toPubkey, msat, riskFactor, maxHops,
	}

	err := lr.client.Call("getroute", params, &r)
	if err != nil {
		return r, stackerr.Wrap(err)
	}

	return r, nil
}

func (lr *lnRouter) getChan(scid string, destNodeId string) (clnChan, error) {
	var r clnChan
	var result struct {
		Channels []clnChan `json:"channels"`
	}
	params := struct {
		ShortChannelId string `json:"short_channel_id"`
	}{scid}

	err := lr.client.Call("listchannels", params, &result)
	if err != nil {
		return r, stackerr.Wrap(err)
	}

	if l := len(result.Channels); l != 2 {
		return r, fmt.Errorf("unexpected getchannels result length: %d", l)
	}

	if result.Channels[0].Destination == destNodeId {
		r = result.Channels[0]
	} else {
		r = result.Channels[1]
	}

	return r, nil
}

func clnDataToPaymentRoute(clnRoute clnRoute, clnChans []clnChan) PaymentRoute {
	hops := make([]Hop, 0, len(clnRoute.Hops))
	for i, v := range clnRoute.Hops {
		scidAsInt, err := shortChannelIdToInt(v.ShortChannelId)
		if err != nil {
			// must never fail, on failure, we want a crash
			panic(stackerr.Wrap(err))
		}
		hop := Hop{
			NodeId:                    v.NodeId,
			ShortChannelId:            scidAsInt,
			CltvExpireDelta:           clnChans[i].Delay,
			HtlcMinimumMsat:           msatWithSuffixToInt(clnChans[i].HtlcMinMsat),
			FeeBaseMsat:               clnChans[i].BaseFeeMsat,
			FeeProportionalMillionths: clnChans[i].FeePerMillionth,
		}
		hops = append(hops, hop)
	}
	return PaymentRoute(hops)
}

func msatWithSuffixToInt(in string) int64 {
	s, _ := strings.CutSuffix(in, "msat")
	r, err := strconv.ParseInt(s, 10, 64)
	// must never fail
	if err != nil {
		panic(stackerr.Wrap(err))
	}
	return r
}

type PaymentRoute []Hop

type Hop struct {
	NodeId                    string `json:"nodeId"`
	ShortChannelId            int64  `json:"shortChannelId"`
	CltvExpireDelta           int32  `json:"cltvExpireDelta"`
	HtlcMinimumMsat           int64  `json:"htlcMinimumMsat"`
	FeeBaseMsat               int64  `json:"feeBaseMsat"`
	FeeProportionalMillionths int64  `json:"feeProportionalMillionths"`
}

type clnRoute struct {
	Hops []clnHop `json:"route"`
}

type clnHop struct {
	NodeId         string `json:"id"`
	ShortChannelId string `json:"channel"`
	AmountMsat     int64  `json:"msatoshi"`
	Cltv           int32  `json:"delay"`
}

type clnChan struct {
	Source          string `json:"source"`
	Destination     string `json:"destination"`
	ShortChannelId  string `json:"short_channel_id"`
	BaseFeeMsat     int64  `json:"base_fee_millisatoshi"`
	FeePerMillionth int64  `json:"fee_per_millionth"`
	Delay           int32  `json:"delay"`
	HtlcMinMsat     string `json:"htlc_minimum_msat"`
	HtlcMaxMsat     string `json:"htlc_maximum_msat"`
}
