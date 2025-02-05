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

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	ASSIGN      // <-
	EQUALS      // =
	LESSGREATER // < or <=
	SUM         // + or -
	PRODUCT     // * or /
	PREFIX      // -x or not x or ~x or isvoid x
	CALL        // method(args)
	DOT         // obj.method
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)

	// Register prefix parsers
	p.registerPrefix(lexer.OBJECTID, p.parseIdentifier)
	p.registerPrefix(lexer.INT_CONST, p.parseIntegerLiteral)
	p.registerPrefix(lexer.STR_CONST, p.parseStringLiteral)
	p.registerPrefix(lexer.BOOL_CONST, p.parseBooleanLiteral)
	p.registerPrefix(lexer.IF, p.parseIfExpression)
	p.registerPrefix(lexer.WHILE, p.parseWhileExpression)
	p.registerPrefix(lexer.NEW, p.parseNewExpression)
	p.registerPrefix(lexer.ISVOID, p.parseIsVoidExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.NEG, p.parsePrefixExpression)

	// Register infix parsers
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.DIVIDE, p.parseInfixExpression)
	p.registerInfix(lexer.TIMES, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.LE, p.parseInfixExpression)
	p.registerInfix(lexer.ASSIGN, p.parseAssignExpression)
	p.registerInfix(lexer.DOT, p.parseDotExpression) // add DOT infix parser

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
		c := p.parseClass()

		if !p.expectAndPeek(lexer.SEMI) {
			continue // sf error handled by method aslan
		}
		p.nextToken()
		prog.Classes = append(prog.Classes, c)
	}
	return prog
}

func (p *Parser) parseClass() *ast.Class {

	c := &ast.Class{Token: p.curToken}
	if !p.expectCurrent(lexer.CLASS) {
		return nil
	}

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}

	c.Name = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	p.nextToken()

	if p.curTokenIs(lexer.INHERITS) {
		if !p.expectAndPeek(lexer.TYPEID) {
			return nil
		}
		c.Parent = &ast.TypeIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		p.nextToken()
	}

	if !p.expectCurrent(lexer.LBRACE) {
		return nil
	}
	// do not consume '{'
	for !p.curTokenIs(lexer.RBRACE) {
		c.Features = append(c.Features, p.parseFeature())
		if !p.expectAndPeek(lexer.SEMI) {
			return nil
		}
		p.nextToken()
	}

	if !p.curTokenIs(lexer.RBRACE) {
		p.currentError(lexer.RBRACE)
		return nil
	}
	p.nextToken() // move past '}'

	return c
}

func (p *Parser) parseFeature() ast.Feature {
	if p.peekTokenIs(lexer.LPAREN) {
		return p.parseMethod()
	}
	return p.parseAttribute()
}

func (p *Parser) parseMethod() *ast.Method {

	method := &ast.Method{
		Name: &ast.ObjectIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
	}

	if !p.expectCurrent(lexer.OBJECTID) {
		return nil
	}

	if !p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		method.Formals = append(method.Formals, p.parseFormals()...)
	}

	if !p.expectAndPeek(lexer.RPAREN) {
		return nil
	}
	if !p.expectAndPeek(lexer.COLON) {
		return nil
	}
	if !p.expectAndPeek(lexer.TYPEID) {
		return nil
	}

	method.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectAndPeek(lexer.LBRACE) {
		return nil
	}

	//method.expression = 	 should i add expressions of a method or what ?

	if !p.expectAndPeek(lexer.RBRACE) {
		return nil
	}

	return method
}

func (p *Parser) parseFormal() *ast.Formal {
	formal := &ast.Formal{
		Name: &ast.ObjectIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
	}
	if !p.expectCurrent(lexer.OBJECTID) {
		return nil
	}

	if !p.expectCurrent(lexer.COLON) {
		return nil
	}

	formal.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}

	return formal
}

func (p *Parser) parseFormals() []*ast.Formal {
	formals := []*ast.Formal{p.parseFormal()}
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		formals = append(formals, p.parseFormal())
	}
	return formals
}

