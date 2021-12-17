package facade

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
)

type (
	ReadTransactor interface {
		ReadTransact(func(ReadTransaction) (interface{}, error)) (interface{}, error)
		DirOpen(path []string) (directory.DirectorySubspace, error)
		DirList(path []string) ([]string, error)
	}

	ReadTransaction interface {
		ReadTransactor
		Get(key fdb.KeyConvertible) fdb.FutureByteSlice
		GetRange(r fdb.Range, options fdb.RangeOptions) fdb.RangeResult
	}

	Transactor interface {
		ReadTransactor
		Transact(func(Transaction) (interface{}, error)) (interface{}, error)
		DirCreateOrOpen(path []string) (directory.DirectorySubspace, error)
	}

	Transaction interface {
		ReadTransaction
		Transactor
		Set(fdb.KeyConvertible, []byte)
		Clear(fdb.KeyConvertible)
	}
)

type (
	readTransactor struct {
		tr fdb.ReadTransactor
	}

	readTransaction struct {
		tr fdb.ReadTransaction
	}

	transactor struct {
		ReadTransactor
		tr fdb.Transactor
	}

	transaction struct {
		ReadTransaction
		tr fdb.Transaction
	}
)

func NewReadTransactor(tr fdb.ReadTransactor) ReadTransactor {
	return &readTransactor{tr}
}

func NewReadTransaction(tr fdb.ReadTransaction) ReadTransaction {
	return &readTransaction{tr}
}

func NewTransactor(tr fdb.Transactor) Transactor {
	return &transactor{&readTransactor{tr}, tr}
}

func NewTransaction(tr fdb.Transaction) Transaction {
	return &transaction{&readTransaction{tr}, tr}
}

var (
	_ ReadTransactor  = &readTransactor{}
	_ ReadTransaction = &readTransaction{}
	_ Transactor      = &transactor{}
	_ Transaction     = &transaction{}
)

func (x *readTransactor) ReadTransact(f func(ReadTransaction) (interface{}, error)) (interface{}, error) {
	return x.tr.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return f(NewReadTransaction(tr))
	})
}

func (x *readTransactor) DirOpen(path []string) (directory.DirectorySubspace, error) {
	return directory.Open(x.tr, path, nil)
}

func (x *readTransactor) DirList(path []string) ([]string, error) {
	return directory.List(x.tr, path)
}

func (x *readTransaction) ReadTransact(f func(ReadTransaction) (interface{}, error)) (interface{}, error) {
	return x.tr.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return f(NewReadTransaction(tr))
	})
}

func (x *readTransaction) DirOpen(path []string) (directory.DirectorySubspace, error) {
	return directory.Open(x.tr, path, nil)
}

func (x *readTransaction) DirList(path []string) ([]string, error) {
	return directory.List(x.tr, path)
}

func (x *readTransaction) Get(key fdb.KeyConvertible) fdb.FutureByteSlice {
	return x.tr.Get(key)
}

func (x *readTransaction) GetRange(rng fdb.Range, options fdb.RangeOptions) fdb.RangeResult {
	return x.tr.GetRange(rng, options)
}

func (x *transactor) Transact(f func(Transaction) (interface{}, error)) (interface{}, error) {
	return x.tr.Transact(func(tr fdb.Transaction) (interface{}, error) {
		return f(NewTransaction(tr))
	})
}

func (x *transactor) DirCreateOrOpen(path []string) (directory.DirectorySubspace, error) {
	return directory.CreateOrOpen(x.tr, path, nil)
}

func (x *transaction) Transact(f func(Transaction) (interface{}, error)) (interface{}, error) {
	return x.tr.Transact(func(tr fdb.Transaction) (interface{}, error) {
		return f(NewTransaction(tr))
	})
}

func (x *transaction) DirCreateOrOpen(path []string) (directory.DirectorySubspace, error) {
	return directory.CreateOrOpen(x.tr, path, nil)
}

func (x *transaction) Set(key fdb.KeyConvertible, val []byte) {
	x.tr.Set(key, val)
}

func (x *transaction) Clear(key fdb.KeyConvertible) {
	x.tr.Clear(key)
}
