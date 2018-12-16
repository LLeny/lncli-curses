package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type paymentListView struct {
	viewBase
	form *formEdit
}

// func (s *lncliStatus) payInvoice(ctxt *lnclicursesContext, payReq string, amt int, feeLimit int, feeLimitPerc int, force bool) (string, error) {

type payInvoiceContainer struct {
	PayReq       string `displayname:"Pay req." length:"70" lines:"3"`
	Amount       int    `displayname:"Amount" length:"16"`
	FeeLimit     int    `displayname:"Fee limit" length:"5"`
	FeeLimitPerc int    `displayname:"Fee limit perc" length:"3"`
	Force        bool   `displayname:"Force"`
}

// type disconnectPeer struct {
// 	NodeAlias string `displayname:"Node alias" length:"32" readonly:"1"`
// 	PubKey    string `displayname:"Pub key" length:"32" readonly:"1" lines:"3"`
// }
func newpaymentListView(physicalView string, fmtnormal string, fmtheader string, fmtselected string) *paymentListView {
	cv := new(paymentListView)
	cv.grid = &dataGrid{}
	cv.mappedToPhysicalView = physicalView
	cv.init(fmtnormal, fmtheader, fmtselected)
	return cv
}
func (cv *paymentListView) init(fmtnormal string, fmtheader string, fmtselected string) {
	cv.grid.fmtForeground = fmtnormal
	cv.grid.fmtHeader = fmtheader
	cv.grid.fmtSelected = fmtselected
	cv.shortcuts = nil
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Up", "Up", gocui.KeyArrowUp, gocui.ModNone, func() { cv.grid.moveSelectionUp() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Down", "Down", gocui.KeyArrowDown, gocui.ModNone, func() { cv.grid.moveSelectionDown() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Pay Invoice", "P", 'p', gocui.ModAlt, cv.payInvoice, true, ""})

	cv.grid.header = "[Payments]"
	cv.grid.addColumn("Creation", "CreationDate", 18, dateRow)
	cv.grid.addColumn("Hash", "PaymentHash", 0, stringRow)
	cv.grid.addColumn("Value mSat", "ValueMsat", 16, intRow)
	cv.grid.addColumn("Fee", "Fee", 16, intRow)
	cv.grid.addColumn("Preimage", "PaymentPreimage", 6, stringRow)
	cv.grid.addColumn("Path", "Path", 0, sliceRow)
}
func (cv *paymentListView) payInvoice() {
	cc := new(payInvoiceContainer)

	cv.form = newFormEdit("payInvoiceVal", "Pay invoice", cc)

	cv.form.callback = func(valid bool) {
		cv.form.getValue()
		cv.form.close(context.gocui)
		cv.form = nil
		if valid {
			_, err := status.payInvoice(&context, cc.PayReq, cc.Amount, cc.FeeLimit, cc.FeeLimitPerc, cc.Force)
			if err != nil {
				logError(err.Error())
				displayMessage("Error : "+err.Error(), nil)
			}
		}
	}

	cv.form.initialize(context.gocui)
}

// func (cv *paymentListView) getSelectedPeer() *lncliPeer {
// 	return cv.grid.getSelectedItem().Interface().(*lncliPeer)
// }
func (cv *paymentListView) refreshView(g *gocui.Gui) {
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
func (cv *paymentListView) getShortCuts() []*keyHandle {
	return cv.shortcuts
}
func (cv *paymentListView) getGrid() *dataGrid {
	return cv.grid
}
func (cv *paymentListView) getPhysicalView() string {
	return cv.mappedToPhysicalView
}
