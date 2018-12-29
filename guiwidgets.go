package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/jroimartin/gocui"
)

/////////////////////////////////////////////

type editKeyHandle struct {
	view   string
	key    interface{}
	mode   gocui.Modifier
	action func(g *gocui.Gui, v *gocui.View) error
}

/////////////////////////////////////////////
type themeGUI struct {
	background   gocui.Attribute
	normal       string
	labelHeader  string
	highlight    string
	inverted     string
	gridHeader   string
	gridSelected string
	bold         string
	error        string
}

/////////////////////////////////////////////
type baseEditI interface {
	layout(*gocui.Gui) error
	getLabel() string
	setLabel(string)
	getName() string
	setName(string)
	getX() int
	setX(int)
	getY() int
	setY(int)
	getHeight() int
	setHeight(int)
	getLabelWidth() int
	setLabelWidth(int)
	getContentWidth() int
	setContentWidth(int)
	setActive(bool)
	getValue() interface{}
	delete(*gocui.Gui)
}

/////////////////////////////////////////////
type baseEdit struct {
	label        string
	showLabel    bool
	name         string
	x, y         int
	height       int
	active       bool
	labelWidth   int
	contentWidth int
	keyHandles   []*editKeyHandle
}

func newBaseEdit(name string, label string, x int, y int, height int, contentwidth int) *baseEdit {
	b := new(baseEdit)
	b.label = label
	b.name = name
	b.x = x
	b.y = y
	b.height = height
	b.contentWidth = contentwidth
	b.active = false
	b.keyHandles = nil
	b.showLabel = true
	return b
}

func (b *baseEdit) getLabel() string {
	return b.label
}

func (b *baseEdit) setLabel(label string) {
	b.label = label
}

func (b *baseEdit) getName() string {
	return b.name
}

func (b *baseEdit) setName(name string) {
	b.name = name
}

func (b *baseEdit) getX() int {
	return b.x
}

func (b *baseEdit) setX(x int) {
	b.x = x
}

func (b *baseEdit) getY() int {
	return b.y
}

func (b *baseEdit) setY(y int) {
	b.y = y
}

func (b *baseEdit) getHeight() int {
	return b.height
}

func (b *baseEdit) setHeight(h int) {
	b.height = h
}

func (b *baseEdit) getLabelWidth() int {
	return b.labelWidth
}

func (b *baseEdit) setLabelWidth(w int) {
	b.labelWidth = w
}

func (b *baseEdit) getContentWidth() int {
	return b.contentWidth
}

func (b *baseEdit) setContentWidth(w int) {
	b.contentWidth = w
}

func (b *baseEdit) setActive(a bool) {
	b.active = a
}

func (b *baseEdit) setShowLabel(a bool) {
	b.showLabel = a
}

func (b *baseEdit) addKeyHandler(kh *editKeyHandle) {
	b.keyHandles = append(b.keyHandles, kh)
}

func (b *baseEdit) registerKeyHandlers(g *gocui.Gui) {
	for _, kh := range b.keyHandles {
		g.SetKeybinding(kh.view, kh.key, kh.mode, kh.action)
	}
}

func (b *baseEdit) unregisterKeyHandlers(g *gocui.Gui) {
	for _, kh := range b.keyHandles {
		g.DeleteKeybinding(kh.view, kh.key, kh.mode)
	}
}

func (b *baseEdit) baseLayout(g *gocui.Gui) *gocui.View {

	width := b.getLabelWidth() + b.getContentWidth() + 3

	v, _ := g.SetView(b.name, b.getX(), b.getY(), b.getX()+width, b.getY()+b.getHeight()+1)

	v.Clear()

	v.BgColor = context.theme.background
	v.Editable = true
	v.Frame = false
	v.Wrap = true

	if b.showLabel && len(b.label) > 0 {
		b.drawLabel(v)
	}

	return v
}

func (b *baseEdit) drawLabel(v *gocui.View) {
	v.MoveCursor(0, 0, false)
	if b.active {
		fmt.Fprint(v, context.theme.highlight)
	} else {
		fmt.Fprint(v, context.theme.normal)
	}
	fmt.Fprintf(v, "%-"+strconv.Itoa(b.getLabelWidth())+"s: ", b.label)
}

func (b *baseEdit) delete(g *gocui.Gui) {
	b.unregisterKeyHandlers(g)
	g.DeleteView(b.name)
}

////////////////////////////////////////////
type boolEdit struct {
	baseEdit
	selected bool
}

func newBoolEdit(name string, label string, def bool) *boolEdit {
	b := new(boolEdit)
	b.setLabel(label)
	b.setName(name)
	b.setHeight(1)
	b.setContentWidth(1)
	b.selected = def
	b.setShowLabel(true)
	return b
}

