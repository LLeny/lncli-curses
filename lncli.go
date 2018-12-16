package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	expect "github.com/google/goexpect"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type lncliStatus struct {
	localNodeInfo      lnrpc.GetInfoResponse
	walletBalance      lnrpc.WalletBalanceResponse
	channels           []*lncliChannel
	peers              []*lncliPeer
	pendingchannels    lncliPendingChannelsContainer
	walletTransactions []*lnrpc.Transaction
	invoices           lncliInvoicesContainer
	payments           []*lncliPayment
	nodes              map[string]lnrpc.NodeInfo
}

type pendingChannelType int

const (
	closingChannel      pendingChannelType = 0
	forceClosingChannel pendingChannelType = 1
	openChannel         pendingChannelType = 2
	waitingCloseChannel pendingChannelType = 3
)

func makeNewPendingChannel(lnChan *lnrpc.PendingChannelsResponse_PendingChannel, t pendingChannelType) *lncliPendingChannel {
	c := new(lncliPendingChannel)
	c.PendingChannelsResponse_PendingChannel = *lnChan
	c.pendingType = t
	return c
}

type lncliPendingChannel struct {
	lnrpc.PendingChannelsResponse_PendingChannel
	pendingType        pendingChannelType
	NodeAlias          string
	closingTxid        string
	limboBalance       int64
	blocksTilMaturity  int32
	maturityHeight     uint32
	pendingHtlcs       []*lnrpc.PendingHTLC
	recoveredBalance   int64
	commitFee          int64
	commitWeight       int64
	confirmationHeight uint32
	feePerKw           int64
}

func (c *lncliPendingChannel) GetType() string {
	switch c.pendingType {
	case openChannel:
		return "O"
	case waitingCloseChannel:
		return "W"
	case forceClosingChannel:
		return "F"
	case closingChannel:
		return "C"
	}
	return " "
}

type lncliPendingChannelsContainer struct {
	totalLimbo      int64
	pendingChannels []*lncliPendingChannel
}

type lncliChannel struct {
	lnrpc.Channel
	NodeAlias string
}

type lncliPeer struct {
	lnrpc.Peer
	Alias string
}

type lncliInvoicesContainer struct {
	currentStartIndex int64
	invoices          []*lncliInvoice
}

type lncliInvoice struct {
	lnrpc.Invoice
}

type lncliPayment struct {
	lnrpc.Payment
}

func (c *lncliPendingChannel) updateNodeAlias(ctxt *lnclicursesContext, stat *lncliStatus) error {
	ni, err := stat.getNodeInfo(ctxt, c.GetRemoteNodePub())
	if err != nil {
		return err
	}
	c.NodeAlias = ni.Node.Alias
	return nil
}

func (c *lncliChannel) updateNodeAlias(ctxt *lnclicursesContext, stat *lncliStatus) error {
	ni, err := stat.getNodeInfo(ctxt, c.GetRemotePubkey())
	if err != nil {
		return err
	}
	c.NodeAlias = ni.Node.Alias
	return nil
}

func (c *lncliPeer) updateNodeAlias(ctxt *lnclicursesContext, stat *lncliStatus) error {
	ni, err := stat.getNodeInfo(ctxt, c.GetPubKey())
	if err != nil {
		return err
	}
	c.Alias = ni.Node.Alias
	return nil
}

func (s *lncliStatus) getNodeInfo(ctxt *lnclicursesContext, pubkey string) (lnrpc.NodeInfo, error) {
	found, ok := s.nodes[pubkey]

	if ok {
		return found, nil
	}

	nodeinfo, err := s.queryNodeInfo(ctxt, pubkey)

	if err != nil {
		return nodeinfo, err
	}

	s.nodes[pubkey] = nodeinfo
	return nodeinfo, nil
}

