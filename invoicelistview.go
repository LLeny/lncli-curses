package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type invoiceListView struct {
	viewBase
	form *formEdit
}

type addInvoiceContainer struct {
	Memo            string `displayname:"Memo (opt)" length:"64"`
	Receipt         string `displayname:"Receipt (opt)" length:"64"`
	Preimage        string `displayname:"Preimage" length:"64"`
	Amt             int    `displayname:"Amount" length:"16"`
	DescriptionHash string `displayname:"Description hash" length:"64"`
	FallbackAddr    string `displayname:"Fallback Adddress" length:"64"`
	Expiry          int    `displayname:"Expiry sec" length:"8"`
	Private         bool   `displayname:"Private"`
}

func newinvoiceListView(physicalView string, fmtnormal string, fmtheader string, fmtselected string) *invoiceListView {
	cv := new(invoiceListView)
	cv.grid = &dataGrid{}
	cv.mappedToPhysicalView = physicalView
	cv.init(fmtnormal, fmtheader, fmtselected)
	return cv
}
func (cv *invoiceListView) init(fmtnormal string, fmtheader string, fmtselected string) {
	cv.grid.fmtForeground = fmtnormal
	cv.grid.fmtHeader = fmtheader
	cv.grid.fmtSelected = fmtselected
	cv.shortcuts = nil
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Up", "Up", gocui.KeyArrowUp, gocui.ModNone, func() { cv.grid.moveSelectionUp() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Down", "Down", gocui.KeyArrowDown, gocui.ModNone, func() { cv.grid.moveSelectionDown() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Add invoice", "A", 'a', gocui.ModAlt, cv.addInvoice, true, ""})

	cv.grid.header = "[Invoices]"
	cv.grid.addColumn("Settled", "GetSettled", 2, boolRow)
	cv.grid.addColumn("Memo", "GetMemo", 0, stringRow)
	cv.grid.addColumn("Value", "GetValue", 16, intRow)
	cv.grid.addColumn("Creation", "GetCreationDate", 18, dateRow)
	cv.grid.addColumn("Settled", "GetSettledDate", 18, dateRow)
	cv.grid.addColumn("Expiry(s)", "GetExpiry", 10, intRow)
	cv.grid.addColumn("Paid mSat", "GetAmtPaidMsat", 16, intRow)
}

func (cv *invoiceListView) addInvoice() {
	cc := new(addInvoiceContainer)

	cv.form = newFormEdit("addInvoiceVal", "Add an invoice", cc)
	cv.form.callback = func(valid bool) {
		cv.form.getValue()
		cv.form.close(context.gocui)
		cv.form = nil
		if valid {
			val, err := status.addInvoice(&context, cc.Amt, cc.DescriptionHash, cc.Expiry, cc.FallbackAddr, cc.Memo, cc.Preimage, cc.Private, cc.Receipt)
			if err != nil {
				logError(err.Error())
				displayMessage("Error: "+err.Error(), nil)
			} else {
				displayMessage("Payment request: "+val, nil)
			}
			updateData()
		}
	}
	cv.form.initialize(context.gocui)
}

func (cv *invoiceListView) getSelectedPeer() *lncliPeer {
	return cv.grid.getSelectedItem().Interface().(*lncliPeer)
}

func (cv *invoiceListView) refreshView(g *gocui.Gui) {
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
func (cv *invoiceListView) getShortCuts() []*keyHandle {
	return cv.shortcuts
}
func (cv *invoiceListView) getGrid() *dataGrid {
	return cv.grid
}
func (cv *invoiceListView) getPhysicalView() string {
	return cv.mappedToPhysicalView
}
