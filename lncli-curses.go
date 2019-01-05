package main

import (
	"sync"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/jroimartin/gocui"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type lnclicursesContext struct {
	gocui           *gocui.Gui
	activeMainView  viewType
	views           map[viewType]viewI
	globalShortcuts []*keyHandle
	theme           themeGUI
	logs            []*logEntry
	printer         *message.Printer
	cliMutex        *sync.Mutex
}

var context lnclicursesContext
var status lncliStatus

func manageError(err error) {
	if err == nil {
		return
	}
	logError(err.Error())
}

func setUpdateTicker() {
	ticker := time.NewTicker(time.Second * time.Duration(getRefreshSec()))
	go func() {
		for range ticker.C {
			updateData()
		}
	}()
}

func updateData() {
	if getShowHeader() {
		manageError(status.updateLocalNodeInfo(&context))
		manageError(status.updateWalletBalance(&context))
	}
	switch context.activeMainView {
	case channelListViewt:
		manageError(status.updateChannelList(&context))
	case peerListViewt:
		manageError(status.updatePeersList(&context))
	case pendingChannelListViewt:
		manageError(status.updatePendingChannelList(&context))
	case invoiceListViewt:
		manageError(status.updateInvoiceList(&context))
	case paymentListViewt:
		manageError(status.updatePaymentList(&context))
	case walletTransactionsViewt:
		manageError(status.updateWallletTransactionsList(&context))
	}
	refreshView()
}

func main() {

	status.nodes = make(map[string]lnrpc.NodeInfo)

	context.printer = message.NewPrinter(language.English)
	context.activeMainView = channelListViewt
	context.views = make(map[viewType]viewI)
	context.cliMutex = &sync.Mutex{}

	if !initConfig() {
		panic("Couldn't read configuration")
	}

	initTheme()
	initGrids()

	setUpdateTicker()
	initViews()
	switchActiveView(channelListViewt)
}

func switchActiveView(view viewType) {
	unregisterKeyHandlers(context.views[context.activeMainView].getShortCuts())
	context.activeMainView = view
	registerKeyHandlers(context.views[view].getShortCuts())
	go updateData()
}

func initGrids() {
	initChannelListGrid()
	initPeerListGrid()
	initPendingChannelListGrid()
	initPaymentListGrid()
	initInvoiceListGrid()
	initWalletTransactionListGrid()
	initLogListGrid()
}

func initChannelListGrid() {
	context.views[channelListViewt] = newchannelListView("main", context.theme.normal, context.theme.gridHeader, context.theme.gridSelected)
	context.globalShortcuts = append(context.globalShortcuts, &keyHandle{getConfigGridShortcutHeader("channels"), "1", '1', gocui.ModAlt, func() { switchActiveView(channelListViewt) }, true, ""})
}

func initPeerListGrid() {
	context.views[peerListViewt] = newpeerListView("main", context.theme.normal, context.theme.gridHeader, context.theme.gridSelected)
	context.globalShortcuts = append(context.globalShortcuts, &keyHandle{getConfigGridShortcutHeader("peers"), "2", '2', gocui.ModAlt, func() { switchActiveView(peerListViewt) }, true, ""})
}

func initPendingChannelListGrid() {
	context.views[pendingChannelListViewt] = newpendingchannelListView("main", context.theme.normal, context.theme.gridHeader, context.theme.gridSelected)
	context.globalShortcuts = append(context.globalShortcuts, &keyHandle{getConfigGridShortcutHeader("pendingChannels"), "3", '3', gocui.ModAlt, func() { switchActiveView(pendingChannelListViewt) }, true, ""})
}

func initPaymentListGrid() {
	context.views[paymentListViewt] = newpaymentListView("main", context.theme.normal, context.theme.gridHeader, context.theme.gridSelected)
	context.globalShortcuts = append(context.globalShortcuts, &keyHandle{getConfigGridShortcutHeader("payments"), "4", '4', gocui.ModAlt, func() { switchActiveView(paymentListViewt) }, true, ""})
}

func initInvoiceListGrid() {
	context.views[invoiceListViewt] = newinvoiceListView("main", context.theme.normal, context.theme.gridHeader, context.theme.gridSelected)
	context.globalShortcuts = append(context.globalShortcuts, &keyHandle{getConfigGridShortcutHeader("invoices"), "5", '5', gocui.ModAlt, func() { switchActiveView(invoiceListViewt) }, true, ""})
}

func initWalletTransactionListGrid() {
	context.views[walletTransactionsViewt] = newwalletTransactionListView("main", context.theme.normal, context.theme.gridHeader, context.theme.gridSelected)
	context.globalShortcuts = append(context.globalShortcuts, &keyHandle{getConfigGridShortcutHeader("walletTransactions"), "6", '6', gocui.ModAlt, func() { switchActiveView(walletTransactionsViewt) }, true, ""})
}

func initLogListGrid() {
	context.views[logViewt] = newlogListView("main", context.theme.normal, context.theme.gridHeader, context.theme.gridSelected)
	context.globalShortcuts = append(context.globalShortcuts, &keyHandle{"Logs", "7", '7', gocui.ModAlt, func() { switchActiveView(logViewt) }, true, ""})
}
