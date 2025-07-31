// Package facade provide interfaces for FDB APIs.
package facade

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
)

type (
	// ReadTransactor provides methods for performing read transactions and for opening
	// & listing directories.
	ReadTransactor interface {
		// ReadTransact opens a new transaction. If this ReadTransactor is backed by a
		// transaction then that transaction is reused.
		ReadTransact(func(ReadTransaction) (interface{}, error)) (interface{}, error)

		// DirOpen opens a directory under the root directory specified by the
		// implementation.
		DirOpen(path []string) (directory.DirectorySubspace, error)

		// DirList lists the directories under the root directory specified by
		// the implementation.
		DirList(path []string) ([]string, error)
	}

	// ReadTransaction provides methods for reading key-values from an open transaction.
	ReadTransaction interface {
		ReadTransactor

		// Get requests the values bytes for the given key.
		Get(key fdb.KeyConvertible) fdb.FutureByteSlice

		// GetRange performs a range-read over the given range.
		GetRange(r fdb.Range, options fdb.RangeOptions) fdb.RangeResult
	}

	// Transactor provides methods for performing read or write transactions and for
	// creating, opening, & listing directories.
	Transactor interface {
		ReadTransactor

		// Transact opens a new transaction. If this Transactor is backed by a
		// transaction then that transaction is reused.
		Transact(func(Transaction) (interface{}, error)) (interface{}, error)

		// DirCreateOrOpen opens a directory (or creates it if it doesn't exist)
		// under the root directory specified by the implementation.
		DirCreateOrOpen(path []string) (directory.DirectorySubspace, error)
	}

	// Transaction provides methods for reading or writing key-values from an open
	// transaction.
	Transaction interface {
		ReadTransaction
		Transactor

		// Set writes a key-value.
		Set(fdb.KeyConvertible, []byte)

		// Set writes a key-value. The key should contain
		// the placeholder for a single versionstamp as
		// described here:
		// https://pkg.go.dev/github.com/apple/foundationdb/bindings/go/src/fdb#Transaction.SetVersionstampedKey
		SetWithVStampKey(fdb.KeyConvertible, []byte)

		// TOOD
		SetWithVStampValue(fdb.KeyConvertible, []byte)

		// Clear deletes a key-value.
		Clear(fdb.KeyConvertible)

		// Watch creates a watch and returns a FutureNil that will become ready when the
		// watch reports a change to the value of the specified key.
		Watch(fdb.KeyConvertible) fdb.FutureNil
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

// NewReadTransactor creates a new instance of a ReadTransactor backed by a fdb.ReadTransactor.
// Any directory operations performed by the returned ReadTransactor will use the given
// directory.Directory as the root.
func NewReadTransactor(tr fdb.ReadTransactor, root directory.Directory) ReadTransactor {
	return &readTransactor{tr, root}
}

// NewReadTransaction creates a new instance of a ReadTransaction backed by a fdb.ReadTransaction.
// Any directory operations performed by the returned ReadTransaction will use the given
// directory.Directory as the root.
func NewReadTransaction(tr fdb.ReadTransaction, root directory.Directory) ReadTransaction {
	return &readTransaction{tr, root}
}

// NewTransactor creates a new instance of a Transactor backed by a fdb.Transactor.
// Any directory operations performed by the returned Transactor will use the given
// directory.Directory as the root.
func NewTransactor(tr fdb.Transactor, root directory.Directory) Transactor {
	return &transactor{NewReadTransactor(tr, root), tr, root}
}

// NewTransaction creates a new instance of a Transaction backed by a fdb.Transaction.
// Any directory operations performed by the returned Transaction will use the given
// directory.Directory as the root.
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

func (x *transaction) SetWithVStampKey(key fdb.KeyConvertible, val []byte) {
	x.tr.SetVersionstampedKey(key, val)
}

func (x *transaction) SetWithVStampValue(key fdb.KeyConvertible, val []byte) {
	x.tr.SetVersionstampedValue(key, val)
}

func (x *transaction) Clear(key fdb.KeyConvertible) {
	x.tr.Clear(key)
}

func (x *transaction) Watch(key fdb.KeyConvertible) fdb.FutureNil {
	return x.tr.Watch(key)
}