func (p *Parser) parseAttribute() *ast.Attribute {
	attribute := &ast.Attribute{
		Name: &ast.ObjectIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
	}
	if !p.expectCurrent(lexer.OBJECTID) {
		return nil
	}
	if !p.expectCurrent(lexer.COLON) {
		return nil
	}
	attribute.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}

	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // consume '<-'
		p.nextToken() // move to expression
		p.parseExpression(LOWEST)
	}

	return attribute
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.curTokenIs(lexer.SEMI) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
	p.nextToken() // advance to avoid infinite loop
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.ObjectIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	num, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: num}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	str := p.curToken.Literal
	return &ast.StringLiteral{Token: p.curToken, Value: str}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal == "true",
	}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // consume the '('

	exp := p.parseExpression(LOWEST)

	if !p.peekTokenIs(lexer.RPAREN) {
		p.peekError(lexer.RPAREN)
		return nil
	}
	p.nextToken() // consume the ')'
	return exp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.UnaryExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	exp.Right = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.BinaryExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)
	return exp
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	exp := &ast.AssignExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	exp.Right = p.parseExpression(LOWEST)
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.curToken}
	p.nextToken() // consume "if"

	exp.Condition = p.parseExpression(LOWEST)

	if !p.peekTokenIs(lexer.THEN) {
		return nil
	}
	p.nextToken() // consume "then"
	p.nextToken() // move past "then"

	exp.Consequence = p.parseExpression(LOWEST)

	if !p.peekTokenIs(lexer.ELSE) {
		return nil
	}
	p.nextToken() // consume "else"
	p.nextToken() // move past "else"

	exp.Alternative = p.parseExpression(LOWEST)

	if !p.peekTokenIs(lexer.FI) {
		return nil
	}
	p.nextToken() // consume "fi"
	p.nextToken() // move past "fi"

	return exp
}

func (p *Parser) parseWhileExpression() ast.Expression {
	exp := &ast.WhileExpression{Token: p.curToken}
	p.nextToken() // consume "while"

	exp.Condition = p.parseExpression(LOWEST)

	if !p.peekTokenIs(lexer.LOOP) {
		return nil
	}
	p.nextToken() // consume "loop"
	p.nextToken() // move past "loop"

	exp.Body = p.parseExpression(LOWEST)

	if !p.peekTokenIs(lexer.POOL) {
		return nil
	}
	p.nextToken() // consume "pool"
	p.nextToken() // move past "pool"

	return exp
}

func (p *Parser) parseNewExpression() ast.Expression {
	exp := &ast.NewExpression{Token: p.curToken}

	if !p.expectCurrent(lexer.NEW) {
		return nil
	}

	exp.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	return exp
}

func (p *Parser) parseIsVoidExpression() ast.Expression {
	exp := &ast.IsVoidExpression{Token: p.curToken}

	p.nextToken()

	exp.Expression = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	token := p.curToken

	if !p.peekTokenIs(lexer.OBJECTID) {
		p.peekError(lexer.OBJECTID)
		return nil
	}
	p.nextToken() // move to method name

	methodName := &ast.ObjectIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	p.nextToken() // consume method name

	exp := &ast.BinaryExpression{
		Token:    token,
		Operator: ".",
		Left:     left,
		Right:    methodName,
	}

	// Handle method call arguments if present
	if p.curTokenIs(lexer.LPAREN) {
		p.nextToken() // consume '('

		first := true
		for !p.curTokenIs(lexer.RPAREN) && p.curToken.Type != lexer.EOF {
			if !first {
				if !p.curTokenIs(lexer.COMMA) {
					p.currentError(lexer.COMMA)
					return nil
				}
				p.nextToken() // consume comma
			}

			_ = p.parseExpression(LOWEST)
			p.nextToken() // consume the expression
			first = false
		}

		if !p.curTokenIs(lexer.RPAREN) {
			p.currentError(lexer.RPAREN)
			return nil
		}
		p.nextToken() // consume ')'
	}

	return exp
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

var precedences = map[lexer.TokenType]int{
	lexer.ASSIGN: ASSIGN,
	lexer.EQ:     EQUALS,
	lexer.LT:     LESSGREATER,
	lexer.LE:     LESSGREATER,
	lexer.PLUS:   SUM,
	lexer.MINUS:  SUM,
	lexer.DIVIDE: PRODUCT,
	lexer.TIMES:  PRODUCT,
	lexer.LPAREN: CALL,
	lexer.DOT:    DOT,
}
