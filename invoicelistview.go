package main

import (
	"fmt"
	"log"
	"time"

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

type invoiceDisplayContainer struct {
	Memo        string `displayname:"Memo" length:"50" readonly:"1"`
	Private     string `displayname:"Private" length:"50" readonly:"1"`
	Amt         string `displayname:"Amount" length:"50" readonly:"1"`
	Expiry      string `displayname:"Expiry(sec)" length:"50" readonly:"1"`
	Paid        string `displayname:"Paid" length:"50" readonly:"1"`
	Settled     string `displayname:"Settled" length:"50" readonly:"1"`
	SettledDate string `displayname:"Settled on" length:"50" readonly:"1"`
	PayReq      string `displayname:"PayReq" length:"50" lines:"5" readonly:"1"`
	QRCode      string `displayname:"PayReq" length:"50" lines:"32" readonly:"1"`
}

func newinvoiceListView(physicalView string, fmtnormal string, fmtheader string, fmtselected string) *invoiceListView {
	cv := new(invoiceListView)
	cv.grid = makeNewDataGrid()
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
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Details invoice", "D", 'd', gocui.ModAlt, cv.detailsInvoice, true, ""})

	cv.grid.key = "invoices"
	cv.grid.addColumn("Settled", "GetSettled", boolRow)       //"Settled", 2
	cv.grid.addColumn("Private", "GetPrivate", boolRow)       //"Private", 2
	cv.grid.addColumn("Memo", "GetMemo", stringRow)           //"Memo", 0
	cv.grid.addColumn("Value", "GetValue", intRow)            //"Value", 16
	cv.grid.addColumn("Creation", "GetCreationDate", dateRow) //"Creation",18
	cv.grid.addColumn("Settled", "GetSettledDate", dateRow)   //"Settled", 18
	cv.grid.addColumn("Expiry", "GetExpiry", intRow)          //"Expiry(s)", 10
	cv.grid.addColumn("Paid", "GetAmtPaidMsat", intRow)       //"Paid mSat", 16
	cv.grid.initConfig()
}

func (cv *invoiceListView) getSelectedInvoice() *lncliInvoice {
	return cv.grid.getSelectedItem().Interface().(*lncliInvoice)
}

func (cv *invoiceListView) detailsInvoice() {

	c := cv.getSelectedInvoice()

	if c == nil {
		return
	}

	cc := new(invoiceDisplayContainer)

	cc.Amt = context.printer.Sprintf("%d", c.GetValue())
	cc.Expiry = context.printer.Sprintf("%d", c.GetExpiry())
	cc.Memo = c.GetMemo()
	cc.Paid = context.printer.Sprintf("%d", c.GetAmtPaidMsat())
	if c.GetPrivate() {
		cc.Private = "X"
	}
	cc.PayReq = c.GetPaymentRequest()
	cc.QRCode = "\n" + getQRString(cc.PayReq)
	if c.GetSettled() {
		cc.Settled = "X"
		cc.SettledDate = context.printer.Sprintf("%s", time.Unix(c.GetSettleDate(), 0).Format("02-01-06 15:04:05"))
	}

	cv.form = newFormEdit("invoicedetailsVal", "Invoice details", cc)

	cv.form.callback = func(valid bool) {
		cv.form.close(context.gocui)
		cv.form = nil
		cc = nil
	}

	cv.form.initialize(context.gocui)
	cv.form.switchActiveEditor(-1, context.gocui)
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
				displayMessageWithSize(val+"\n"+getQRString(val), nil, 60, 32)
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
