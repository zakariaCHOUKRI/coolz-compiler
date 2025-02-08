package parser

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
	"strconv"
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

// func (p *Parser) expectCurrent(t lexer.TokenType) bool {
// 	if p.curTokenIs(t) {
// 		p.nextToken()
// 		return true
// 	}
// 	p.currentError(t)
// 	return false
// }

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

	// Check if current token is CLASS
	if !p.curTokenIs(lexer.CLASS) {
		p.currentError(lexer.CLASS)
		return nil
	}
	p.nextToken() // Move to class name

	// Parse class name
	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}
	c.Name = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken() // Move past class name

	// Check for inheritance
	if p.curTokenIs(lexer.INHERITS) {
		p.nextToken()
		if !p.curTokenIs(lexer.TYPEID) {
			p.currentError(lexer.TYPEID)
			return nil
		}
		c.Parent = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}

	// Parse class body
	if !p.curTokenIs(lexer.LBRACE) {
		p.currentError(lexer.LBRACE)
		return nil
	}
	p.nextToken()

	// Parse features
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		feature := p.parseFeature()
		c.Features = append(c.Features, feature)
		if !p.curTokenIs(lexer.SEMI) {
			p.currentError(lexer.SEMI)
			return nil
		}
		p.nextToken()
	}

	if !p.curTokenIs(lexer.RBRACE) {
		p.currentError(lexer.RBRACE)
		return nil
	}
	p.nextToken()

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

	// Store the name token before advancing
	nameToken := p.curToken

	// Parse method name (OBJECTID)
	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}
	m.Name = &ast.ObjectIdentifier{Token: nameToken, Value: nameToken.Literal}
	p.nextToken()

	// Parse opening parenthesis
	if !p.curTokenIs(lexer.LPAREN) {
		p.currentError(lexer.LPAREN)
		return nil
	}
	p.nextToken()

	// Parse formals (parameters)
	m.Formals = p.parseFormals()

	// Parse closing parenthesis
	if !p.curTokenIs(lexer.RPAREN) {
		p.currentError(lexer.RPAREN)
		return nil
	}
	p.nextToken()

	// Parse colon
	if !p.curTokenIs(lexer.COLON) {
		p.currentError(lexer.COLON)
		return nil
	}
	p.nextToken()

	// Parse return type
	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}
	m.ReturnType = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// Parse method body
	if !p.curTokenIs(lexer.LBRACE) {
		p.currentError(lexer.LBRACE)
		return nil
	}
	p.nextToken()

	// Parse body expression
	m.Body = p.parseExpression()

	// Parse closing brace
	if !p.curTokenIs(lexer.RBRACE) {
		p.currentError(lexer.RBRACE)
		return nil
	}
	p.nextToken()

	return m
}

func (p *Parser) parseAttribute() *ast.Attribute {
	a := &ast.Attribute{Token: p.curToken}

	// Parse attribute name (we're already at the identifier)
	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}
	a.Name = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// Parse colon
	if !p.curTokenIs(lexer.COLON) {
		p.currentError(lexer.COLON)
		return nil
	}
	p.nextToken()

	// Parse type
	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}
	a.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// Handle initialization if present
	if p.curTokenIs(lexer.ASSIGN) {
		p.nextToken()
		a.Init = p.parseExpression()
	}

	return a
}

func (p *Parser) parseFormals() []*ast.Formal {
	var formals []*ast.Formal

	// Empty parameter list
	if p.curTokenIs(lexer.RPAREN) {
		return formals
	}

	// First parameter
	formal := p.parseFormal()
	if formal != nil {
		formals = append(formals, formal)
	}

	// Additional parameters
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken()
		formal := p.parseFormal()
		if formal != nil {
			formals = append(formals, formal)
		}
	}

	return formals
}

func (p *Parser) parseFormal() *ast.Formal {
	formal := &ast.Formal{Token: p.curToken}

	// Parse parameter name
	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}
	formal.Name = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// Parse colon
	if !p.curTokenIs(lexer.COLON) {
		p.currentError(lexer.COLON)
		return nil
	}
	p.nextToken()

	// Parse type
	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}
	formal.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	return formal
}

func (p *Parser) parseExpression() ast.Expression {
	var left ast.Expression

	switch p.curToken.Type {
	case lexer.INT_CONST:
		val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
		left = &ast.IntegerLiteral{Token: p.curToken, Value: val}
		p.nextToken()
	case lexer.STR_CONST:
		left = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	case lexer.TRUE, lexer.FALSE:
		left = &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Type == lexer.TRUE}
		p.nextToken()
	case lexer.OBJECTID:
		left = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	case lexer.NOT:
		left = p.parseNotExpression()
	case lexer.NEG:
		left = &ast.UnaryExpression{
			Token:    p.curToken,
			Operator: "~",
			Right:    p.parseExpression(),
		}
	case lexer.IF:
		left = p.parseIfExpression()
	case lexer.WHILE:
		left = p.parseWhileExpression()
	case lexer.NEW:
		left = p.parseNewExpression()
	case lexer.ISVOID:
		left = p.parseIsVoidExpression()
	case lexer.LBRACE:
		left = p.parseBlockExpression()
	case lexer.LET:
		left = p.parseLetExpression()
	case lexer.LPAREN:
		p.nextToken()
		left = p.parseExpression()
		if !p.expectCurrent(lexer.RPAREN) {
			return nil
		}
	}

	// Parse binary expressions if we have an operator next
	precedence := p.curPrecedence()
	for !p.curTokenIs(lexer.SEMI) && !p.curTokenIs(lexer.EOF) && precedence < p.peekPrecedence() {
		left = p.parseBinaryExpression(precedence)
	}

	return left
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
		value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
		if err != nil {
			p.errors = append(p.errors, fmt.Sprintf("could not parse integer literal: %v", err))
			return nil
		}
		return &ast.IntegerLiteral{Token: p.curToken, Value: value}
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
