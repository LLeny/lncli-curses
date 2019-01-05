package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jroimartin/gocui"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type viewType int

const (
	nodeInfoViewt           viewType = 0
	walletBalanceViewt      viewType = 1
	channelListViewt        viewType = 2
	peerListViewt           viewType = 3
	walletTransactionsViewt viewType = 4
	pendingChannelListViewt viewType = 5
	paymentListViewt        viewType = 6
	invoiceListViewt        viewType = 7
	channelDetailsViewt     viewType = 102
	menuViewt               viewType = 1000
	globalViewt             viewType = 1001
	mainViewt               viewType = 1002
	logViewt                viewType = 1003
)

type viewBase struct {
	grid                 *dataGrid
	shortcuts            []*keyHandle
	mappedToPhysicalView string
}

type viewI interface {
	init(fmtnormal string, fmtheader string, fmtselected string)
	refreshView(g *gocui.Gui)
	getShortCuts() []*keyHandle
	getGrid() *dataGrid
	getPhysicalView() string
}

type keyHandle struct {
	header    string
	keyHeader string
	key       interface{}
	mod       gocui.Modifier
	action    func()
	visible   bool
	view      string
}

func initViews() {

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	context.gocui = g

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	for _, h := range context.globalShortcuts {
		registerKeyHandler(g, h)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func registerKeyHandlers(kh []*keyHandle) {
	for _, k := range kh {
		registerKeyHandler(context.gocui, k)
	}
}

func unregisterKeyHandlers(kh []*keyHandle) {
	for _, k := range kh {
		unregisterKeyHandler(context.gocui, k)
	}
}

func unregisterKeyHandler(g *gocui.Gui, handle *keyHandle) error {
	if err := g.DeleteKeybinding(handle.view, handle.key, handle.mod); err != nil {
		return err
	}
	return nil
}

func registerKeyHandler(g *gocui.Gui, handle *keyHandle) error {
	if err := g.SetKeybinding(handle.view, handle.key, handle.mod,
		func(g *gocui.Gui, v *gocui.View) error {
			handle.action()
			return nil
		}); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func refreshView() {
	context.gocui.Update(func(g *gocui.Gui) error {
		if getShowHeader() {
			refreshNodeInfoView(g)
			refreshWalletBalanceView(g)
		}
		refreshMainView(g)
		refreshMenuView(g)

		return nil
	})
}

func refreshWalletBalanceView(g *gocui.Gui) {
	v, err := g.View("balance")
	if err != nil {
		log.Panicln(err.Error())
		return
	}
	v.Clear()
	p := message.NewPrinter(language.English)
	p.Fprintf(v, context.theme.normal+"Total       "+context.theme.highlight+"%14d"+context.theme.normal+" mSat\n", status.walletBalance.TotalBalance)
	p.Fprintf(v, context.theme.normal+"Confirmed   "+context.theme.highlight+"%14d"+context.theme.normal+" mSat\n", status.walletBalance.ConfirmedBalance)
	p.Fprintf(v, context.theme.normal+"Unconfirmed "+context.theme.highlight+"%14d"+context.theme.normal+" mSat", status.walletBalance.UnconfirmedBalance)
}

func refreshNodeInfoView(g *gocui.Gui) {
	v, err := g.View("nodeinfo")
	if err != nil {
		log.Panicln(err.Error())
		return
	}
	v.Clear()
	fmt.Fprintf(v, context.theme.normal+"Alias    "+context.theme.highlight+"%s\n", status.localNodeInfo.Alias)
	fmt.Fprintf(v, context.theme.normal+"Pubkey   "+context.theme.highlight+"%s\n", status.localNodeInfo.IdentityPubkey)
	fmt.Fprintf(v, context.theme.normal+"Version  "+context.theme.highlight+"%s\n", status.localNodeInfo.Version)
	fmt.Fprintf(v, context.theme.normal+"Chains   "+context.theme.highlight+"%s", strings.Join(status.localNodeInfo.Chains, ","))
	if status.localNodeInfo.Testnet {
		fmt.Fprint(v, " testnet")
	} else {
		fmt.Fprint(v, " mainnet")
	}
	if status.localNodeInfo.SyncedToChain {
		fmt.Fprintln(v, " synced")
	} else {
		fmt.Fprintln(v, " not synced")
	}
	fmt.Fprintf(v, context.theme.normal+"Peers    "+context.theme.highlight+"%d\n", status.localNodeInfo.NumPeers)
	fmt.Fprintf(v, context.theme.normal+"Channels active "+context.theme.highlight+"%d "+context.theme.normal+"inactive "+context.theme.highlight+"%d "+context.theme.normal+"pending "+context.theme.highlight+"%d", status.localNodeInfo.NumActiveChannels, status.localNodeInfo.NumInactiveChannels, status.localNodeInfo.NumPendingChannels)
}

func refreshMainView(g *gocui.Gui) {
	context.views[context.activeMainView].refreshView(g)
}

func getModifierString(mod gocui.Modifier) string {
	switch mod {
	case gocui.ModAlt:
		return "A"
	case gocui.ModNone:
		return " "
	}
	return " "
}

func refreshMenuView(g *gocui.Gui) {
	v, err := g.View("menu")
	if err != nil {
		log.Panicln(err.Error())
		return
	}
	v.Clear()

	x, _ := v.Size()

	var buffer bytes.Buffer
	var bufferlocal bytes.Buffer
	var menulen int
	var menulocallen int

	for _, entry := range context.globalShortcuts {
		buffer.WriteString(
			fmt.Sprintf("%s%s%s+%s%s %s ",
				context.theme.gridHeader, context.theme.gridSelected,
				getModifierString(entry.mod), entry.keyHeader,
				context.theme.gridHeader,
				entry.header))
		menulen += (4 + len(entry.keyHeader) + len(entry.header))
	}

	for _, entry := range context.views[context.activeMainView].getShortCuts() {
		if entry.visible {
			bufferlocal.WriteString(fmt.Sprintf("%s%s%s+%s%s %s ",
				context.theme.gridHeader, context.theme.gridSelected,
				getModifierString(entry.mod), entry.keyHeader,
				context.theme.gridHeader,
				entry.header))
			menulocallen += (4 + len(entry.keyHeader) + len(entry.header))
		}
	}

	fmt.Fprintf(v, "%s", buffer.String())
	fmt.Fprintf(v, "%s%"+strconv.Itoa(x-menulen-menulocallen)+"s", context.theme.gridHeader, " ")
	fmt.Fprintf(v, "%s", bufferlocal.String())
}

func layout(g *gocui.Gui) error {

	maxX, maxY := g.Size()

	headerHeight := 1

	if getShowHeader() {

		headerHeight = 7

		if v, err := g.SetView("nodeinfo", -1, -1, maxX-36, headerHeight); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Editable = false
			v.Frame = false
			v.BgColor = context.theme.background
		}
		refreshNodeInfoView(g)

		if v, err := g.SetView("balance", maxX-37, -1, maxX, headerHeight); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Editable = false
			v.Frame = false
			v.BgColor = context.theme.background
		}
		refreshWalletBalanceView(g)
	}

	if v, err := g.SetView("main", -1, headerHeight-2, maxX, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = false
		v.Frame = false
		v.BgColor = context.theme.background
	}
	refreshMainView(g)

	if v, err := g.SetView("menu", -1, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = false
		v.Frame = false
		v.BgColor = context.theme.background
	}
	refreshMenuView(g)

	return nil
}
