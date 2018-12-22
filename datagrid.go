package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
)

type rowFormat int

const (
	stringRow rowFormat = 1
	intRow    rowFormat = 2
	boolRow   rowFormat = 3
	dateRow   rowFormat = 4
	sliceRow  rowFormat = 5
)

type dataGridColumn struct {
	propertyName string
	displayWidth int
	format       rowFormat
}

type dataGridColumnDisplay struct {
	dataGridColumn
	header string
	width  int
}

type dataGrid struct {
	key               string
	header            string
	items             interface{}
	availableColumns  map[string]*dataGridColumn
	columns           []*dataGridColumnDisplay
	visibleStartIndex int
	visibleWidth      int
	visibleHeight     int
	selectedIndex     int
	fmtForeground     string
	fmtHeader         string
	fmtSelected       string
}

func makeNewDataGrid() *dataGrid {
	g := new(dataGrid)
	g.availableColumns = make(map[string]*dataGridColumn)
	return g
}

func (dg *dataGrid) getSelectedItem() reflect.Value {
	return reflect.ValueOf(dg.items).Index(dg.selectedIndex)
}

func (dg *dataGrid) setRenderSize(width int, height int) {
	dg.visibleWidth = width
	dg.visibleHeight = height
	dg.balanceColumnsWidth()
}

func (dg *dataGrid) initConfig() {
	dg.header = getConfigGridHeader(dg.key)
	for _, col := range getConfigGridColumns(dg.key) {
		dg.addDisplayColumn(col.key, col.header, col.width)
	}
}

func (dg *dataGrid) addDisplayColumn(key string, header string, width int) {

	c, ok := dg.availableColumns[key]

	if !ok {
		logError(fmt.Sprintf("Column '%s' not available", key))
		return
	}

	dg.columns = append(dg.columns, &dataGridColumnDisplay{*c, header, width})
}

func (dg *dataGrid) addColumn(key string, propertyName string, format rowFormat) {
	dg.availableColumns[key] = &dataGridColumn{propertyName, 0, format}
}

func (dg *dataGrid) balanceColumnsWidth() {
	var usedWidth = 0
	var autoCount = 0

	for i := 0; i < len(dg.columns); i++ {
		usedWidth += dg.columns[i].width
		dg.columns[i].displayWidth = dg.columns[i].width
		if dg.columns[i].width == 0 {
			autoCount++
		}
	}

	if autoCount > 0 {
		autoColWidth := (dg.visibleWidth - usedWidth) / autoCount
		for i := 0; i < len(dg.columns); i++ {
			if dg.columns[i].width == 0 {
				dg.columns[i].displayWidth = autoColWidth
			}
		}
	}
}

func (dg *dataGrid) displayRow(rowData reflect.Value, selected bool) string {

	var buffer bytes.Buffer

	for _, col := range dg.columns {
		val := dg.getRowValue(rowData, col.propertyName)

		if selected {
			buffer.WriteString(dg.fmtSelected)
		} else {
			buffer.WriteString(dg.fmtForeground)
		}

		var tmpStr string
		colWidth := strconv.Itoa(col.displayWidth - 1)

		switch col.format {
		case boolRow:
			if val.Bool() {
				tmpStr = "X"
			} else {
				tmpStr = " "
			}
		case intRow:
			switch val.Kind() {
			case reflect.Int64, reflect.Int16, reflect.Int32:
				tmpStr = context.printer.Sprintf("%-"+colWidth+"d", val.Int())
			case reflect.Uint64, reflect.Uint16, reflect.Uint32:
				tmpStr = context.printer.Sprintf("%-"+colWidth+"d", val.Uint())
			}
		case stringRow:
			tmpStr = fmt.Sprintf("%-"+colWidth+"s", val.String())
		case dateRow:
			if val.IsValid() {
				tmpStr = fmt.Sprintf("%-"+colWidth+"s", time.Unix(val.Int(), 0).Format("02-01-06 15:04:05"))
			} else {
				tmpStr = fmt.Sprintf("%-"+colWidth+"s", " ")
			}
		case sliceRow:
			tmpStr = fmt.Sprintf("%-"+colWidth+"s", getSliceString(val))
		default:
			tmpStr = fmt.Sprintf("%-"+colWidth+"s", " ")
		}

		buffer.WriteString(cutTo(tmpStr, col.displayWidth-1))
		buffer.WriteString("│")
	}

	return buffer.String()
}