func (ctxt *lnclicursesContext) getlncliArgs() []string {
	var args []string

	args = nil

	if len(ctxt.opts.WorkDir) > 0 {
		args = append(args, "--lnddir="+ctxt.opts.WorkDir)
	}
	if len(ctxt.opts.RPCServer) > 0 {
		args = append(args, "--rpcserver="+ctxt.opts.RPCServer)
	}
	if len(ctxt.opts.TLSCertPath) > 0 {
		args = append(args, "--tlscertpath="+ctxt.opts.TLSCertPath)
	}
	if ctxt.opts.NoMacaroons {
		args = append(args, "--no-macaroons")
	}
	if len(ctxt.opts.MacaroonPath) > 0 {
		args = append(args, "--macaroonpath="+ctxt.opts.MacaroonPath)
	}
	if ctxt.opts.MacaroonTimeOut > 0 {
		args = append(args, fmt.Sprintf("--macaroontimeout=%d", ctxt.opts.MacaroonTimeOut))
	}
	if len(ctxt.opts.MacaroonIP) > 0 {
		args = append(args, "--macaroonip="+ctxt.opts.MacaroonIP)
	}

	return args
}

func (ctxt *lnclicursesContext) execlncliCommand(command string) ([]byte, error) {

	args := ctxt.getlncliArgs()
	args = append(args, strings.Split(command, " ")...)

	cmd := exec.Command(ctxt.opts.LncliExec)
	cmd.Args = args

	ctxt.cliMutex.Lock()
	out, err := cmd.Output()
	ctxt.cliMutex.Unlock()

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *lncliStatus) updateLocalNodeInfo(ctxt *lnclicursesContext) error {
	txt, err := ctxt.execlncliCommand("getinfo")
	if err != nil {
		return err
	}
	return jsonpb.Unmarshal(bytes.NewReader(txt), &s.localNodeInfo)
}

func (s *lncliStatus) queryNodeInfo(ctxt *lnclicursesContext, pubkey string) (lnrpc.NodeInfo, error) {
	txt, err := ctxt.execlncliCommand("getnodeinfo " + pubkey)
	var nodeinfo lnrpc.NodeInfo
	if err != nil {
		return nodeinfo, err
	}
	err = jsonpb.Unmarshal(bytes.NewReader(txt), &nodeinfo)
	return nodeinfo, err
}

func (s *lncliStatus) updateWalletBalance(ctxt *lnclicursesContext) error {
	txt, err := ctxt.execlncliCommand("walletbalance")
	if err != nil {
		return err
	}
	return jsonpb.Unmarshal(bytes.NewReader(txt), &s.walletBalance)
}

func (s *lncliStatus) walletNewAdress(ctxt *lnclicursesContext, addressType string) (string, error) {
	txt, err := ctxt.execlncliCommand("newaddress " + addressType)
	if err != nil {
		return "", err
	}
	var na lnrpc.NewAddressResponse
	if err = jsonpb.Unmarshal(bytes.NewReader(txt), &na); err != nil {
		return "", err
	}
	return na.Address, nil
}

func (s *lncliStatus) updateWallletTransactionsList(ctxt *lnclicursesContext) error {
	txt, err := ctxt.execlncliCommand("listchaintxns")
	if err != nil {
		return err
	}
	var trans lnrpc.TransactionDetails
	if err = jsonpb.Unmarshal(bytes.NewReader(txt), &trans); err != nil {
		return err
	}

	s.walletTransactions = nil

	for _, c := range trans.Transactions {
		s.walletTransactions = append(s.walletTransactions, c)
	}

	ctxt.views[walletTransactionsViewt].getGrid().items = s.walletTransactions
	return nil
}

func (s *lncliStatus) updateInvoiceList(ctxt *lnclicursesContext) error {

	txt, err := ctxt.execlncliCommand("listinvoices --reversed --max_invoices 100 --index_offset " + strconv.FormatInt(s.invoices.currentStartIndex, 10))
	if err != nil {
		return err
	}
	var chans lnrpc.ListInvoiceResponse
	if err = jsonpb.Unmarshal(bytes.NewReader(txt), &chans); err != nil {
		return err
	}

	s.invoices.invoices = nil

	for _, c := range chans.Invoices {
		nc := lncliInvoice{*c}
		s.invoices.invoices = append(s.invoices.invoices, &nc)
		if err != nil {
			logError(err.Error())
		}
	}

	ctxt.views[invoiceListViewt].getGrid().items = s.invoices.invoices

	return nil
}

