package kind

import (
	q "github.com/janderland/fdbq/keyval/keyval"
)

type dirVisitor struct {
	kind subKind
}

func (x *dirVisitor) VisitString(q.String) {
	// Do nothing.
}

func (x *dirVisitor) VisitVariable(q.Variable) {
	x.kind = variableSubKind
}

type tupVisitor struct {
	kind subKind
}

func (x *tupVisitor) VisitTuple(e q.Tuple) {
	x.kind = tupKind(e)
}

func (x *tupVisitor) VisitVariable(q.Variable) {
	x.kind = variableSubKind
}

func (x *tupVisitor) VisitMaybeMore(q.MaybeMore) {
	x.kind = variableSubKind
}

func (x *tupVisitor) VisitNil(q.Nil) {
	// Do nothing.
}

func (x *tupVisitor) VisitInt(q.Int) {
	// Do nothing.
}

func (x *tupVisitor) VisitUint(q.Uint) {
	// Do nothing.
}

func (x *tupVisitor) VisitBool(q.Bool) {
	// Do nothing.
}

func (x *tupVisitor) VisitFloat(q.Float) {
	// Do nothing.
}

func (x *tupVisitor) VisitBigInt(q.BigInt) {
	// Do nothing.
}

func (x *tupVisitor) VisitString(q.String) {
	// Do nothing.
}

func (x *tupVisitor) VisitUUID(q.UUID) {
	// Do nothing.
}

func (x *tupVisitor) VisitBytes(q.Bytes) {
	// Do nothing.
}

type valVisitor struct {
	kind subKind
}

func (x *valVisitor) VisitTuple(e q.Tuple) {
	x.kind = tupKind(e)
}

func (x *valVisitor) VisitVariable(q.Variable) {
	x.kind = variableSubKind
}

func (x *valVisitor) VisitClear(q.Clear) {
	x.kind = clearSubKind
}

func (x *valVisitor) VisitNil(q.Nil) {
	// Do nothing.
}

func (x *valVisitor) VisitInt(q.Int) {
	// Do nothing.
}

func (x *valVisitor) VisitUint(q.Uint) {
	// Do nothing.
}

func (x *valVisitor) VisitBool(q.Bool) {
	// Do nothing.
}

func (x *valVisitor) VisitFloat(q.Float) {
	// Do nothing.
}

func (x *valVisitor) VisitString(q.String) {
	// Do nothing.
}

func (x *valVisitor) VisitUUID(q.UUID) {
	// Do nothing.
}

func (x *valVisitor) VisitBytes(q.Bytes) {
	// Do nothing.
}
