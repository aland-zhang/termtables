// Copyright 2012 Apcera Inc. All rights reserved.

package termtables

import (
	"strings"

	"github.com/apcera/termtables/locale"
	"github.com/apcera/termtables/term"
)

var useUTF8ByDefault = false
var MaxColumns = 80

type Element interface {
	Render(*renderStyle) string
}

type Table struct {
	Style *TableStyle

	elements []Element
	minWidth int
	headers  []interface{}
	title    interface{}
}

// Unconditionally use UTF-8 box-drawing characters for the table, unless
// explicitly overriden in style.
func EnableUTF8() {
	useUTF8ByDefault = true
}

// If the current locale indicates a UTF-8 charmap, then enable use of
// UTF-8 bow-drawing characters by default.
func EnableUTF8PerLocale() {
	charmap := locale.GetCharmap()
	if strings.EqualFold(charmap, "UTF-8") {
		useUTF8ByDefault = true
	}
}

func init() {
	// do not enable UTF-8 per locale by default, breaks tests
	sz, err := term.GetSize()
	if err == nil && sz.Columns != 0 {
		MaxColumns = sz.Columns
	}
}

func CreateTable() *Table {
	t := &Table{elements: []Element{}, Style: DefaultStyle}
	if useUTF8ByDefault {
		t.Style.setUtfBoxStyle()
	}
	return t
}

func (t *Table) AddSeparator() {
	t.elements = append(t.elements, &Separator{})
}

func (t *Table) AddRow(items ...interface{}) *Row {
	row := CreateRow(items)
	t.elements = append(t.elements, row)
	return row
}

func (t *Table) AddTitle(title interface{}) {
	t.title = title

	t.minWidth = len(renderValue(title))
}

func (t *Table) AddHeaders(headers ...interface{}) {
	t.headers = headers[:]
}

func (t *Table) UTF8Box() {
	t.Style.setUtfBoxStyle()
}

func (t *Table) Render() (buffer string) {
	// elements is already populated with row data

	// initial top line
	if !t.Style.SkipBorder {
		if t.title != nil && t.headers == nil {
			t.elements = append([]Element{&Separator{where: LINE_SUBTOP}}, t.elements...)
		} else if t.title == nil && t.headers == nil {
			t.elements = append([]Element{&Separator{where: LINE_TOP}}, t.elements...)
		} else {
			t.elements = append([]Element{&Separator{where: LINE_INNER}}, t.elements...)
		}
	}

	// if we have headers, include them
	if t.headers != nil {
		ne := make([]Element, 2)
		ne[1] = CreateRow(t.headers)
		if t.title != nil {
			ne[0] = &Separator{where: LINE_SUBTOP}
		} else {
			ne[0] = &Separator{where: LINE_TOP}
		}
		t.elements = append(ne, t.elements...)
	}

	// if we have a title, write them
	if t.title != nil {
		ne := []Element{
			&StraightSeparator{where: LINE_TOP},
			CreateRow([]interface{}{CreateCell(t.title, &CellStyle{Alignment: AlignCenter, ColSpan: 999})}),
		}
		t.elements = append(ne, t.elements...)
	}

	// generate the runtime style
	style := createRenderStyle(t)

	// loop over the elements and render them
	for _, e := range t.elements {
		buffer += e.Render(style) + "\n"
	}

	// add bottom line
	if !style.SkipBorder {
		buffer += (&Separator{where: LINE_BOTTOM}).Render(style) + "\n"
	}

	return buffer
}