func (s *lncliStatus) updatePaymentList(ctxt *lnclicursesContext) error {

	txt, err := ctxt.execlncliCommand("listpayments")
	if err != nil {
		return err
	}
	var chans lnrpc.ListPaymentsResponse
	if err = jsonpb.Unmarshal(bytes.NewReader(txt), &chans); err != nil {
		return err
	}

	s.payments = nil

	for _, c := range chans.Payments {
		nc := lncliPayment{*c}
		s.payments = append(s.payments, &nc)
		if err != nil {
			logError(err.Error())
		}
	}

	ctxt.views[paymentListViewt].getGrid().items = s.payments

	return nil
}

func (s *lncliStatus) updateChannelList(ctxt *lnclicursesContext) error {

	txt, err := ctxt.execlncliCommand("listchannels")
	if err != nil {
		return err
	}
	var chans lnrpc.ListChannelsResponse
	if err = jsonpb.Unmarshal(bytes.NewReader(txt), &chans); err != nil {
		return err
	}

	s.channels = nil

	for _, c := range chans.Channels {
		nc := lncliChannel{*c, ""}
		s.channels = append(s.channels, &nc)
		go func() {
			err := nc.updateNodeAlias(ctxt, s)
			if err != nil {
				logError(err.Error())
			}
			refreshView()
		}()
	}

	ctxt.views[channelListViewt].getGrid().items = s.channels

	return nil
}

func (s *lncliStatus) updatePendingChannelList(ctxt *lnclicursesContext) error {

	txt, err := ctxt.execlncliCommand("pendingchannels")
	if err != nil {
		return err
	}
	var chans lnrpc.PendingChannelsResponse
	if err = jsonpb.Unmarshal(bytes.NewReader(txt), &chans); err != nil {
		return err
	}

	s.pendingchannels.pendingChannels = nil
	s.pendingchannels.totalLimbo = chans.GetTotalLimboBalance()

	for _, c := range chans.PendingClosingChannels {
		nc := makeNewPendingChannel(c.GetChannel(), closingChannel)
		nc.closingTxid = c.GetClosingTxid()
		s.pendingchannels.pendingChannels = append(s.pendingchannels.pendingChannels, nc)
		go func() {
			err := nc.updateNodeAlias(ctxt, s)
			if err != nil {
				logError(err.Error())
			}
			refreshView()
		}()
	}

	for _, c := range chans.PendingForceClosingChannels {
		nc := makeNewPendingChannel(c.GetChannel(), forceClosingChannel)
		nc.closingTxid = c.GetClosingTxid()
		nc.blocksTilMaturity = c.GetBlocksTilMaturity()
		nc.limboBalance = c.GetLimboBalance()
		nc.maturityHeight = c.GetMaturityHeight()
		nc.pendingHtlcs = c.GetPendingHtlcs()
		nc.recoveredBalance = c.GetRecoveredBalance()
		s.pendingchannels.pendingChannels = append(s.pendingchannels.pendingChannels, nc)
		go func() {
			err := nc.updateNodeAlias(ctxt, s)
			if err != nil {
				logError(err.Error())
			}
			refreshView()
		}()
	}

	for _, c := range chans.PendingOpenChannels {
		nc := makeNewPendingChannel(c.GetChannel(), openChannel)
		nc.commitFee = c.GetCommitFee()
		nc.commitWeight = c.GetCommitWeight()
		nc.confirmationHeight = c.GetConfirmationHeight()
		nc.feePerKw = c.GetFeePerKw()
		s.pendingchannels.pendingChannels = append(s.pendingchannels.pendingChannels, nc)
		go func() {
			err := nc.updateNodeAlias(ctxt, s)
			if err != nil {
				logError(err.Error())
			}
			refreshView()
		}()
	}

	for _, c := range chans.WaitingCloseChannels {
		nc := makeNewPendingChannel(c.GetChannel(), waitingCloseChannel)
		nc.limboBalance = c.GetLimboBalance()
		s.pendingchannels.pendingChannels = append(s.pendingchannels.pendingChannels, nc)
		go func() {
			err := nc.updateNodeAlias(ctxt, s)
			if err != nil {
				logError(err.Error())
			}
			refreshView()
		}()
	}

	ctxt.views[pendingChannelListViewt].getGrid().items = s.pendingchannels.pendingChannels

	return nil
}