func getSliceString(val reflect.Value) string {

	var buffer bytes.Buffer

	for i := 0; i < val.Len(); i++ {
		if buffer.Len() != 0 {
			buffer.WriteString(", ")
		}
		v := val.Index(i)
		buffer.WriteString(fmt.Sprintf("%s", v))
	}

	return buffer.String()
}

func (dg *dataGrid) getRowValue(rowData reflect.Value, propertyName string) reflect.Value {

	v := rowData.FieldByName(propertyName)

	if v.IsValid() {
		return v
	}

	var inputs []reflect.Value

	t := rowData.Type()

	mt, exists := t.MethodByName(propertyName)

	if !exists {
		el := reflect.New(t)
		t = el.Type()
		mt, exists = t.MethodByName(propertyName)
		inputs = make([]reflect.Value, 1)
		inputs[0] = rowData.Addr()
	}

	if exists {
		r := mt.Func.Call(inputs)
		if len(r) > 0 {
			return r[0]
		}
	}

	return reflect.Value{}
}

func (dg *dataGrid) getGridRows() []string {

	if dg.items == nil {
		return nil
	}

	lastIndex := dg.visibleStartIndex + dg.visibleHeight - 2

	items := reflect.ValueOf(dg.items)

	if lastIndex > items.Len() {
		lastIndex = items.Len()
	}

	var ret = make([]string, dg.visibleHeight)

	ret[0] = dg.generateHeader()
	ret[1] = dg.generateColumnHeaders()

	var dest = 2

	for i := dg.visibleStartIndex; i < lastIndex; i++ {
		o := items.Index(i).Elem()
		if !o.IsValid() {
			break
		}
		ret[dest] = dg.displayRow(o, dg.selectedIndex == i)
		dest++
	}

	return ret
}

func cutTo(s string, length int) string {

	if length <= 0 {
		return ""
	}

	for utf8.RuneCountInString(s) > length {
		_, l := utf8.DecodeLastRuneInString(s)
		s = s[:len(s)-l]
	}

	return s
}

func (dg *dataGrid) generateHeader() string {
	return fmt.Sprintf(dg.fmtHeader+context.theme.bold+"%-"+strconv.Itoa(dg.visibleWidth)+"s", dg.header)
}

func (dg *dataGrid) generateColumnHeaders() string {
	var buffer bytes.Buffer

	buffer.WriteString(dg.fmtHeader)

	for _, col := range dg.columns {
		display := cutTo(col.header, col.displayWidth-1)
		buffer.WriteString(fmt.Sprintf("%-"+strconv.Itoa(col.displayWidth-1)+"s│", display))
	}

	return buffer.String()
}

func (dg *dataGrid) moveSelectionUp() {

	if dg.selectedIndex == 0 {
		return
	}

	dg.selectedIndex--

	if dg.selectedIndex < dg.visibleStartIndex {
		dg.visibleStartIndex--
	}
}

func (dg *dataGrid) moveSelectionDown() {

	len := reflect.ValueOf(dg.items).Len()

	if dg.selectedIndex == len-1 {
		return
	}

	dg.selectedIndex++

	lastPossibleIndex := len - 1

	if dg.selectedIndex > lastPossibleIndex {
		dg.selectedIndex = lastPossibleIndex
		return
	}

	if dg.selectedIndex > dg.visibleStartIndex+dg.visibleHeight-3 {
		dg.visibleStartIndex++
	}
}
