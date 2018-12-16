package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jroimartin/gocui"
)

type logLevel int

const (
	info    logLevel = 0
	warning logLevel = 1
	errorr  logLevel = 2
)

type logEntry struct {
	Timestamp time.Time
	Level     logLevel
	Message   string
}

type logListView struct {
	viewBase
	form *formEdit
}

func newlogListView(physicalView string, fmtnormal string, fmtheader string, fmtselected string) *logListView {
	cv := new(logListView)

	cv.grid = &dataGrid{}
	cv.mappedToPhysicalView = physicalView

	cv.init(fmtnormal, fmtheader, fmtselected)

	return cv
}

func (cv *logListView) init(fmtnormal string, fmtheader string, fmtselected string) {

	cv.grid.fmtForeground = fmtnormal
	cv.grid.fmtHeader = fmtheader
	cv.grid.fmtSelected = fmtselected

	cv.shortcuts = nil
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Up", "Up", gocui.KeyArrowUp, gocui.ModNone, func() { cv.grid.moveSelectionUp() }, false, ""})
	cv.shortcuts = append(cv.shortcuts, &keyHandle{"Scroll Down", "Down", gocui.KeyArrowDown, gocui.ModNone, func() { cv.grid.moveSelectionDown() }, false, ""})

	cv.grid.header = "[Logs]"
	cv.grid.addColumn("Level", "Level", 6, stringRow)
	cv.grid.addColumn("Timestamp", "Timestamp", 18, dateRow)
	cv.grid.addColumn("Message", "Message", 0, stringRow)
}

func (cv *logListView) refreshView(g *gocui.Gui) {
	v, err := g.View(cv.mappedToPhysicalView)
	if err != nil {
		log.Panicln(err.Error())
		return
	}
	v.Clear()

	x, y := v.Size()

	cv.grid.items = context.logs
	cv.grid.setRenderSize(x, y)

	for _, row := range cv.grid.getGridRows() {
		fmt.Fprintln(v, row)
	}

	if cv.form != nil {
		cv.form.layout(g)
	}
}

func (cv *logListView) getShortCuts() []*keyHandle {
	return cv.shortcuts
}

func (cv *logListView) getGrid() *dataGrid {
	return cv.grid
}

func (cv *logListView) getPhysicalView() string {
	return cv.mappedToPhysicalView
}

func writelog(lvl logLevel, e string) {
	context.logs = append(context.logs, &logEntry{time.Now(), lvl, e})
}

func logError(e string) {
	writelog(errorr, e)
}
