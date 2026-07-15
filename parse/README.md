# parse — SQL lexer, parser, and AST

Turns a SQL string into typed AST objects. Contains a hand-written lexer, a
recursive-descent parser for a small SQL subset, and the AST structs consumed by
the query planner.

## Responsibilities

- Tokenize SQL input into delimiters, integer/string constants, keywords, and identifiers.
- Parse queries (`SELECT`) and update commands (`INSERT`/`DELETE`/`UPDATE`/`CREATE TABLE`/`CREATE VIEW`/`CREATE INDEX`).
- Parse predicates as AND-conjunctions of equality terms, delegating term/expression/constant construction to the `query` and `record` packages.
- Produce AST value objects (`QueryData` and the `UpdateCmd` implementations) for downstream planning.
- Report parse failures as `BadSyntaxError`.

## Key types

- `Lexer` — streaming tokenizer; keeps one lookahead `token`, exposes `Match*`/`Eat*` methods. Keywords are matched case-insensitively.
- `lexer` — interface describing the `Match*`/`Eat*` contract (`Lexer` satisfies it).
- `Parser` — recursive-descent parser wrapping a `Lexer`; entry points `Query()` and `UpdateCmd()`.
- `PredParser` — standalone predicate-only parser used for validation (no AST is built).
- `QueryData` — parsed `SELECT`: field list, table list, optional `*query.Predicate`.
- `UpdateCmd` — sealed interface (`updateCmd()`) implemented by the update AST types below.
- `InsertData`, `ModifyData`, `DeleteData`, `CreateTableData`, `CreateViewData`, `CreateIndexData` — AST nodes for the update commands.
- `BadSyntaxError` — error type prefixing messages with `"bad syntax: "`.

## Key API

- `NewLexer(input string) (*Lexer, error)` — construct a lexer and read the first token.
- `func (*Lexer) MatchDelim(d rune) bool`, `MatchIntConstant()`, `MatchStringConstant()`, `MatchKeyWord(word string)`, `MatchIdentifier()` — non-consuming lookahead.
- `func (*Lexer) EatDelim(d rune) error`, `EatIntConstant() (int32, error)`, `EatStringConstant() (string, error)`, `EatKeyword(word string) error`, `EatIdentifier() (string, error)` — consume-or-error.
- `NewParser(input string) (*Parser, error)` — build a parser over `input`.
- `func (*Parser) Query() (*QueryData, error)` — parse a `SELECT` statement.
- `func (*Parser) UpdateCmd() (UpdateCmd, error)` — dispatch on the leading keyword to parse an insert/delete/update/create command.
- `func (*Parser) Predicate() (*query.Predicate, error)`, `Term()`, `Expression()`, `Constant()`, `Field()` — grammar sub-rules (exported).
- `func (*Parser) Insert() (*InsertData, error)`, `Delete() (*DeleteData, error)`, `Modify() (*ModifyData, error)`, `CreateTable() (*CreateTableData, error)`, `CreateView() (*CreateViewData, error)`, `CreateIndex() (*CreateIndexData, error)` — per-command parsers.
- `NewPredParser(input string) (*PredParser, error)` / `func (*PredParser) Predicate() error` — validate a predicate without producing an AST.
- Constructors `NewQueryData`, `NewInsertData`, `NewModifyData`, `NewDeleteData`, `NewCreateTableData`, `NewCreateViewData`, `NewCreateIndexData` plus accessor methods on each AST type.

## Supported SQL

```sql
-- Field/expression grammar
<Field>       := IdTok
<Constant>    := StrTok | IntTok
<Expression>  := <Field> | <Constant>
<Term>        := <Expression> = <Expression>
<Predicate>   := <Term> [ AND <Predicate> ]

-- Queries
<Query>       := SELECT <SelectList> FROM <TableList> [ WHERE <Predicate> ]
<SelectList>  := <Field> [ , <SelectList> ]
<TableList>   := IdTok [ , <TableList> ]

-- Update commands
<UpdateCmd>   := <Insert> | <Delete> | <Modify> | <Create>
<Create>      := <CreateTable> | <CreateView> | <CreateIndex>

<Insert>      := INSERT INTO IdTok ( <FieldList> ) VALUES ( <ConstList> )
<FieldList>   := <Field> [ , <FieldList> ]
<ConstList>   := <Constant> [ , <ConstList> ]

<Delete>      := DELETE FROM IdTok [ WHERE <Predicate> ]

<Modify>      := UPDATE IdTok SET <Field> = <Expression> [ WHERE <Predicate> ]

<CreateTable> := CREATE TABLE IdTok ( <FieldDefs> )
<FieldDefs>   := <FieldDef> [ , <FieldDefs> ]
<FieldDef>    := IdTok <TypeDef>
<TypeDef>     := INT | VARCHAR ( IntTok )

<CreateView>  := CREATE VIEW IdTok AS <Query>
<CreateIndex> := CREATE INDEX IdTok ON IdTok ( <Field> )
```

## How it fits

Depends on `record` (constants, schema) and `query` (expressions, terms,
predicates) to build AST nodes. It is driven by the `plan` package: `Planner`
calls `Parser.Query()` / `Parser.UpdateCmd()`, and `BasicQueryPlanner` re-parses
stored view definitions.

## Notes

- The only comparison operator is `=`; predicates are pure AND-conjunctions of equality terms. There is no `OR`, no inequality/range comparison, and no parentheses in predicates.
- No `ORDER BY`, `GROUP BY`, aggregate functions, `JOIN` syntax, `DISTINCT`, `LIMIT`, or subqueries are recognized by the grammar. (Sorting, grouping, and joins exist only as physical plans in the `plan` package, not as parsable SQL.)
- Constants are only integers (`int32`) and single-quoted strings; string literals cannot contain an embedded quote (the lexer stops at the first `'`).
- Identifiers and keywords are lowercased; column/table types are limited to `INT` and `VARCHAR(n)`.
- `selectList` / `tableList` tolerate a trailing comma (a `,` not followed by an identifier ends the list).
- Tests: `lexer_test.go`, `parser_test.go`, `pred_parser_test.go`.
