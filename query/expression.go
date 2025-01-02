package query

import "s1mpleasia.com/tinydb/record"

// Represent an expression in predicate that can be evaluated by constant or fieldName
type Expression struct {
	val       *record.Constant
	fieldName *string
}

func NewExpressionWithConstant(val *record.Constant) *Expression {
	return &Expression{val: val}
}

func NewExpressionWithField(fieldName string) *Expression {
	return &Expression{fieldName: &fieldName}
}

func (e *Expression) Evaluate(scan Scan) (*record.Constant, error) {
	if e.val != nil {
		return e.val, nil
	}

	return scan.GetVal(*e.fieldName)
}

func (e *Expression) IsFieldName() bool {
	return e.fieldName != nil
}

func (e *Expression) AsConstant() *record.Constant {
	return e.val
}

func (e *Expression) AsFieldName() string {
	return *e.fieldName
}

func (e *Expression) AppliesTo(sch *record.Schema) bool {
	if e.val != nil {
		return true
	}

	return sch.HasField(*e.fieldName)
}

func (e *Expression) String() string {
	if e.val != nil {
		return e.val.String()
	}

	return *e.fieldName
}
