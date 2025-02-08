package parser

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectAndPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) expectCurrent(t lexer.TokenType) bool {
	if p.curTokenIs(t) {
		p.nextToken()
		return true
	}
	p.currentError(t)
	return false
}

func (p *Parser) peekError(t lexer.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("Expected next token to be %v, got %v line %d col %d", t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column))
}

func (p *Parser) currentError(t lexer.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("Expected current token to be %v, got %v line %d col %d", t, p.curToken.Type, p.peekToken.Line, p.peekToken.Column))
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}
	for p.curToken.Type != lexer.EOF && p.curToken.Type != lexer.ERROR {
		c := p.ParseClass()
		if c != nil {
			prog.Classes = append(prog.Classes, c)
		}
		p.nextToken()
	}
	return prog
}

func (p *Parser) ParseClass() *ast.Class {
	c := &ast.Class{Token: p.curToken}
	if !p.expectCurrent(lexer.CLASS) {
		return nil
	}

	c.Name = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectAndPeek(lexer.TYPEID) {
		return nil
	}

	if p.peekTokenIs(lexer.INHERITS) {
		p.nextToken()
		c.Parent = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectAndPeek(lexer.TYPEID) {
			return nil
		}
	}

	if !p.expectAndPeek(lexer.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		feature := p.parseFeature()
		if feature != nil {
			c.Features = append(c.Features, feature)
		}
		if !p.expectAndPeek(lexer.SEMI) {
			return nil
		}
	}

	if !p.expectAndPeek(lexer.RBRACE) {
		return nil
	}

	return c
}

func (p *Parser) parseFeature() ast.Feature {
	if p.peekTokenIs(lexer.LPAREN) {
		return p.parseMethod()
	}
	return p.parseAttribute()
}

func (p *Parser) parseMethod() *ast.Method {
	m := &ast.Method{Token: p.curToken}
	m.Name = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectAndPeek(lexer.OBJECTID) {
		return nil
	}

	if !p.expectAndPeek(lexer.LPAREN) {
		return nil
	}

	m.Formals = p.parseFormals()

	if !p.expectAndPeek(lexer.RPAREN) {
		return nil
	}

	if !p.expectAndPeek(lexer.COLON) {
		return nil
	}

	m.ReturnType = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectAndPeek(lexer.TYPEID) {
		return nil
	}

	if !p.expectAndPeek(lexer.LBRACE) {
		return nil
	}

	m.Body = p.parseExpression()

	if !p.expectAndPeek(lexer.RBRACE) {
		return nil
	}

	return m
}

func (p *Parser) parseAttribute() *ast.Attribute {
	a := &ast.Attribute{Token: p.curToken}
	a.Name = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectAndPeek(lexer.OBJECTID) {
		return nil
	}

	if !p.expectAndPeek(lexer.COLON) {
		return nil
	}

	a.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectAndPeek(lexer.TYPEID) {
		return nil
	}

	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken()
		a.Init = p.parseExpression()
	}

	return a
}

func (p *Parser) parseFormals() []*ast.Formal {
	formals := []*ast.Formal{}
	for !p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		formal := &ast.Formal{Token: p.curToken}
		formal.Name = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectAndPeek(lexer.OBJECTID) {
			return nil
		}

		if !p.expectAndPeek(lexer.COLON) {
			return nil
		}

		formal.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectAndPeek(lexer.TYPEID) {
			return nil
		}

		formals = append(formals, formal)

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken()
		}
	}
	return formals
}

func (p *Parser) parseExpression() ast.Expression {
	switch p.curToken.Type {
	case lexer.INT_CONST:
		return &ast.IntegerLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.STR_CONST:
		return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.BOOL_CONST:
		return &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Literal == "true"}
	case lexer.OBJECTID:
		return &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.LPAREN:
		p.nextToken()
		exp := p.parseExpression()
		if !p.expectAndPeek(lexer.RPAREN) {
			return nil
		}
		return exp
	case lexer.IF:
		return p.parseIfExpression()
	case lexer.WHILE:
		return p.parseWhileExpression()
	case lexer.LET:
		return p.parseLetExpression()
	case lexer.NEW:
		return p.parseNewExpression()
	case lexer.ISVOID:
		return p.parseIsVoidExpression()
	case lexer.NOT:
		return p.parseNotExpression()
	case lexer.LBRACE:
		return p.parseBlockExpression()
	default:
		return p.parseBinaryExpression(0)
	}
}

func (p *Parser) parseIfExpression() *ast.Conditional {
	exp := &ast.Conditional{Token: p.curToken}
	if !p.expectAndPeek(lexer.IF) {
		return nil
	}

	exp.Condition = p.parseExpression()

	if !p.expectAndPeek(lexer.THEN) {
		return nil
	}

	exp.ThenBranch = p.parseExpression()

	if !p.expectAndPeek(lexer.ELSE) {
		return nil
	}

	exp.ElseBranch = p.parseExpression()

	if !p.expectAndPeek(lexer.FI) {
		return nil
	}

	return exp
}

