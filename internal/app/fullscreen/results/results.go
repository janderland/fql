package results

import (
	"container/list"
	"fmt"
	"strings"

	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/janderland/fdbq/engine/stream"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/parser/format"
)

type keyMap struct {
	PageDown     key.Binding
	PageUp       key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	Down         key.Binding
	Up           key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", " ", "f"),
			key.WithHelp("f/pgdn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("b/pgup", "page up"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
	}
}

type result struct {
	i     int
	value any
}

type Model struct {
	keyMap keyMap
	format format.Format

	// height is the max number of lines
	// that will be rendered.
	height int

	// builder is used by the View method
	// to construct the final output.
	builder *strings.Builder

	// list contains all the scrollable items.
	// Newer items are placed at the front.
	list *list.List

	// cursor points at the newest item
	// that will be displayed.
	cursor *list.Element

	// endCursor points at the oldest item
	// which cursor is allowed to scroll to.
	// This prevents scrolling past the
	// final page.
	endCursor *list.Element
}

func New() Model {
	return Model{
		keyMap:  defaultKeyMap(),
		format:  format.New(format.Cfg{}),
		builder: &strings.Builder{},
		list:    list.New(),
	}
}

func (x *Model) Reset() {
	x.list = list.New()
	x.cursor = nil
	x.endCursor = nil
}

func (x *Model) Height(height int) {
	x.height = height
	x.updateCursors()
}

func (x *Model) PushMany(list *list.List) {
	for cursor := list.Front(); cursor != nil; cursor = cursor.Next() {
		x.push(cursor.Value)
	}
	x.updateCursors()
}

func (x *Model) Push(val any) {
	x.push(val)
	x.updateCursors()
}

func (x *Model) push(val any) {
	x.list.PushFront(result{
		i:     x.list.Len() + 1,
		value: val,
	})
}

func (x *Model) updateCursors() {
	if x.list.Len() == 0 {
		return
	}

	x.endCursor = x.list.Back()
	for i := 0; i < x.height; i++ {
		if x.endCursor.Prev() == nil {
			break
		}

		// As we move the end cursor back through
		// the list, if we encounter the start
		// cursor then move it along with us.
		if x.cursor == x.endCursor {
			x.cursor = x.endCursor.Prev()
		}
		x.endCursor = x.endCursor.Prev()
	}
}

func (x *Model) View() string {
	if x.height == 0 || x.list.Len() == 0 {
		return ""
	}

	// If we have scrolled back through
	// the list then start our local
	// cursor there. Otherwise, start
	// at the front of the list.
	cursor := x.cursor
	if cursor == nil {
		cursor = x.list.Front()
	}

	for i := 0; i < x.height-1; i++ {
		if cursor.Next() == nil {
			break
		}
		cursor = cursor.Next()
	}

	x.builder.Reset()
	for i := 0; i < x.height; i++ {
		if cursor == nil {
			break
		}
		res := cursor.Value.(result)
		x.builder.WriteString(fmt.Sprintf("%d  %s\n", res.i, x.view(res.value)))
		cursor = cursor.Prev()
	}
	return x.builder.String()
}

func (x *Model) view(item any) string {
	switch val := item.(type) {
	case error:
		return fmt.Sprintf("ERR! %s", val)

	case string:
		return fmt.Sprintf("# %s", val)

	case keyval.KeyValue:
		x.format.Reset()
		x.format.KeyValue(val)
		return x.format.String()

	case directory.DirectorySubspace:
		x.format.Reset()
		x.format.Directory(convert.FromStringArray(val.GetPath()))
		return x.format.String()

	case stream.KeyValErr:
		if val.Err != nil {
			return x.view(val.Err)
		}
		return x.view(val.KV)

	case stream.DirErr:
		if val.Err != nil {
			return x.view(val.Err)
		}
		return x.view(val.Dir)

	default:
		return fmt.Sprintf("ERR! unexpected %T", val)
	}
}

func (x *Model) Update(msg tea.Msg) Model {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, x.keyMap.PageDown):
			x.scrollDown(x.height - 1)

		case key.Matches(msg, x.keyMap.PageUp):
			x.scrollUp(x.height - 1)

		case key.Matches(msg, x.keyMap.HalfPageDown):
			x.scrollDown(x.height / 2)

		case key.Matches(msg, x.keyMap.HalfPageUp):
			x.scrollUp(x.height / 2)

		case key.Matches(msg, x.keyMap.Down):
			x.scrollDown(1)

		case key.Matches(msg, x.keyMap.Up):
			x.scrollUp(1)
		}

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelDown:
			x.scrollDown(1)

		case tea.MouseWheelUp:
			x.scrollUp(1)
		}
	}

	return *x
}

func (x *Model) scrollDown(lines int) {
	if x.cursor == nil {
		return
	}
	for i := 0; i < lines; i++ {
		x.cursor = x.cursor.Prev()
		if x.cursor == nil {
			break
		}
	}
}

func (x *Model) scrollUp(lines int) {
	if x.list.Len() == 0 {
		return
	}
	if x.cursor == nil {
		x.cursor = x.list.Front()
	}
	for i := 0; i < lines; i++ {
		if x.cursor == x.endCursor {
			break
		}
		newCursor := x.cursor.Next()
		if newCursor == nil {
			break
		}
		x.cursor = newCursor
	}
}
