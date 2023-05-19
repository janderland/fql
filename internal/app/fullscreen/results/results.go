package results

import (
	"container/list"
	"fmt"
	"strings"

	"github.com/janderland/fdbq/engine/stream"
	"github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/janderland/fdbq/parser/format"
)

type Model struct {
	list  *list.List
	lines []string
}

func New() Model {
	return Model{list: list.New()}
}

func (x *Model) Reset() {
	x.list = list.New()
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
	if len(x.lines) == 0 {
		return ""
	}

	i := 0
	for item := x.list.Front(); item != nil; item = item.Next() {
		x.lines[i] = view(item.Value)
		i++

		if i == len(x.lines) {
			break
		}
	}

	var results strings.Builder
	for j := i - 1; j >= 0; j-- {
		results.WriteString(x.lines[j])
		results.WriteRune('\n')
	}
	return results.String()
}

func view(item any) string {
	f := format.New(format.Cfg{})

	switch val := item.(type) {
	case error:
		return fmt.Sprintf("ERR! %s", val)

	case string:
		return fmt.Sprintf("# %s", val)

	case keyval.KeyValue:
		f.Reset()
		f.KeyValue(val)
		return f.String()

	case stream.KeyValErr:
		if val.Err != nil {
			return view(val.Err)
		}
		f.Reset()
		f.KeyValue(val.KV)
		return f.String()

	case stream.DirErr:
		if val.Err != nil {
			return view(val.Err)
		}
		f.Reset()
		f.Directory(convert.FromStringArray(val.Dir.GetPath()))
		return f.String()

	default:
		return fmt.Sprintf("ERR! unexpected %T", val)
	}
}