func (b *boolEdit) layout(g *gocui.Gui) error {
	v := b.baseLayout(g)

	v.Editor = gocui.EditorFunc(b.editor)

	v, err := g.View(b.name)

	if err != nil {
		log.Printf("updateLayout %s\n", err.Error())
		return err
	}

	fmt.Fprint(v, context.theme.inverted)

	if b.selected {
		fmt.Fprint(v, "X")
	} else {
		fmt.Fprint(v, " ")
	}

	return nil
}

func (b *boolEdit) editor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case key == gocui.KeyEnter || key == gocui.KeySpace:
		b.selected = !b.selected
	}
}

func (b boolEdit) getValue() interface{} {
	return b.selected
}

////////////////////////////////////////////
type intEdit struct {
	baseEdit
	value string
}

func newIntEdit(name string, label string, width int, def int64) *intEdit {
	b := new(intEdit)
	b.setLabel(label)
	b.setName(name)
	b.setHeight(1)
	b.setContentWidth(width)
	b.setShowLabel(true)
	b.value = strconv.FormatInt(def, 10)
	return b
}

func (b *intEdit) layout(g *gocui.Gui) error {
	v := b.baseLayout(g)

	v.Editor = gocui.EditorFunc(b.editor)

	v, err := g.View(b.name)

	if err != nil {
		log.Printf("layout %s\n", err.Error())
		return err
	}

	fmt.Fprint(v, context.theme.inverted)
	fmt.Fprintf(v, "%-"+strconv.Itoa(b.getContentWidth())+"s", b.value)

	return nil
}

func (b *intEdit) editor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		if len(b.value) > 0 {
			b.value = b.value[:len(b.value)-1]
		}
	case len(b.value) < b.getContentWidth() && (ch == '-' || (ch >= '0' && ch <= '9')):
		if ch == '-' && len(b.value) > 0 {
			break
		}
		b.value += string(ch)
	}
}

func (b *intEdit) getValue() interface{} {
	r, _ := strconv.Atoi(b.value)
	return r
}

////////////////////////////////////////////
type textEdit struct {
	baseEdit
	value    string
	readonly bool
	lines    int
}

func newTextEdit(name string, label string, length int, readonly bool, def string, lines int) *textEdit {
	b := new(textEdit)
	b.setLabel(label)
	b.setName(name)
	b.setHeight(lines)
	b.setContentWidth(length)
	b.setShowLabel(true)
	b.value = def
	b.readonly = readonly
	b.lines = lines
	return b
}

func (b *textEdit) layout(g *gocui.Gui) error {
	v := b.baseLayout(g)

	v, err := g.View(b.name)

	v.Editor = gocui.EditorFunc(b.editor)
	v.Wrap = true

	if err != nil {
		log.Printf("layout %s\n", err.Error())
		return err
	}

	if b.readonly {
		fmt.Fprint(v, context.theme.normal)
	} else {
		fmt.Fprint(v, context.theme.inverted)
	}

	fmt.Fprintf(v, "%-"+strconv.Itoa(b.getContentWidth())+"s", b.value)

	return nil
}

func (b *textEdit) editor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {

	if b.readonly {
		return
	}

	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		if len(b.value) > 0 {
			b.value = b.value[:len(b.value)-1]
		}
	case len(b.value) < b.getContentWidth()*b.lines:
		b.value += string(ch)
	}
}

func (b *textEdit) getValue() interface{} {
	return b.value
}

//////////////////////////////////////////
type buttonEdit struct {
	baseEdit
	selected bool
	action   func()
}

func newButtonEdit(name string, label string) *buttonEdit {
	b := new(buttonEdit)
	b.setLabel(label)
	b.setName(name)
	b.setHeight(1)
	b.setContentWidth(len(label))
	b.setShowLabel(false)
	b.selected = false
	return b
}

func (b *buttonEdit) layout(g *gocui.Gui) error {
	v := b.baseLayout(g)

	v, err := g.View(b.name)

	if err != nil {
		log.Printf("layout %s\n", err.Error())
		return err
	}

	if b.selected {
		fmt.Fprint(v, context.theme.highlight)
	} else {
		fmt.Fprint(v, context.theme.normal)
	}

	fmt.Fprint(v, b.label)

	return nil
}

func (b *buttonEdit) getValue() interface{} {
	return nil
}

//////////////////////////////////////////

type formEdit struct {
	baseEdit
	title               string
	editors             []baseEditI
	selectedEditorIndex int
	ok                  *buttonEdit
	cancel              *buttonEdit
	callback            func(valid bool)
	toMap               interface{}
	minWidth            int
	minHeight           int
}

