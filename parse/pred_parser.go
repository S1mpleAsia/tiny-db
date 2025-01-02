package parse

type PredParser struct {
	lex *Lexer
}

func NewPredParser(input string) (*PredParser, error) {
	lex, err := NewLexer(input)
	if err != nil {
		return nil, err
	}

	return &PredParser{
		lex: lex,
	}, nil
}

// <Field> := IdTok
func (p *PredParser) Field() error {
	if _, err := p.lex.EatIdentifier(); err != nil {
		return err
	}

	return nil
}

// <Constant> := StrTok | IntTok
func (p *PredParser) Constant() error {
	if p.lex.MatchStringConstant() {
		if _, err := p.lex.EatStringConstant(); err != nil {
			return err
		}
	} else {
		if _, err := p.lex.EatIntConstant(); err != nil {
			return err
		}
	}

	return nil
}

// <Expression>	:= <Field> | <Constant>
func (p *PredParser) Expression() error {
	if p.lex.MatchIdentifier() {
		if err := p.Field(); err != nil {
			return err
		}
	} else {
		if err := p.Constant(); err != nil {
			return err
		}
	}

	return nil
}

// <Term> := <Expression> = <Expression>
func (p *PredParser) Term() error {
	if err := p.Expression(); err != nil {
		return err
	}

	if err := p.lex.EatDelim('='); err != nil {
		return err
	}

	if err := p.Expression(); err != nil {
		return err
	}

	return nil
}

// <Predicate> := <Term> [ AND <Predicate> ]
func (p *PredParser) Predicate() error {
	if err := p.Term(); err != nil {
		return err
	}

	if p.lex.MatchKeyWord("and") {
		if err := p.lex.EatKeyword("and"); err != nil {
			return err
		}

		if err := p.Predicate(); err != nil {
			return err
		}
	}

	return nil
}
