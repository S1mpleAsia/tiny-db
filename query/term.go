package query

import (
	"fmt"
	"math"

	"s1mpleasia.com/tinydb/record"
)

type Term struct {
	lhs *Expression
	rhs *Expression
}

func NewTerm(lhs *Expression, rhs *Expression) *Term {
	return &Term{lhs, rhs}
}

func (t *Term) IsSatisfied(scan Scan) (bool, error) {
	lhsVal, err := t.lhs.Evaluate(scan)
	if err != nil {
		return false, err
	}

	rhsVal, err := t.rhs.Evaluate(scan)
	if err != nil {
		return false, err
	}

	return rhsVal.Equals(lhsVal), nil
}

func (t *Term) AppliesTo(sch *record.Schema) bool {
	return t.lhs.AppliesTo(sch) && t.rhs.AppliesTo(sch)
}

func (t *Term) String() string {
	return fmt.Sprintf("%s = %s", t.lhs, t.rhs)
}

// Determine if this term is of the form "F=C"
func (t *Term) equatesWithConstant(fieldName string) *record.Constant {
	if t.lhs.IsFieldName() && t.lhs.AsFieldName() == fieldName && !t.rhs.IsFieldName() {
		return t.rhs.AsConstant()
	} else if t.rhs.IsFieldName() && t.rhs.AsFieldName() == fieldName && !t.lhs.IsFieldName() {
		return t.lhs.AsConstant()
	} else {
		return nil
	}
}

// Determine if this term is of the form "F1=F2"
func (t *Term) equatesWithField(fieldName string) string {
	if t.lhs.IsFieldName() && t.lhs.AsFieldName() == fieldName && t.rhs.IsFieldName() {
		return t.rhs.AsFieldName()
	} else if t.rhs.IsFieldName() && t.rhs.AsFieldName() == fieldName && t.lhs.IsFieldName() {
		return t.lhs.AsFieldName()
	} else {
		return ""
	}
}

// The method reductionFactor determines the expected number of records that will sastify the predicate
func (t *Term) reductionFactor(p PlanLike) int32 {
	if t.lhs.IsFieldName() && t.rhs.IsFieldName() {
		lhsName := t.lhs.AsFieldName()
		rhsName := t.rhs.AsFieldName()
		return max(p.DistinctValues(lhsName), p.DistinctValues(rhsName))
	}

	if t.lhs.IsFieldName() {
		return p.DistinctValues(t.lhs.AsFieldName())
	}

	if t.rhs.IsFieldName() {
		return p.DistinctValues(t.rhs.AsFieldName())
	}

	if t.lhs.AsConstant().Equals(t.rhs.AsConstant()) {
		return 1
	}

	return math.MaxInt32
}
