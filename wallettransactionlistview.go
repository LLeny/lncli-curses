package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type walletTransactionListView struct {
	viewBase
	form *formEdit
}

type walletNewAddressRequestContainer struct {
	AdressType string `displayname:"Address type" length:"6"`
}

type walletNewAddressReponseContainer struct {
	AdressType string `displayname:"Address type" length:"6" readonly:"1"`
	Adress     string `displayname:"Address" length:"50" readonly:"1"`
}

func newwalletTransactionListView(physicalView string, fmtnormal string, fmtheader string, fmtselected string) *walletTransactionListView {
	cv := new(walletTransactionListView)

	cv.grid = makeNewDataGrid()
	cv.mappedToPhysicalView = physicalView

	cv.init(fmtnormal, fmtheader, fmtselected)

	return cv
}

func (cv *walletTransactionListView) init(fmtnormal string, fmtheader string, fmtselected string) {

	cv.grid.fmtForeground = fmtnormal
	cv.grid.fmtHeader = fmtheader
	cv.grid.fmtSelected = fmtselected

	cv.shortcuts = nil
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Up", "Up", gocui.KeyArrowUp, gocui.ModNone, func() { cv.grid.moveSelectionUp() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Down", "Down", gocui.KeyArrowDown, gocui.ModNone, func() { cv.grid.moveSelectionDown() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"New address", "N", 'n', gocui.ModAlt, cv.newaddress, true, ""})

	cv.grid.key = "walletTransactions"
	cv.grid.addColumn("Amount", "Amount", intRow)                  //"Amount",12
	cv.grid.addColumn("Confirmations", "NumConfirmations", intRow) //"Conf.",8
	cv.grid.addColumn("BlockHeight", "BlockHeight", intRow)        //"Bloc kHeight", 8
	cv.grid.addColumn("Fees", "TotalFees", intRow)                 //"Fees",8
	cv.grid.addColumn("Timestamp", "TimeStamp", dateRow)           //"Timestamp",18
	cv.grid.addColumn("TxHash", "TxHash", stringRow)               //"Tx Hash", 0
	cv.grid.addColumn("BlockHash", "BlockHash", stringRow)         //"Block Hash",0
	cv.grid.addColumn("Destination", "DestAddresses", sliceRow)    //"Dest.",0
	cv.grid.initConfig()
}

func (cv *walletTransactionListView) newaddress() {
	cc := new(walletNewAddressRequestContainer)

	cc.AdressType = "np2wkh"

	cv.form = newFormEdit("walletnewaddressreq", "New adress", cc)

	cv.form.callback = func(valid bool) {
		cv.form.getValue()

		var na string
		var err error

		if valid {
			na, err = status.walletNewAdress(&context, cc.AdressType)
			if err != nil {
				logError(err.Error())
			}
		}
		cv.form.close(context.gocui)

		if err != nil {
			return
		}

		cv.displayAddress(cc.AdressType, na)
	}

	cv.form.initialize(context.gocui)
	cv.form.switchActiveEditor(-1, context.gocui)
}

func (cv *walletTransactionListView) displayAddress(at string, a string) {
	cr := new(walletNewAddressReponseContainer)

	cr.AdressType = at
	cr.Adress = a

	cv.form = newFormEdit("walletnewaddressresp", "New adress", cr)

	cv.form.callback = func(valid bool) {
		cv.form.close(context.gocui)
		cv.form = nil
	}

	cv.form.initialize(context.gocui)
	cv.form.switchActiveEditor(-1, context.gocui)
}

func (cv *walletTransactionListView) refreshView(g *gocui.Gui) {
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

func (cv *walletTransactionListView) getShortCuts() []*keyHandle {
	return cv.shortcuts
}

func (cv *walletTransactionListView) getGrid() *dataGrid {
	return cv.grid
}

func (cv *walletTransactionListView) getPhysicalView() string {
	return cv.mappedToPhysicalView
}
