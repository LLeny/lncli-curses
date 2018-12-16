package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type pendingchannelListView struct {
	viewBase
	form *formEdit
}

type pendingOpenChannelDisplayContainer struct {
	NodeAlias          string `displayname:"Node alias" length:"32" readonly:"1"`
	ChannelPoint       string `displayname:"Channel point" length:"32" readonly:"1" lines:"2"`
	Capacity           string `displayname:"Capacity" length:"16" readonly:"1"`
	LocalBalance       string `displayname:"Local balance" length:"16" readonly:"1"`
	RemoteBalance      string `displayname:"Remote balance" length:"16" readonly:"1"`
	CommitFee          string `displayname:"Commit Fee" length:"16" readonly:"1"`
	CommitWeight       string `displayname:"Commit Weight" length:"16" readonly:"1"`
	FeePerKw           string `displayname:"Fee per Kw" length:"16" readonly:"1"`
	ConfirmationHeight string `displayname:"Confirmation Height" length:"16" readonly:"1"`
}

type pendingClosingChannelDisplayContainer struct {
	NodeAlias     string `displayname:"Node alias" length:"32" readonly:"1"`
	ChannelPoint  string `displayname:"Channel point" length:"32" readonly:"1" lines:"2"`
	Capacity      string `displayname:"Capacity" length:"16" readonly:"1"`
	LocalBalance  string `displayname:"Local balance" length:"16" readonly:"1"`
	RemoteBalance string `displayname:"Remote balance" length:"16" readonly:"1"`
	ClosingTxid   string `displayname:"Closing Txid" length:"32" readonly:"1"`
}

type pendingForceClosingChannelDisplayContainer struct {
	NodeAlias         string `displayname:"Node alias" length:"32" readonly:"1"`
	ChannelPoint      string `displayname:"Channel point" length:"32" readonly:"1" lines:"2"`
	Capacity          string `displayname:"Capacity" length:"16" readonly:"1"`
	LocalBalance      string `displayname:"Local balance" length:"16" readonly:"1"`
	RemoteBalance     string `displayname:"Remote balance" length:"16" readonly:"1"`
	LimboBalance      string `displayname:"Limbo balance" length:"16" readonly:"1"`
	RecoveredBalance  string `displayname:"Recovered balance" length:"16" readonly:"1"`
	MaturityHeight    string `displayname:"Maturity height" length:"16" readonly:"1"`
	BlocksTilMaturity string `displayname:"Blocks until maturity" length:"16" readonly:"1"`
	PendingHtlcs      string `displayname:"Pending HTLCs" length:"16" readonly:"1"`
}

type pendingWaitingOpenChannelDisplayContainer struct {
	NodeAlias     string `displayname:"Node alias" length:"32" readonly:"1"`
	ChannelPoint  string `displayname:"Channel point" length:"32" readonly:"1" lines:"2"`
	Capacity      string `displayname:"Capacity" length:"16" readonly:"1"`
	LocalBalance  string `displayname:"Local balance" length:"16" readonly:"1"`
	RemoteBalance string `displayname:"Remote balance" length:"16" readonly:"1"`
	LimboBalance  string `displayname:"Limbo balance" length:"16" readonly:"1"`
}

func newpendingchannelListView(physicalView string, fmtnormal string, fmtheader string, fmtselected string) *pendingchannelListView {
	cv := new(pendingchannelListView)

	cv.grid = &dataGrid{}
	cv.mappedToPhysicalView = physicalView

	cv.init(fmtnormal, fmtheader, fmtselected)

	return cv
}

