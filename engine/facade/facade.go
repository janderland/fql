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
		tr   fdb.ReadTransactor
		root directory.Directory
	}

	readTransaction struct {
		tr   fdb.ReadTransaction
		root directory.Directory
	}

	transactor struct {
		ReadTransactor
		tr   fdb.Transactor
		root directory.Directory
	}

	transaction struct {
		ReadTransaction
		tr   fdb.Transaction
		root directory.Directory
	}
)

var (
	_ ReadTransactor  = &readTransactor{}
	_ ReadTransaction = &readTransaction{}
	_ Transactor      = &transactor{}
	_ Transaction     = &transaction{}
)

func NewReadTransactor(tr fdb.ReadTransactor, root directory.Directory) ReadTransactor {
	return &readTransactor{tr, root}
}

func NewReadTransaction(tr fdb.ReadTransaction, root directory.Directory) ReadTransaction {
	return &readTransaction{tr, root}
}

func NewTransactor(tr fdb.Transactor, root directory.Directory) Transactor {
	return &transactor{NewReadTransactor(tr, root), tr, root}
}

func NewTransaction(tr fdb.Transaction, root directory.Directory) Transaction {
	return &transaction{NewReadTransaction(tr, root), tr, root}
}

func (x *readTransactor) ReadTransact(f func(ReadTransaction) (interface{}, error)) (interface{}, error) {
	return x.tr.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return f(NewReadTransaction(tr, x.root))
	})
}

func (x *readTransactor) DirOpen(path []string) (directory.DirectorySubspace, error) {
	return x.root.Open(x.tr, path, nil)
}

func (x *readTransactor) DirList(path []string) ([]string, error) {
	return x.root.List(x.tr, path)
}

func (x *readTransaction) ReadTransact(f func(ReadTransaction) (interface{}, error)) (interface{}, error) {
	return x.tr.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return f(NewReadTransaction(tr, x.root))
	})
}

func (x *readTransaction) DirOpen(path []string) (directory.DirectorySubspace, error) {
	return x.root.Open(x.tr, path, nil)
}

func (x *readTransaction) DirList(path []string) ([]string, error) {
	return x.root.List(x.tr, path)
}

func (x *readTransaction) Get(key fdb.KeyConvertible) fdb.FutureByteSlice {
	return x.tr.Get(key)
}

func (x *readTransaction) GetRange(rng fdb.Range, options fdb.RangeOptions) fdb.RangeResult {
	return x.tr.GetRange(rng, options)
}

func (x *transactor) Transact(f func(Transaction) (interface{}, error)) (interface{}, error) {
	return x.tr.Transact(func(tr fdb.Transaction) (interface{}, error) {
		return f(NewTransaction(tr, x.root))
	})
}

func (x *transactor) DirCreateOrOpen(path []string) (directory.DirectorySubspace, error) {
	return x.root.CreateOrOpen(x.tr, path, nil)
}

func (x *transaction) Transact(f func(Transaction) (interface{}, error)) (interface{}, error) {
	return x.tr.Transact(func(tr fdb.Transaction) (interface{}, error) {
		return f(NewTransaction(tr, x.root))
	})
}

func (x *transaction) DirCreateOrOpen(path []string) (directory.DirectorySubspace, error) {
	return x.root.CreateOrOpen(x.tr, path, nil)
}

func (x *transaction) Set(key fdb.KeyConvertible, val []byte) {
	x.tr.Set(key, val)
}

func (x *transaction) Clear(key fdb.KeyConvertible) {
	x.tr.Clear(key)
}