func (p *Parser) parseWhileExpression() *ast.Loop {
	exp := &ast.Loop{Token: p.curToken}
	if !p.expectAndPeek(lexer.WHILE) {
		return nil
	}

	exp.Condition = p.parseExpression()

	if !p.expectAndPeek(lexer.LOOP) {
		return nil
	}

	exp.Body = p.parseExpression()

	if !p.expectAndPeek(lexer.POOL) {
		return nil
	}

	return exp
}

func (p *Parser) parseLetExpression() *ast.Let {
	exp := &ast.Let{Token: p.curToken}
	if !p.expectAndPeek(lexer.LET) {
		return nil
	}

	for {
		decl := &ast.LetDeclaration{Token: p.curToken}
		decl.Name = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectAndPeek(lexer.OBJECTID) {
			return nil
		}

		if !p.expectAndPeek(lexer.COLON) {
			return nil
		}

		decl.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectAndPeek(lexer.TYPEID) {
			return nil
		}

		if p.peekTokenIs(lexer.ASSIGN) {
			p.nextToken()
			decl.Init = p.parseExpression()
		}

		exp.Declarations = append(exp.Declarations, decl)

		if !p.peekTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken()
	}

	if !p.expectAndPeek(lexer.IN) {
		return nil
	}

	exp.Body = p.parseExpression()

	return exp
}

func (p *Parser) parseNewExpression() *ast.New {
	exp := &ast.New{Token: p.curToken}
	if !p.expectAndPeek(lexer.NEW) {
		return nil
	}

	exp.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectAndPeek(lexer.TYPEID) {
		return nil
	}

	return exp
}

func (p *Parser) parseIsVoidExpression() *ast.IsVoid {
	exp := &ast.IsVoid{Token: p.curToken}
	if !p.expectAndPeek(lexer.ISVOID) {
		return nil
	}

	exp.Expr = p.parseExpression()

	return exp
}

func (p *Parser) parseNotExpression() *ast.UnaryExpression {
	exp := &ast.UnaryExpression{Token: p.curToken, Operator: lexer.NOT}
	if !p.expectAndPeek(lexer.NOT) {
		return nil
	}

	exp.Right = p.parseExpression()

	return exp
}

func (p *Parser) parseBlockExpression() *ast.Block {
	exp := &ast.Block{Token: p.curToken}
	if !p.expectAndPeek(lexer.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		exp.Expressions = append(exp.Expressions, p.parseExpression())
		if !p.expectAndPeek(lexer.SEMI) {
			return nil
		}
	}

	if !p.expectAndPeek(lexer.RBRACE) {
		return nil
	}

	return exp
}

func (p *Parser) parseBinaryExpression(precedence int) ast.Expression {
	leftExp := p.parseUnaryExpression()

	for !p.peekTokenIs(lexer.SEMI) && precedence < p.peekPrecedence() {
		p.nextToken()
		operator := p.curToken.Type
		rightExp := p.parseBinaryExpression(p.curPrecedence())
		leftExp = &ast.BinaryExpression{
			Token:    p.curToken,
			Left:     leftExp,
			Operator: operator,
			Right:    rightExp,
		}
	}

	return leftExp
}

func (p *Parser) parseUnaryExpression() ast.Expression {
	switch p.curToken.Type {
	case lexer.NOT, lexer.NEG:
		operator := p.curToken.Type
		p.nextToken()
		return &ast.UnaryExpression{
			Token:    p.curToken,
			Operator: operator,
			Right:    p.parseUnaryExpression(),
		}
	default:
		return p.parsePrimaryExpression()
	}
}

func (p *Parser) parsePrimaryExpression() ast.Expression {
	switch p.curToken.Type {
	case lexer.INT_CONST:
		return &ast.IntegerLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.STR_CONST:
		return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.BOOL_CONST:
		return &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Literal == "true"}
	case lexer.OBJECTID:
		return &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.LPAREN:
		p.nextToken()
		exp := p.parseExpression()
		if !p.expectAndPeek(lexer.RPAREN) {
			return nil
		}
		return exp
	default:
		return nil
	}
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

var precedences = map[lexer.TokenType]int{
	lexer.EQ:     EQUALS,
	lexer.LT:     LESSGREATER,
	lexer.LE:     LESSGREATER,
	lexer.PLUS:   SUM,
	lexer.MINUS:  SUM,
	lexer.TIMES:  PRODUCT,
	lexer.DIVIDE: PRODUCT,
	lexer.ASSIGN: ASSIGN,
	lexer.DARROW: ASSIGN,
	lexer.NOT:    PREFIX,
	lexer.NEG:    PREFIX,
	lexer.LPAREN: CALL,
	lexer.LBRACE: CALL,
	lexer.DOT:    CALL,
	lexer.AT:     CALL,
}

const (
	LOWEST = iota + 1
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	ASSIGN
)
