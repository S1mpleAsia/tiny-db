package parse

import (
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

/*
TinyDB Grammars:

<Field> 		:= IdTok
<Constant>		:= StrTok | IntTok
<Expression>	:= <Field> | <Constant>
<Term>			:= <Expression> = <Expression>
<Predicate>		:= <Term> [ AND <Predicate> ]

<Query>			:= SELECT <SelectList> FROM <TableList> [ WHERE <Predicate> ]
<SelectList>	:= <Field> [ , <SelectList> ]
<TableList>		:= IdTok [ , <TableList> ]

<UpdateCmd>		:= <Insert> | <Delete> | <Modify> | <Create>
<Create>		:= <CreateTable> | <CreateView> | <CreateIndex>

<Insert>		:= INSERT INTO IdTok ( <FieldList> ) VALUES ( <ConstList> )
<FieldList>		:= <Field> [ , <FieldList> ]
<ConstList>		:= <Constant> [ , <ConstList> ]

<Delete>		:= DELETE FROM IdTok [ WHERE <Predicate> ]

<Modify>		:= UPDATE IdTok SET <Field> = <Expression> [ WHERE <Predicate> ]

<CreateTable>	:= CREATE TABLE IdTok ( <FieldDefs> )
<FieldDefs>		:= <FieldDef> [ , <FieldDefs> ]
<FieldDef>		:= IdTok <TypeDef>
<TypeDef>		:= INT | VARCHAR (IntTok)

<CREATEVIEW>	:= CREATE VIEW IdTok AS <Query>
<CREATEINDEX>	:= CREATE INDEX IdTok ON IdTok ( <Field> )
*/
type Parser struct {
	lex *Lexer
}

func NewParser(input string) (*Parser, error) {
	lex, err := NewLexer(input)
	if err != nil {
		return nil, err
	}

	return &Parser{
		lex: lex,
	}, nil
}

func (p *Parser) Field() (string, error) {
	return p.lex.EatIdentifier()
}

func (p *Parser) Constant() (*record.Constant, error) {
	if p.lex.MatchStringConstant() {
		value, err := p.lex.EatStringConstant()
		if err != nil {
			return nil, err
		}

		return record.NewConstantWithString(value), nil
	} else {
		value, err := p.lex.EatIntConstant()
		if err != nil {
			return nil, err
		}

		return record.NewConstantWithInt(value), nil
	}
}

func (p *Parser) Expression() (*query.Expression, error) {
	if p.lex.MatchIdentifier() {
		fieldName, err := p.Field()
		if err != nil {
			return nil, err
		}

		return query.NewExpressionWithField(fieldName), nil
	} else {
		value, err := p.Constant()
		if err != nil {
			return nil, err
		}

		return query.NewExpressionWithConstant(value), nil
	}
}

func (p *Parser) Term() (*query.Term, error) {
	lhs, err := p.Expression()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim('='); err != nil {
		return nil, err
	}

	rhs, err := p.Expression()
	if err != nil {
		return nil, err
	}
	return query.NewTerm(lhs, rhs), nil
}

func (p *Parser) Predicate() (*query.Predicate, error) {
	term, err := p.Term()
	if err != nil {
		return nil, err
	}

	pred := query.NewPredicateWithTerm(term)

	if p.lex.MatchKeyWord("and") {
		if err := p.lex.EatKeyword("and"); err != nil {
			return nil, err
		}

		rhs, err := p.Predicate()
		if err != nil {
			return nil, err
		}

		pred.ConjoinWith(rhs)
	}

	return pred, nil
}

func (p *Parser) Query() (*QueryData, error) {
	if err := p.lex.EatKeyword("select"); err != nil {
		return nil, err
	}

	fields, err := p.selectList()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatKeyword("from"); err != nil {
		return nil, err
	}

	tables, err := p.tableList()
	if err != nil {
		return nil, err
	}

	pred, err := p.whereOpt()
	if err != nil {
		return nil, err
	}

	return NewQueryData(fields, tables, pred), nil
}

func (p *Parser) whereOpt() (*query.Predicate, error) {
	if p.lex.MatchKeyWord("where") {
		if err := p.lex.EatKeyword("where"); err != nil {
			return nil, err
		}

		pred, err := p.Predicate()
		if err != nil {
			return nil, err
		}

		return pred, nil
	} else {
		return query.NewPredicate(), nil
	}
}

// <Query>			:= SELECT <SelectList> FROM <TableList> [ WHERE <Predicate> ]
// <SelectList>		:= <Field> [ , <SelectList> ]
func (p *Parser) selectList() ([]string, error) {
	field, err := p.Field()
	if err != nil {
		return nil, err
	}

	fields := []string{field}

	if p.lex.MatchDelim(',') {
		if err := p.lex.EatDelim(','); err != nil {
			return nil, err
		}

		if !p.lex.MatchIdentifier() {
			return fields, nil
		}

		rest, err := p.selectList()
		if err != nil {
			return nil, err
		}

		fields = append(fields, rest...)
	}

	return fields, nil
}

// <TableList>		:= IdTok [ , <TableList> ]
func (p *Parser) tableList() ([]string, error) {
	table, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	tables := []string{table}

	if p.lex.MatchDelim(',') {
		if err := p.lex.EatDelim(','); err != nil {
			return nil, err
		}

		if !p.lex.MatchIdentifier() {
			return tables, nil
		}

		rest, err := p.tableList()
		if err != nil {
			return nil, err
		}

		tables = append(tables, rest...)
	}

	return tables, nil
}

// <UpdateCmd> := <Insert> | <Delete> | <Modify> | <Create>
func (p *Parser) UpdateCmd() (UpdateCmd, error) {
	if p.lex.MatchKeyWord("insert") {
		// <Insert>
		return p.Insert()
	} else if p.lex.MatchKeyWord("delete") {
		// <Delete>
		return p.Delete()
	} else if p.lex.MatchKeyWord("update") {
		// <Modify>
		return p.Modify()
	} else {
		// <Create>
		return p.create()
	}
}

func (p *Parser) create() (UpdateCmd, error) {
	if err := p.lex.EatKeyword("create"); err != nil {
		return nil, err
	}

	if p.lex.MatchKeyWord("table") {
		return p.CreateTable()
	} else if p.lex.MatchKeyWord("view") {
		return p.CreateView()
	} else {
		return p.CreateIndex()
	}
}

func (p *Parser) Delete() (*DeleteData, error) {
	if err := p.lex.EatKeyword("delete"); err != nil {
		return nil, err
	}

	if err := p.lex.EatKeyword("from"); err != nil {
		return nil, err
	}

	tableName, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	pred, err := p.whereOpt()
	if err != nil {
		return nil, err
	}

	return NewDeleteData(tableName, pred), nil
}

func (p *Parser) Insert() (*InsertData, error) {
	if err := p.lex.EatKeyword("insert"); err != nil {
		return nil, err
	}

	if err := p.lex.EatKeyword("into"); err != nil {
		return nil, err
	}

	tableName, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim('('); err != nil {
		return nil, err
	}

	fields, err := p.fieldList()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim(')'); err != nil {
		return nil, err
	}

	if err := p.lex.EatKeyword("values"); err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim('('); err != nil {
		return nil, err
	}

	values, err := p.constList()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim(')'); err != nil {
		return nil, err
	}

	return NewInsertData(tableName, fields, values), nil
}

func (p *Parser) fieldList() ([]string, error) {
	field, err := p.Field()
	if err != nil {
		return nil, err
	}

	fields := []string{field}

	if p.lex.MatchDelim(',') {
		if err := p.lex.EatDelim(','); err != nil {
			return nil, err
		}

		rest, err := p.fieldList()
		if err != nil {
			return nil, err
		}

		fields = append(fields, rest...)
	}

	return fields, nil
}

func (p *Parser) constList() ([]*record.Constant, error) {
	val, err := p.Constant()

	if err != nil {
		return nil, err
	}

	values := []*record.Constant{val}

	if p.lex.MatchDelim(',') {
		if err := p.lex.EatDelim(','); err != nil {
			return nil, err
		}

		rest, err := p.constList()
		if err != nil {
			return nil, err
		}

		values = append(values, rest...)
	}

	return values, nil
}

// <Modify> := UPDATE IdTok SET <Field> = <Expression> [ WHERE <Predicate> ]
func (p *Parser) Modify() (*ModifyData, error) {
	if err := p.lex.EatKeyword("update"); err != nil {
		return nil, err
	}

	tableName, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatKeyword("set"); err != nil {
		return nil, err
	}

	fieldName, err := p.Field()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim('='); err != nil {
		return nil, err
	}

	newValue, err := p.Expression()
	if err != nil {
		return nil, err
	}

	pred, err := p.whereOpt()
	if err != nil {
		return nil, err
	}

	return NewModifyData(tableName, fieldName, newValue, pred), nil
}

// <CreateTable> := CREATE TABLE IdTok ( <FieldDefs> )
func (p *Parser) CreateTable() (*CreateTableData, error) {
	if err := p.lex.EatKeyword("table"); err != nil {
		return nil, err
	}

	tableName, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim('('); err != nil {
		return nil, err
	}

	schema, err := p.fieldDefs()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim(')'); err != nil {
		return nil, err
	}

	return NewCreateTableData(tableName, schema), nil
}

func (p *Parser) fieldDefs() (*record.Schema, error) {
	sch, err := p.fieldDef()
	if err != nil {
		return nil, err
	}

	if p.lex.MatchDelim(',') {
		if err := p.lex.EatDelim(','); err != nil {
			return nil, err
		}

		rest, err := p.fieldDefs()
		if err != nil {
			return nil, err
		}

		sch.AddAll(rest)
	}

	return sch, nil
}

func (p *Parser) fieldDef() (*record.Schema, error) {
	fieldName, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	schema, err := p.fieldType(fieldName)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (p *Parser) fieldType(fieldName string) (*record.Schema, error) {
	schema := record.NewSchema()

	if p.lex.MatchKeyWord("int") {
		if err := p.lex.EatKeyword("int"); err != nil {
			return nil, err
		}

		schema.AddIntField(fieldName)
	} else {
		if err := p.lex.EatKeyword("varchar"); err != nil {
			return nil, err
		}

		if err := p.lex.EatDelim('('); err != nil {
			return nil, err
		}

		strLen, err := p.lex.EatIntConstant()
		if err != nil {
			return nil, err
		}
		if err := p.lex.EatDelim(')'); err != nil {
			return nil, err
		}

		schema.AddStringField(fieldName, strLen)
	}

	return schema, nil
}

// <CreateView> := CREATE VIEW IdTok AS <Query>
func (p *Parser) CreateView() (*CreateViewData, error) {
	if err := p.lex.EatKeyword("view"); err != nil {
		return nil, err
	}

	viewName, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatKeyword("as"); err != nil {
		return nil, err
	}

	query, err := p.Query()
	if err != nil {
		return nil, err
	}

	return NewCreateViewData(viewName, query), nil
}

// <CreateIndex> := CREATE INDEX IdTok ON IdTok ( <Field> )
func (p *Parser) CreateIndex() (*CreateIndexData, error) {
	if err := p.lex.EatKeyword("index"); err != nil {
		return nil, err
	}

	indexName, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatKeyword("on"); err != nil {
		return nil, err
	}

	tableName, err := p.lex.EatIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim('('); err != nil {
		return nil, err
	}

	fieldName, err := p.Field()
	if err != nil {
		return nil, err
	}

	if err := p.lex.EatDelim(')'); err != nil {
		return nil, err
	}

	return NewCreateIndexData(indexName, tableName, fieldName), nil
}
