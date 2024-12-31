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

	channelsData := make(map[int64][]clnChan, len(clnRoute.Hops))
	for _, v := range clnRoute.Hops {
		cd, err := lr.getChan(v.ShortChannelId)
		if err != nil {
			return nil, stackerr.Wrap(err)
		}
		channelsData[mustShortChannelIdToInt(cd[0].ShortChannelId)] = cd
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
		maxHops    = 5
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
	for i, hop := range r.Hops {
		r.Hops[i].ShortChannelIdInt = mustShortChannelIdToInt(hop.ShortChannelId)
	}

	return r, nil
}

func (lr *lnRouter) getChan(scid string) ([]clnChan, error) {
	var r struct {
		Channels []clnChan `json:"channels"`
	}
	params := struct {
		ShortChannelId string `json:"short_channel_id"`
	}{scid}

	err := lr.client.Call("listchannels", params, &r)
	if err != nil {
		return r.Channels, stackerr.Wrap(err)
	}

	if l := len(r.Channels); l != 2 {
		return r.Channels, fmt.Errorf("unexpected getchannels result length: %d", l)
	}

	return r.Channels, nil
}

func clnDataToPaymentRoute(
	clnRoute clnRoute, clnChans map[int64][]clnChan,
) PaymentRoute {

	hops := []Hop{}
	for i, clnHop := range clnRoute.Hops {
		var hop Hop

		// in the first hop we don't know previous node id (id of this node)
		// let's discover it
		if i == 0 &&
			clnChans[clnHop.ShortChannelIdInt][0].Destination == clnHop.NodeId {
			hop.NodeId = clnChans[clnHop.ShortChannelIdInt][0].Source
		} else if i == 0 {
			hop.NodeId = clnChans[clnHop.ShortChannelIdInt][0].Destination
		} else {
			// we are sure there's a previous hop
			hop.NodeId = clnRoute.Hops[i-1].NodeId
		}

		hop.ShortChannelId = clnHop.ShortChannelIdInt

		var sourceChan clnChan
		if clnChans[clnHop.ShortChannelIdInt][0].Source == hop.NodeId {
			sourceChan = clnChans[clnHop.ShortChannelIdInt][0]
		} else {
			sourceChan = clnChans[clnHop.ShortChannelIdInt][1]
		}

		hop.CltvExpiryDelta = int16(sourceChan.Delay)
		hop.HtlcMinimumMsat = msatWithSuffixToInt(sourceChan.HtlcMinMsat)
		hop.FeeBaseMsat = int32(sourceChan.BaseFeeMsat)
		hop.FeeProportionalMillionths = int32(sourceChan.FeePerMillionth)

		hops = append(hops, hop)
	}

	return hops
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
	CltvExpiryDelta           int16  `json:"cltvExpireDelta"`
	HtlcMinimumMsat           int64  `json:"htlcMinimumMsat"`
	FeeBaseMsat               int32  `json:"feeBaseMsat"`
	FeeProportionalMillionths int32  `json:"feeProportionalMillionths"`
}

type clnRoute struct {
	Hops []clnHop `json:"route"`
}

type clnHop struct {
	NodeId            string `json:"id"`
	ShortChannelId    string `json:"channel"`
	ShortChannelIdInt int64
	AmountMsat        int64 `json:"msatoshi"`
	Delay             int32 `json:"delay"`
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
