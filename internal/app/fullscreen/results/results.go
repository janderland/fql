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

type Model struct {
	keyMap keyMap
	format format.Format
	list   *list.List
	cursor *list.Element
	lines  []string
}

func New() Model {
	return Model{
		keyMap: defaultKeyMap(),
		format: format.New(format.Cfg{}),
		list:   list.New(),
	}
}

func (x *Model) Reset() {
	x.list = list.New()
	x.cursor = nil
}

func (x *Model) Height(height int) {
	x.lines = make([]string, height)
}

func (x *Model) PushMany(list *list.List) {
	for item := list.Front(); item != nil; item = item.Next() {
		x.list.PushFront(item.Value)
	}
}

func (x *Model) Push(item any) {
	x.list.PushFront(item)
}

func (x *Model) View() string {
	if x.height() == 0 {
		return ""
	}

	cursor := x.cursor
	if cursor == nil {
		cursor = x.list.Front()
	}

	i := 0
	for i = range x.lines {
		if cursor == nil {
			break
		}

		x.lines[i] = x.view(cursor.Value)
		cursor = cursor.Next()
	}

	var results strings.Builder
	for j := i; j >= 0; j-- {
		results.WriteString(x.lines[j])
		results.WriteRune('\n')
	}
	for j := i + 1; j < x.height(); j++ {
		results.WriteRune('\n')
	}
	return results.String()
}

func (x *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, x.keyMap.PageDown):
			if x.cursor == nil {
				break
			}
			for i := 0; i < x.height()/2; i++ {
				x.cursor = x.cursor.Prev()
				if x.cursor == nil {
					break
				}
			}

		case key.Matches(msg, x.keyMap.PageUp):
			if x.list.Front() == nil {
				break
			}
			if x.cursor == nil {
				x.cursor = x.list.Front()
			}
			for i := 0; i < x.height()/2; i++ {
				newCursor := x.cursor.Next()
				if newCursor == nil {
					break
				}
				x.cursor = newCursor
			}

		case key.Matches(msg, x.keyMap.HalfPageDown):
			/*
				lines := m.HalfViewDown()
				if m.HighPerformanceRendering {
					cmd = ViewDown(m, lines)
				}
			*/

		case key.Matches(msg, x.keyMap.HalfPageUp):
			/*
				lines := m.HalfViewUp()
				if m.HighPerformanceRendering {
					cmd = ViewUp(m, lines)
				}
			*/

		case key.Matches(msg, x.keyMap.Down):
			/*
				lines := m.LineDown(1)
				if m.HighPerformanceRendering {
					cmd = ViewDown(m, lines)
				}
			*/

		case key.Matches(msg, x.keyMap.Up):
			/*
				lines := m.LineUp(1)
				if m.HighPerformanceRendering {
					cmd = ViewUp(m, lines)
				}
			*/
		}

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			/*
				lines := m.LineUp(m.MouseWheelDelta)
				if m.HighPerformanceRendering {
					cmd = ViewUp(m, lines)
				}
			*/

		case tea.MouseWheelDown:
			/*
				lines := m.LineDown(m.MouseWheelDelta)
				if m.HighPerformanceRendering {
					cmd = ViewDown(m, lines)
				}
			*/
		}
	}

	return *x, nil
}

func (x *Model) height() int {
	return len(x.lines)
}

func (x *Model) cursorValue() any {
	if x.cursor == nil {
		return nil
	}
	return x.cursor.Value
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
		out := x.format.String()
		if x.cursorValue() == item {
			out = "* " + out
		}
		return out

	case directory.DirectorySubspace:
		x.format.Reset()
		x.format.Directory(convert.FromStringArray(val.GetPath()))
		out := x.format.String()
		if x.cursorValue() == item {
			out = "* " + out
		}
		return out

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