func newFormEditWithSize(name string, title string, toMap interface{}, width int, height int) *formEdit {
	f := new(formEdit)
	f.setName(name)
	f.selectedEditorIndex = 0
	f.title = fmt.Sprintf(" %s ", title)
	f.setShowLabel(false)
	f.ok = newButtonEdit(name+"bok", "OK")
	f.cancel = newButtonEdit(name+"bcancel", "Cancel")
	f.toMap = toMap
	f.minWidth = width
	f.minHeight = height
	f.fromStruct(toMap)
	return f
}

func newFormEdit(name string, title string, toMap interface{}) *formEdit {
	f := new(formEdit)
	f.setName(name)
	f.selectedEditorIndex = 0
	f.title = fmt.Sprintf(" %s ", title)
	f.setShowLabel(false)
	f.ok = newButtonEdit(name+"bok", "OK")
	f.cancel = newButtonEdit(name+"bcancel", "Cancel")
	f.toMap = toMap
	f.minWidth = -1
	f.minHeight = -1
	f.fromStruct(toMap)
	return f
}

func (f *formEdit) fromStruct(toMap interface{}) {
	v := reflect.ValueOf(toMap).Elem()
	t := reflect.TypeOf(toMap).Elem()

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		vf := v.Field(i)
		switch vf.Interface().(type) {
		case int:
			f.intEditor(sf, vf.Int())
		case string:
			f.stringEditor(sf, vf.String())
		case bool:
			f.boolEditor(sf, vf.Bool())
		}
	}
}

func (f *formEdit) getWidth(fieldWidth int, err error) int {
	if f.minWidth >= 0 {
		if err != nil {
			return f.minWidth
		}
		if f.minWidth > fieldWidth {
			return f.minWidth
		}
	}
	return fieldWidth
}

func (f *formEdit) boolEditor(v reflect.StructField, def bool) {
	displayName := v.Tag.Get("displayname")
	e := newBoolEdit(v.Name, displayName, def)
	f.addEditor(e)
}

func (f *formEdit) intEditor(v reflect.StructField, def int64) {
	displayName := v.Tag.Get("displayname")
	length := f.getWidth(strconv.Atoi(v.Tag.Get("length")))
	e := newIntEdit(v.Name, displayName, length, def)
	f.addEditor(e)
}

func (f *formEdit) stringEditor(v reflect.StructField, def string) {
	displayName := v.Tag.Get("displayname")
	readonly, hasreadonly := v.Tag.Lookup("readonly")
	lines, haslines := v.Tag.Lookup("lines")
	length := f.getWidth(strconv.Atoi(v.Tag.Get("length")))
	l := 1
	if haslines {
		l, _ = strconv.Atoi(lines)
		if f.minHeight > 0 && l < f.minHeight {
			l = f.minHeight
		}
	}
	e := newTextEdit(v.Name, displayName, length, hasreadonly && readonly == "1", def, l)
	f.addEditor(e)
}

func (f *formEdit) registerKeys(g *gocui.Gui) {
	kh := new(editKeyHandle)
	kh.view = ""
	kh.key = gocui.KeyTab
	kh.mode = gocui.ModNone
	kh.action = func(g *gocui.Gui, v *gocui.View) error {
		f.switchActiveEditor(1, g)
		return nil
	}
	f.addKeyHandler(kh)

	khb := new(editKeyHandle)
	khb.view = ""
	khb.key = gocui.KeyTab
	khb.mode = gocui.ModAlt
	khb.action = func(g *gocui.Gui, v *gocui.View) error {
		f.switchActiveEditor(-1, g)
		return nil
	}
	f.addKeyHandler(khb)

	khe := new(editKeyHandle)
	khe.view = ""
	khe.key = gocui.KeyEnter
	khe.mode = gocui.ModNone
	khe.action = func(g *gocui.Gui, v *gocui.View) error {
		if !f.ok.selected && !f.cancel.selected {
			return nil
		}
		f.callback(f.ok.selected)
		return nil
	}
	f.addKeyHandler(khe)

	khc := new(editKeyHandle)
	khc.view = ""
	khc.key = gocui.KeyEsc
	khc.mode = gocui.ModNone
	khc.action = func(g *gocui.Gui, v *gocui.View) error {
		f.callback(false)
		return nil
	}
	f.addKeyHandler(khc)
}

func (f *formEdit) switchActiveEditor(delta int, g *gocui.Gui) {
	if delta == 0 {
		f.selectedEditorIndex = 0
	}

	f.selectedEditorIndex += delta

	l := len(f.editors)

	if f.selectedEditorIndex < 0 {
		switch {
		case f.selectedEditorIndex == -1:
			f.ok.selected = true
			f.cancel.selected = false
		case f.selectedEditorIndex == -2:
			f.ok.selected = false
			f.cancel.selected = true
		case f.selectedEditorIndex == -3:
			f.selectedEditorIndex = l - 1
		}
	}

	if f.selectedEditorIndex >= l {
		switch {
		case f.selectedEditorIndex == l:
			f.ok.selected = false
			f.cancel.selected = true
		case f.selectedEditorIndex == l+1:
			f.ok.selected = true
			f.cancel.selected = false
		case f.selectedEditorIndex == l+2:
			f.selectedEditorIndex = 0
		}
	}

	if f.selectedEditorIndex >= 0 && f.selectedEditorIndex < l {
		f.ok.selected = false
		f.cancel.selected = false
		g.SetCurrentView(f.editors[f.selectedEditorIndex].getName())
	}
}

