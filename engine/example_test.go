package engine_test

import (
	"encoding/binary"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"

	"github.com/janderland/fdbq/engine"
	"github.com/janderland/fdbq/engine/facade"
	"github.com/janderland/fdbq/keyval"
)

func Example() {
	eg := engine.New(facade.NewTransactor(fdb.MustOpenDefault(), directory.Root()))

	key := keyval.Key{
		Directory: keyval.Directory{keyval.String("hello"), keyval.String("there")},
		Tuple:     keyval.Tuple{keyval.Float(33.3)},
	}

	// /hello/there{33.3}=10
	query := keyval.KeyValue{Key: key, Value: keyval.Int(10)}
	if err := eg.Set(query, binary.BigEndian); err != nil {
		panic(err)
	}

	didWrite, err := eg.Transact(func(e engine.Engine) (interface{}, error) {
		// /hello/there{33.3}=<>
		query = keyval.KeyValue{Key: key, Value: keyval.Variable{}}
		result, err := eg.ReadSingle(query, engine.SingleOpts{ByteOrder: binary.BigEndian})
		if err != nil {
			return nil, err
		}
		if result != nil {
			return false, nil
		}

		// /hello/there{33.3}=15
		query = keyval.KeyValue{Key: key, Value: keyval.Int(15)}
		if err := eg.Set(query, binary.BigEndian); err != nil {
			return nil, err
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}

	if didWrite != false {
		panic("didWrite should be false")
	}
}
