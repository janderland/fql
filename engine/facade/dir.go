package facade

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
)

type (
	nilDirectory struct{}

	nilDirectorySubspace struct {
		subspace.Subspace
		directory.Directory
	}
)

var (
	_ directory.Directory         = &nilDirectory{}
	_ directory.DirectorySubspace = &nilDirectorySubspace{}
)

// NewNilDirectory returns a nil implementation of directory.Directory
// where every operation is a no-op.
func NewNilDirectory() directory.Directory {
	return &nilDirectory{}
}

// NewNilDirectorySubspace returns a nil implementation of directory.DirectorySubspace
// where every operation is a no-op.
func NewNilDirectorySubspace() directory.DirectorySubspace {
	return &nilDirectorySubspace{NewNilSubspace(), NewNilDirectory()}
}

func (x *nilDirectory) CreateOrOpen(_ fdb.Transactor, _ []string, _ []byte) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilDirectory) Open(_ fdb.ReadTransactor, _ []string, _ []byte) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilDirectory) Create(_ fdb.Transactor, _ []string, _ []byte) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilDirectory) CreatePrefix(_ fdb.Transactor, _ []string, _ []byte, _ []byte) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilDirectory) Move(_ fdb.Transactor, _ []string, _ []string) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilDirectory) MoveTo(_ fdb.Transactor, _ []string) (directory.DirectorySubspace, error) {
	return NewNilDirectorySubspace(), nil
}

func (x *nilDirectory) Remove(_ fdb.Transactor, _ []string) (bool, error) {
	return false, nil
}

func (x *nilDirectory) Exists(_ fdb.ReadTransactor, _ []string) (bool, error) {
	return false, nil
}

func (x *nilDirectory) List(_ fdb.ReadTransactor, _ []string) ([]string, error) {
	return nil, nil
}

func (x *nilDirectory) GetLayer() []byte {
	return nil
}

func (x *nilDirectory) GetPath() []string {
	return nil
}
