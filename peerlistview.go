package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type peerListView struct {
	viewBase
	form *formEdit
}

type connectPeer struct {
	PubKey string `displayname:"Pub key" length:"64"`
	Host   string `displayname:"Host" length:"16"`
	Port   int    `displayname:"Port" length:"5"`
}

type disconnectPeer struct {
	NodeAlias string `displayname:"Node alias" length:"32" readonly:"1"`
	PubKey    string `displayname:"Pub key" length:"32" readonly:"1" lines:"3"`
}

func newpeerListView(physicalView string, fmtnormal string, fmtheader string, fmtselected string) *peerListView {
	cv := new(peerListView)

	cv.grid = makeNewDataGrid()
	cv.mappedToPhysicalView = physicalView

	cv.init(fmtnormal, fmtheader, fmtselected)

	return cv
}

func (cv *peerListView) init(fmtnormal string, fmtheader string, fmtselected string) {

	cv.grid.fmtForeground = fmtnormal
	cv.grid.fmtHeader = fmtheader
	cv.grid.fmtSelected = fmtselected

	cv.shortcuts = nil
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Up", "Up", gocui.KeyArrowUp, gocui.ModNone, func() { cv.grid.moveSelectionUp() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Down", "Down", gocui.KeyArrowDown, gocui.ModNone, func() { cv.grid.moveSelectionDown() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Connect", "C", 'c', gocui.ModAlt, cv.connect, true, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Disconnect", "D", 'd', gocui.ModAlt, cv.disconnect, true, ""})

	cv.grid.key = "peers"
	cv.grid.addColumn("Alias", "Alias", stringRow)      // "Alias", 0
	cv.grid.addColumn("Address", "Address", stringRow)  //"Address", 22
	cv.grid.addColumn("BytesSent", "BytesSent", intRow) //"Bytes sent",13
	cv.grid.addColumn("BytesRec", "BytesRecv", intRow)  //"Bytes rec.",13
	cv.grid.addColumn("SatSent", "SatSent", intRow)     //"Sat sent", 12
	cv.grid.addColumn("SatRec", "SatRecv", intRow)      //"Sat rec.", 12
	cv.grid.addColumn("Inbound", "Inbound", boolRow)    //"Inbound",2
	cv.grid.addColumn("Ping", "PingTime", intRow)       //"Ping",6
	cv.grid.initConfig()
}

func (cv *peerListView) connect() {

	cc := new(connectPeer)
	cc.Port = 9735

	cv.form = newFormEdit("connectPeerVal", "Connect to peer", cc)

	cv.form.callback = func(valid bool) {
		cv.form.getValue()
		if valid {
			status.connectToPeer(&context, cc.PubKey, cc.Host, cc.Port)
			updateData()
		}
		cv.form.close(context.gocui)
		cv.form = nil
		cc = nil
	}

	cv.form.initialize(context.gocui)
}

func (cv *peerListView) getSelectedPeer() *lncliPeer {
	return cv.grid.getSelectedItem().Interface().(*lncliPeer)
}

func (cv *peerListView) disconnect() {

	c := cv.getSelectedPeer()

	cc := new(disconnectPeer)

	cc.NodeAlias = c.Alias
	cc.PubKey = c.PubKey

	cv.form = newFormEdit("disconnectPeerForm", "Disconnect peer", cc)

	cv.form.callback = func(valid bool) {
		if valid {
			status.disconnectPeer(&context, c)
			updateData()
		}
		cv.form.close(context.gocui)
		cv.form = nil
		cc = nil
	}

	cv.form.initialize(context.gocui)
}

func (cv *peerListView) refreshView(g *gocui.Gui) {
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

func (cv *peerListView) getShortCuts() []*keyHandle {
	return cv.shortcuts
}

func (cv *peerListView) getGrid() *dataGrid {
	return cv.grid
}

func (cv *peerListView) getPhysicalView() string {
	return cv.mappedToPhysicalView
}
