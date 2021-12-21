package facade

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
)

type (
	nilReadTransactor struct{}

	nilReadTransaction struct{}

	nilTransactor struct {
		ReadTransactor
	}

	nilTransaction struct {
		ReadTransaction
	}
)

var (
	_ ReadTransactor  = &nilReadTransactor{}
	_ ReadTransaction = &nilReadTransaction{}
	_ Transactor      = &nilTransactor{}
	_ Transaction     = &nilTransaction{}
)

func NewNilReadTransactor() ReadTransactor {
	return &nilReadTransactor{}
}

func NewNilReadTransaction() ReadTransaction {
	return &nilReadTransaction{}
}

func NewNilTransactor() Transactor {
	return &nilTransactor{NewNilReadTransactor()}
}

func NewNilTransaction() Transaction {
	return &nilTransaction{NewNilReadTransaction()}
}

func (x *nilReadTransactor) ReadTransact(f func(ReadTransaction) (interface{}, error)) (interface{}, error) {
	return f(NewNilReadTransaction())
}

func (x *nilReadTransactor) DirOpen(_ []string) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilReadTransactor) DirList(_ []string) ([]string, error) {
	return nil, nil
}

func (x *nilReadTransaction) ReadTransact(f func(ReadTransaction) (interface{}, error)) (interface{}, error) {
	return f(x)
}

func (x *nilReadTransaction) DirOpen(_ []string) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilReadTransaction) DirList(_ []string) ([]string, error) {
	return nil, nil
}

func (x *nilReadTransaction) Get(_ fdb.KeyConvertible) fdb.FutureByteSlice {
	return NewNilFutureByteSlice()
}

func (x *nilReadTransaction) GetRange(_ fdb.Range, _ fdb.RangeOptions) fdb.RangeResult {
	return fdb.RangeResult{}
}

func (x *nilTransactor) Transact(f func(Transaction) (interface{}, error)) (interface{}, error) {
	return f(NewNilTransaction())
}

func (x *nilTransactor) DirCreateOrOpen(_ []string) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilTransaction) Transact(f func(Transaction) (interface{}, error)) (interface{}, error) {
	return f(x)
}

func (x *nilTransaction) DirCreateOrOpen(_ []string) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilTransaction) Set(_ fdb.KeyConvertible, _ []byte) {}

func (x *nilTransaction) Clear(_ fdb.KeyConvertible) {}
