package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/jroimartin/gocui"
)

type channelListView struct {
	viewBase
	form *formEdit
}

type closeChannelContainer struct {
	NodeAlias     string `displayname:"Node alias" length:"32" readonly:"1"`
	LocalBalance  string `displayname:"Local balance" length:"16" readonly:"1"`
	RemoteBalance string `displayname:"Remote balance" length:"16" readonly:"1"`
	ChannelPoint  string `displayname:"Channel point" length:"32" readonly:"1" lines:"2"`
	Force         bool   `displayname:"Force close"`
}

type openChannelContainer struct {
	NodeKey        string `displayname:"Node public key" length:"72" lines:"1"`
	Connect        string `displayname:"Host:port (opt)" length:"22"`
	LocalAmt       int    `displayname:"Local amount" length:"12"`
	PushAmt        int    `displayname:"Push amount" length:"12"`
	Private        bool   `displayname:"Private" length:"1"`
	Block          bool   `displayname:"Block and wait" length:"1"`
	MinConfs       int    `displayname:"Min confs (opt)" length:"12"`
	ConfTarget     int    `displayname:"Conf target (opt)" length:"12"`
	SatPerByte     int    `displayname:"Sat per byte (opt)" length:"12"`
	MinHtlcmSat    int    `displayname:"Min htlc mSat (opt)" length:"12"`
	RemoteCsvDelay int    `displayname:"Remote csv delay (opt)" length:"12"`
}

func newchannelListView(physicalView string, fmtnormal string, fmtheader string, fmtselected string) *channelListView {
	cv := new(channelListView)

	cv.grid = makeNewDataGrid()
	cv.mappedToPhysicalView = physicalView

	cv.init(fmtnormal, fmtheader, fmtselected)

	return cv
}

func (cv *channelListView) init(fmtnormal string, fmtheader string, fmtselected string) {

	cv.grid.fmtForeground = fmtnormal
	cv.grid.fmtHeader = fmtheader
	cv.grid.fmtSelected = fmtselected

	cv.shortcuts = nil
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Up", "Up", gocui.KeyArrowUp, gocui.ModNone, func() { cv.grid.moveSelectionUp() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Down", "Down", gocui.KeyArrowDown, gocui.ModNone, func() { cv.grid.moveSelectionDown() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Close channel", "C", 'c', gocui.ModAlt, cv.closeChannel, true, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Open channel", "O", 'o', gocui.ModAlt, cv.openChannel, true, ""})

	cv.grid.key = "channels"
	cv.grid.addColumn("Active", "Active", boolRow)               //Active, 2
	cv.grid.addColumn("Private", "Private", boolRow)             //Private, 2
	cv.grid.addColumn("Node", "NodeAlias", stringRow)            //Node, 0
	cv.grid.addColumn("Capacity", "Capacity", intRow)            //Capacity, 13
	cv.grid.addColumn("Local", "LocalBalance", intRow)           //Local, 13
	cv.grid.addColumn("Remote", "RemoteBalance", intRow)         //Remote, 13
	cv.grid.addColumn("ComFee", "CommitFee", intRow)             //Com. fee, 9
	cv.grid.addColumn("ComWeight", "CommitWeight", intRow)       //Com. weight, 12
	cv.grid.addColumn("FeeKw", "FeePerKw", intRow)               //Fee/Kw, 7
	cv.grid.addColumn("Unsettled", "UnsettledBalance", intRow)   //Unsettled, 13
	cv.grid.addColumn("TotSent", "TotalSatoshisSent", intRow)    //Tot. sent, 13
	cv.grid.addColumn("TotRec", "TotalSatoshisReceived", intRow) //Tot. rec., 13
	cv.grid.initConfig()
}

func (cv *channelListView) getSelectedChannel() *lncliChannel {
	return cv.grid.getSelectedItem().Interface().(*lncliChannel)
}

func (cv *channelListView) openChannel() {
	cc := new(openChannelContainer)

	cc.LocalAmt = 0
	cc.PushAmt = 0
	cc.MinConfs = 1

	cv.form = newFormEdit("openChanVal", "Open channel", cc)

	cv.form.callback = func(valid bool) {
		cv.form.getValue()
		cv.form.close(context.gocui)
		cv.form = nil
		if valid {
			txid, err := status.openChannel(&context, cc.NodeKey, cc.Connect, cc.LocalAmt, cc.PushAmt, cc.Private, cc.Block, cc.MinConfs, cc.ConfTarget, cc.SatPerByte, cc.MinHtlcmSat, cc.RemoteCsvDelay)
			if err != nil {
				logError(err.Error())
				displayMessage("Error : "+err.Error(), nil)
			} else {
				displayMessage("Txid : "+txid, nil)
			}
		}
	}

	cv.form.initialize(context.gocui)
}

func (cv *channelListView) closeChannel() {
	c := cv.getSelectedChannel()

	if c == nil {
		return
	}

	cc := new(closeChannelContainer)

	cc.ChannelPoint = c.RemotePubkey
	cc.NodeAlias = c.NodeAlias
	cc.LocalBalance = strconv.FormatInt(c.LocalBalance, 10)
	cc.RemoteBalance = strconv.FormatInt(c.RemoteBalance, 10)

	cv.form = newFormEdit("closeChanVal", "Close channel", cc)

	cv.form.callback = func(valid bool) {
		cv.form.getValue()
		cv.form.close(context.gocui)
		cv.form = nil
		if valid {
			txid, err := status.closeChannel(&context, c, cc.Force)
			if err != nil {
				logError(err.Error())
				displayMessage("Error : "+err.Error(), nil)
			} else {
				displayMessage("txid : "+txid, nil)
			}
		}
	}

	cv.form.initialize(context.gocui)
}

func (cv *channelListView) refreshView(g *gocui.Gui) {
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

func (cv *channelListView) getShortCuts() []*keyHandle {
	return cv.shortcuts
}

func (cv *channelListView) getGrid() *dataGrid {
	return cv.grid
}

func (cv *channelListView) getPhysicalView() string {
	return cv.mappedToPhysicalView
}