func (s *lncliStatus) closeChannel(ctxt *lnclicursesContext, channel *lncliChannel, force bool) (string, error) {
	cp := strings.Split(channel.ChannelPoint, ":")
	cm := "closechannel " + cp[0]
	if force {
		cm = cm + " --force"
	}
	out, err := ctxt.execlncliCommand(cm)

	if err != nil {
		return "", err
	}

	return getAttributeStr(out, "closing_txid")
}

func (s *lncliStatus) connectToPeer(ctxt *lnclicursesContext, pubkey string, host string, port int) (string, error) {
	cm := fmt.Sprintf("connect %s@%s:%d", pubkey, host, port)
	o, err := ctxt.execlncliCommand(cm)
	if err != nil {
		return "", err
	}
	return string(o), nil
}

func (s *lncliStatus) disconnectPeer(ctxt *lnclicursesContext, peer *lncliPeer) (string, error) {
	cm := fmt.Sprintf("disconnect %s", peer.PubKey)
	o, err := ctxt.execlncliCommand(cm)
	if err != nil {
		return "", err
	}
	return string(o), nil
}

func (s *lncliStatus) updatePeersList(ctxt *lnclicursesContext) error {

	txt, err := ctxt.execlncliCommand("listpeers")
	if err != nil {
		return err
	}
	var peers lnrpc.ListPeersResponse
	if err = jsonpb.Unmarshal(bytes.NewReader(txt), &peers); err != nil {
		return err
	}

	s.peers = nil

	for _, p := range peers.Peers {
		np := lncliPeer{*p, ""}
		s.peers = append(s.peers, &np)
		go func() {
			err := np.updateNodeAlias(ctxt, s)
			if err != nil {
				logError(err.Error())
			}
			refreshView()
		}()
	}

	ctxt.views[peerListViewt].getGrid().items = s.peers

	return nil
}

func (s *lncliStatus) payInvoice(ctxt *lnclicursesContext, payReq string, amt int, feeLimit int, feeLimitPerc int, force bool) (string, error) {
	var args []string
	args = nil

	args = append(args, ctxt.opts.LncliExec)
	args = append(args, ctxt.getlncliArgs()...)
	args = append(args, "payinvoice")

	if len(payReq) > 0 {
		args = append(args, "--pay_req")
		args = append(args, payReq)
	}

	if amt > 0 {
		args = append(args, "--amt")
		args = append(args, strconv.Itoa(amt))
	}

	if feeLimit > 0 {
		args = append(args, "--fee_limit")
		args = append(args, strconv.Itoa(feeLimit))
	}

	if feeLimitPerc > 0 {
		args = append(args, "--fee_limit_percent")
		args = append(args, strconv.Itoa(feeLimitPerc))
	}

	if force {
		args = append(args, "--force")
	}

	ctxt.cliMutex.Lock()

	ex, _, err := expect.Spawn(strings.Join(args, " "), -1)

	if err != nil {
		ctxt.cliMutex.Unlock()
		return "", err
	}

	s1, _, err := ex.Expect(regexp.MustCompile("\\/no\\):"), 10*time.Second)

	if err != nil {
		logError(err.Error())
		ctxt.cliMutex.Unlock()
		return "", err
	}

	go displayMessage(s1, func(valid bool) {
		if !valid {
			ex.Close()
			ctxt.cliMutex.Unlock()
			return
		}

		ex.Send("yes\n")

		out, _, err := ex.Expect(regexp.MustCompile("(confirmed)|(})"), 120*time.Second)

		if err != nil {
			ex.Close()
			ctxt.cliMutex.Unlock()
			logError(err.Error())
			return
		}

		re, _ := regexp.Compile("\"payment_error\": \".+\"")

		errFound := re.FindString(out)

		ctxt.cliMutex.Unlock()
		ex.Close()

		var finalMsg string

		if len(errFound) > 0 {
			finalMsg = strings.Replace(strings.Replace(errFound, "\"", "", -1), "payment_error: ", "", -1)
		} else {
			finalMsg = "Success"
		}

		displayMessage(finalMsg, func(valid bool) {
			updateData()
		})
	})

	return "", nil
}

