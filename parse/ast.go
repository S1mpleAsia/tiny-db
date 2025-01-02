package parse

import (
	"fmt"
	"strings"

	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

type QueryData struct {
	fields []string
	tables []string
	pred   *query.Predicate
}

func NewQueryData(fields, tables []string, pred *query.Predicate) *QueryData {
	return &QueryData{
		fields: fields,
		tables: tables,
		pred:   pred,
	}
}

func (q *QueryData) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "select %s from %s", strings.Join(q.fields, ", "), strings.Join(q.tables, ", "))

	if pred := q.pred.String(); pred != "" {
		fmt.Fprintf(&sb, " where %s", pred)
	}

	return sb.String()
}

func (q *QueryData) Fields() []string {
	return q.fields
}

func (q *QueryData) Tables() []string {
	return q.tables
}

func (q *QueryData) Predicate() *query.Predicate {
	return q.pred
}

// UpdateCmd
type UpdateCmd interface {
	updateCmd()
}

func (*InsertData) updateCmd()      {}
func (*ModifyData) updateCmd()      {}
func (*DeleteData) updateCmd()      {}
func (*CreateTableData) updateCmd() {}
func (*CreateViewData) updateCmd()  {}
func (*CreateIndexData) updateCmd() {}

// -----
type InsertData struct {
	tableName string
	fields    []string
	values    []*record.Constant
}

func NewInsertData(tableName string, fields []string, values []*record.Constant) *InsertData {
	return &InsertData{
		tableName: tableName,
		fields:    fields,
		values:    values,
	}
}

func (insertData *InsertData) TableName() string {
	return insertData.tableName
}

func (insertData *InsertData) Fields() []string {
	return insertData.fields
}

func (insertData *InsertData) Values() []*record.Constant {
	return insertData.values
}

// -----
type ModifyData struct {
	tableName   string
	targetField string
	newValue    *query.Expression
	pred        *query.Predicate
}

func NewModifyData(tableName string, targetField string, newValue *query.Expression, pred *query.Predicate) *ModifyData {
	return &ModifyData{
		tableName:   tableName,
		targetField: targetField,
		newValue:    newValue,
		pred:        pred,
	}
}

func (md *ModifyData) TableName() string {
	return md.tableName
}

func (md *ModifyData) TargetField() string {
	return md.targetField
}

func (md *ModifyData) NewValue() *query.Expression {
	return md.newValue
}

func (md *ModifyData) Pred() *query.Predicate {
	return md.pred
}

// -----
type DeleteData struct {
	tableName string
	pred      *query.Predicate
}

func NewDeleteData(tableName string, pred *query.Predicate) *DeleteData {
	return &DeleteData{
		tableName: tableName,
		pred:      pred,
	}
}

func (dd *DeleteData) TableName() string {
	return dd.tableName
}

func (dd *DeleteData) Pred() *query.Predicate {
	return dd.pred
}

// -----
type CreateTableData struct {
	tableName string
	newSchema *record.Schema
}

func NewCreateTableData(tableName string, newSchema *record.Schema) *CreateTableData {
	return &CreateTableData{
		tableName: tableName,
		newSchema: newSchema,
	}
}

func (td *CreateTableData) TableName() string {
	return td.tableName
}

func (td *CreateTableData) NewSchema() *record.Schema {
	return td.newSchema
}

// -----
type CreateViewData struct {
	viewName  string
	queryData *QueryData
}

func NewCreateViewData(viewName string, queryData *QueryData) *CreateViewData {
	return &CreateViewData{
		viewName:  viewName,
		queryData: queryData,
	}
}

func (cv *CreateViewData) ViewDef() string {
	return cv.queryData.String()
}

// -----
type CreateIndexData struct {
	indexName string
	tableName string
	fieldName string
}

func NewCreateIndexData(indexName string, tableName string, fieldName string) *CreateIndexData {
	return &CreateIndexData{
		indexName: indexName,
		tableName: tableName,
		fieldName: fieldName,
	}
}

func (ci *CreateIndexData) IndexName() string {
	return ci.indexName
}

func (ci *CreateIndexData) TableName() string {
	return ci.tableName
}

func (ci *CreateIndexData) FieldName() string {
	return ci.fieldName
}