func (cv *pendingchannelListView) init(fmtnormal string, fmtheader string, fmtselected string) {

	cv.grid.fmtForeground = fmtnormal
	cv.grid.fmtHeader = fmtheader
	cv.grid.fmtSelected = fmtselected

	cv.shortcuts = nil
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Up", "Up", gocui.KeyArrowUp, gocui.ModNone, func() { cv.grid.moveSelectionUp() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Down", "Down", gocui.KeyArrowDown, gocui.ModNone, func() { cv.grid.moveSelectionDown() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Details channel", "D", 'd', gocui.ModAlt, cv.detailsChannel, true, ""})

	cv.grid.header = "[Pending channels]"
	cv.grid.addColumn("T", "GetType", 2, stringRow)
	cv.grid.addColumn("Node", "NodeAlias", 0, stringRow)
	cv.grid.addColumn("Capacity", "GetCapacity", 10, intRow)
	cv.grid.addColumn("Local", "GetLocalBalance", 10, intRow)
	cv.grid.addColumn("Remote", "GetRemoteBalance", 10, intRow)
}

func (cv *pendingchannelListView) detailsChannel() {
	c := cv.getSelectedChannel()

	if c == nil {
		return
	}

	switch c.pendingType {
	case openChannel:
		cv.detailsOpenChannel(c)
	case closingChannel:
		cv.detailsClosingChannel(c)
	case waitingCloseChannel:
		cv.detailswaitingCloseChannel(c)
	case forceClosingChannel:
		cv.detailsForceClosingChannel(c)
	}
}

func (cv *pendingchannelListView) detailswaitingCloseChannel(c *lncliPendingChannel) {
	cc := new(pendingWaitingOpenChannelDisplayContainer)

	cc.ChannelPoint = c.GetChannelPoint()
	cc.NodeAlias = c.NodeAlias
	cc.LocalBalance = context.printer.Sprintf("%d", c.GetLocalBalance())
	cc.RemoteBalance = context.printer.Sprintf("%d", c.GetRemoteBalance())
	cc.Capacity = context.printer.Sprintf("%d", c.GetCapacity())
	cc.LimboBalance = context.printer.Sprintf("%d", c.limboBalance)

	cv.form = newFormEdit("pendingwaitingcloseChanVal", "Waiting Close Channel", cc)

	cv.form.callback = func(valid bool) {
		cv.form.close(context.gocui)
		cv.form = nil
		cc = nil
	}

	cv.form.initialize(context.gocui)
	cv.form.switchActiveEditor(-1, context.gocui)
}

func (cv *pendingchannelListView) detailsForceClosingChannel(c *lncliPendingChannel) {
	cc := new(pendingForceClosingChannelDisplayContainer)

	cc.ChannelPoint = c.GetChannelPoint()
	cc.NodeAlias = c.NodeAlias
	cc.LocalBalance = context.printer.Sprintf("%d", c.GetLocalBalance())
	cc.RemoteBalance = context.printer.Sprintf("%d", c.GetRemoteBalance())
	cc.Capacity = context.printer.Sprintf("%d", c.GetCapacity())
	cc.BlocksTilMaturity = context.printer.Sprintf("%d", c.blocksTilMaturity)
	cc.LimboBalance = context.printer.Sprintf("%d", c.limboBalance)
	cc.MaturityHeight = context.printer.Sprintf("%d", c.maturityHeight)
	cc.PendingHtlcs = "" //p.Sprintf("%d", c.GetCapacity())
	cc.RecoveredBalance = context.printer.Sprintf("%d", c.recoveredBalance)

	cv.form = newFormEdit("pendingforcecloseChanVal", "Pending Force Closing Channel", cc)

	cv.form.callback = func(valid bool) {
		cv.form.close(context.gocui)
		cv.form = nil
		cc = nil
	}

	cv.form.initialize(context.gocui)
	cv.form.switchActiveEditor(-1, context.gocui)
}

func (cv *pendingchannelListView) detailsClosingChannel(c *lncliPendingChannel) {
	cc := new(pendingClosingChannelDisplayContainer)

	cc.ChannelPoint = c.GetChannelPoint()
	cc.NodeAlias = c.NodeAlias
	cc.LocalBalance = context.printer.Sprintf("%d", c.GetLocalBalance())
	cc.RemoteBalance = context.printer.Sprintf("%d", c.GetRemoteBalance())
	cc.Capacity = context.printer.Sprintf("%d", c.GetCapacity())
	cc.ClosingTxid = c.closingTxid

	cv.form = newFormEdit("pendingcloseChanVal", "Pending Closing Channel", cc)

	cv.form.callback = func(valid bool) {
		cv.form.close(context.gocui)
		cv.form = nil
		cc = nil
	}

	cv.form.initialize(context.gocui)
	cv.form.switchActiveEditor(-1, context.gocui)
}

func (cv *pendingchannelListView) detailsOpenChannel(c *lncliPendingChannel) {
	cc := new(pendingOpenChannelDisplayContainer)

	cc.ChannelPoint = c.GetChannelPoint()
	cc.NodeAlias = c.NodeAlias
	cc.LocalBalance = context.printer.Sprintf("%d", c.GetLocalBalance())
	cc.RemoteBalance = context.printer.Sprintf("%d", c.GetRemoteBalance())
	cc.Capacity = context.printer.Sprintf("%d", c.GetCapacity())
	cc.CommitFee = context.printer.Sprintf("%d", c.commitFee)
	cc.CommitWeight = context.printer.Sprintf("%d", c.commitWeight)
	cc.ConfirmationHeight = context.printer.Sprintf("%d", c.confirmationHeight)
	cc.FeePerKw = context.printer.Sprintf("%d", c.feePerKw)

	cv.form = newFormEdit("pendingopenChanVal", "Pending Open Channel", cc)

	cv.form.callback = func(valid bool) {
		cv.form.close(context.gocui)
		cv.form = nil
		cc = nil
	}

	cv.form.initialize(context.gocui)
	cv.form.switchActiveEditor(-1, context.gocui)
}

func (cv *pendingchannelListView) getSelectedChannel() *lncliPendingChannel {
	return cv.grid.getSelectedItem().Interface().(*lncliPendingChannel)
}

func (cv *pendingchannelListView) refreshView(g *gocui.Gui) {
	v, err := g.View(cv.mappedToPhysicalView)
	if err != nil {
		log.Panicln(err.Error())
		return
	}
	v.Clear()

	x, y := v.Size()

	cv.grid.setRenderSize(x, y)

	for _, row := range cv.grid.getGridRows() {
		fmt.Fprintln(v, row)
	}

	if cv.form != nil {
		cv.form.layout(g)
	}
}

func (cv *pendingchannelListView) getShortCuts() []*keyHandle {
	return cv.shortcuts
}

func (cv *pendingchannelListView) getGrid() *dataGrid {
	return cv.grid
}

func (cv *pendingchannelListView) getPhysicalView() string {
	return cv.mappedToPhysicalView
}