func (f *formEdit) addEditor(e baseEditI) {
	f.editors = append(f.editors, e)
}

func (f *formEdit) layout(g *gocui.Gui) error {
	f.baseLayout(g)

	w, h := g.Size()
	xv := (w - f.getContentWidth()) / 2
	yv := (h - f.getHeight()) / 2

	f.setX(xv)
	f.setY(yv)
	f.setButtonLocs()
	v, _ := g.SetView(f.name, xv, yv, xv+f.getContentWidth(), yv+f.getHeight())

	v.Editable = true
	v.Title = f.title
	v.Frame = true

	y := f.getY() + 1
	x := f.getX() + 1

	for i, e := range f.editors {
		e.setY(y)
		e.setX(x)
		e.setActive(i == f.selectedEditorIndex)
		err := e.layout(g)
		if err != nil {
			return err
		}
		y += e.getHeight()
	}

	f.ok.layout(g)
	f.cancel.layout(g)

	return nil
}

func (f *formEdit) initialize(g *gocui.Gui) {
	f.registerKeys(g)
	f.registerKeyHandlers(g)
	f.setEditorsLabelWidth()
	f.setSize()
	f.setButtonLocs()
	f.layout(g)
	f.switchActiveEditor(0, g)
}

func (f *formEdit) setButtonLocs() {
	y := f.getY() + f.getHeight() - 2
	x := f.getX() + f.getContentWidth() - 3

	f.ok.setY(y)
	f.ok.setX(x - 2)
	f.cancel.setY(y)
	f.cancel.setX(x - 10)
}

func (f *formEdit) setEditorsLabelWidth() {
	maxLabelWidth := 0

	for _, e := range f.editors {
		w := len(e.getLabel())
		if w > maxLabelWidth {
			maxLabelWidth = w
		}
	}

	for _, e := range f.editors {
		e.setLabelWidth(maxLabelWidth)
	}
}

func (f *formEdit) setSize() {
	maxWidth := 0
	height := 0

	for _, e := range f.editors {
		height += e.getHeight()
		w := e.getLabelWidth() + e.getContentWidth() + 5
		if w > maxWidth {
			maxWidth = w
		}
	}

	f.setContentWidth(maxWidth)
	f.setHeight(height + 1 + 3)
}

func (f *formEdit) getValue() interface{} {
	if f == nil {
		return nil
	}

	v := reflect.ValueOf(f.toMap).Elem()
	t := reflect.TypeOf(f.toMap).Elem()

	for i := 0; i < v.NumField(); i++ {
		vf := v.Field(i)
		sf := t.Field(i)
		e := f.getEditor(sf.Name)
		ev := (*e).getValue()

		switch vf.Interface().(type) {
		case int:
			vf.SetInt(int64(ev.(int)))
		case string:
			vf.SetString(ev.(string))
		case bool:
			vf.SetBool(ev.(bool))
		}
	}

	return f.toMap
}

func (f *formEdit) getEditor(name string) *baseEditI {
	for _, e := range f.editors {
		if e.getName() == name {
			return &e
		}
	}

	return nil
}

func (f *formEdit) close(g *gocui.Gui) {
	for _, e := range f.editors {
		e.delete(g)
	}
	f.ok.delete(g)
	f.cancel.delete(g)
	f.delete(g)
}

/////////////////////////////////////////////////////

type displayMessageContainer struct {
	Message string `displayname:"" length:"48" lines:"8" readonly:"1"`
}

func displayMessageWithSize(msg string, closed func(bool), width int, height int) {

	tmp := displayMessageContainer{msg}

	form := newFormEditWithSize("displayForm", "Information", &tmp, width, height)

	form.callback = func(valid bool) {
		form.close(context.gocui)
		form = nil
		if closed != nil {
			closed(valid)
		}
	}

	form.switchActiveEditor(-1, context.gocui)
	form.initialize(context.gocui)
}

func displayMessage(msg string, closed func(bool)) {

	tmp := displayMessageContainer{msg}

	form := newFormEdit("displayForm", "Information", &tmp)

	form.callback = func(valid bool) {
		form.close(context.gocui)
		form = nil
		if closed != nil {
			closed(valid)
		}
	}

	form.switchActiveEditor(-1, context.gocui)
	form.initialize(context.gocui)
}