func (s *lncliStatus) openChannel(ctxt *lnclicursesContext, nk string, cct string, lamnt int, pamnt int, pri bool, blk bool, mcf int, cftgt int, spb int, minht int, remcsv int) (string, error) {

	var args []string

	if len(nk) > 0 {
		args = append(args, "--node_key")
		args = append(args, nk)
	}

	if len(cct) > 0 {
		args = append(args, "--connect")
		args = append(args, cct)
	}

	if lamnt > 0 {
		args = append(args, "--local_amt")
		args = append(args, strconv.Itoa(lamnt))
	}

	if pamnt > 0 {
		args = append(args, "--push_amt")
		args = append(args, strconv.Itoa(pamnt))
	}

	if blk {
		args = append(args, "--block")
	}

	if pri {
		args = append(args, "--private")
	}

	if cftgt > 0 {
		args = append(args, "--conf_target")
		args = append(args, strconv.Itoa(cftgt))
	}

	if spb > 0 {
		args = append(args, "--sat_per_byte")
		args = append(args, strconv.Itoa(spb))
	}

	if minht > 0 {
		args = append(args, "--min_htlc_msat")
		args = append(args, strconv.Itoa(minht))
	}

	if remcsv > 0 {
		args = append(args, "--remote_csv_delay")
		args = append(args, strconv.Itoa(remcsv))
	}

	if mcf > 0 {
		args = append(args, "--min_confs")
		args = append(args, strconv.Itoa(mcf))
	}

	cm := fmt.Sprintf("openchannel %s", strings.Join(args, " "))
	out, err := ctxt.execlncliCommand(cm)

	if err != nil {
		return "", err
	}

	return getAttributeStr(out, "funding_txid")
}

func getAttributeStr(in []byte, attributeName string) (string, error) {
	var attrs map[string]string
	if err := json.Unmarshal(in, &attrs); err != nil {
		return "", err
	}
	if val, ok := attrs[attributeName]; ok {
		return val, nil
	}
	return "", errors.New("json attribute not found")
}

func (s *lncliStatus) addInvoice(ctxt *lnclicursesContext, amt int, deshash string, expiry int, fallbackaddr string, memo string, preimage string, private bool, receipt string) (string, error) {

	var args []string

	if len(memo) > 0 {
		args = append(args, "--memo")
		args = append(args, memo)
	}

	if len(receipt) > 0 {
		args = append(args, "--receipt")
		args = append(args, receipt)
	}

	if len(preimage) > 0 {
		args = append(args, "--preimage")
		args = append(args, preimage)
	}

	if amt > 0 {
		args = append(args, "--amt")
		args = append(args, strconv.Itoa(amt))
	}

	if len(deshash) > 0 {
		args = append(args, "--description_hash")
		args = append(args, deshash)
	}

	if len(fallbackaddr) > 0 {
		args = append(args, "--fallback_addr")
		args = append(args, fallbackaddr)
	}

	if expiry > 0 {
		args = append(args, "--expiry")
		args = append(args, strconv.Itoa(expiry))
	}

	if private {
		args = append(args, "--private")
	}

	cm := fmt.Sprintf("addinvoice %s", strings.Join(args, " "))
	out, err := ctxt.execlncliCommand(cm)

	if err != nil {
		return "", err
	}

	return getAttributeStr(out, "pay_req")
}
