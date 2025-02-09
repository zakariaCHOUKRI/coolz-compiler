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
	switch p.curToken.Type {
	case lexer.OBJECTID:
		if p.peekTokenIs(lexer.LPAREN) {
			return p.parseMethod()
		}
		return p.parseAttribute()
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token: %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseMethod() *ast.Method {
	method := &ast.Method{Token: p.curToken}

	// Parse method name
	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}
	method.Name = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// Parse parameter list
	if !p.curTokenIs(lexer.LPAREN) {
		p.currentError(lexer.LPAREN)
		return nil
	}
	p.nextToken()

	method.Formals = p.parseFormals()

	if !p.curTokenIs(lexer.RPAREN) {
		p.currentError(lexer.RPAREN)
		return nil
	}
	p.nextToken()

	// Parse return type
	if !p.curTokenIs(lexer.COLON) {
		p.currentError(lexer.COLON)
		return nil
	}
	p.nextToken()

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}
	method.ReturnType = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// Parse method body
	if !p.curTokenIs(lexer.LBRACE) {
		p.currentError(lexer.LBRACE)
		return nil
	}
	p.nextToken()

	// Handle empty method body
	if p.curTokenIs(lexer.RBRACE) {
		method.Body = &ast.Block{
			Token:       p.curToken,
			Expressions: []ast.Expression{},
		}
		p.nextToken()
		return method
	}

	// Parse method body expressions
	method.Body = p.parseBlockExpression()

	return method
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
	left := p.parsePrefixExpression()
	if left == nil {
		return nil
	}

	for !p.peekTokenIs(lexer.SEMI) && !p.peekTokenIs(lexer.EOF) {
		if p.peekTokenIs(lexer.RPAREN) || p.peekTokenIs(lexer.RBRACE) {
			break
		}

		precedence := p.curPrecedence()
		if precedence >= p.peekPrecedence() {
			break
		}

		if !p.expectAndPeek(p.peekToken.Type) {
			return nil
		}

		switch p.curToken.Type {
		case lexer.NOT:
			left = p.parseNotExpression()
		case lexer.PLUS, lexer.MINUS, lexer.TIMES, lexer.DIVIDE, lexer.LT, lexer.LE, lexer.EQ:
			left = &ast.BinaryExpression{
				Token:    p.curToken,
				Left:     left,
				Operator: p.curToken.Type,
				Right:    p.parsePrefixExpression(),
			}
		case lexer.ASSIGN:
			if id, ok := left.(*ast.ObjectIdentifier); ok {
				left = &ast.Assignment{
					Token: p.curToken,
					Left:  id,
					Value: p.parseExpression(),
				}
			} else {
				p.errors = append(p.errors, "left side of assignment must be an identifier")
				return nil
			}
		}
	}

	return left
}

func (p *Parser) parseIfExpression() *ast.Conditional {
	conditional := &ast.Conditional{Token: p.curToken}

	p.nextToken() // move past 'if'

	conditional.Predicate = p.parseExpression()

	if !p.curTokenIs(lexer.THEN) {
		p.currentError(lexer.THEN)
		return nil
	}
	p.nextToken() // move past 'then'

	conditional.ThenBranch = p.parseExpression()

	if !p.curTokenIs(lexer.ELSE) {
		p.currentError(lexer.ELSE)
		return nil
	}
	p.nextToken() // move past 'else'

	conditional.ElseBranch = p.parseExpression()

	if !p.curTokenIs(lexer.FI) {
		p.currentError(lexer.FI)
		return nil
	}
	p.nextToken() // move past 'fi'

	return conditional
}

// Update parseWhileExpression to handle WHILE, LOOP, and POOL correctly
func (p *Parser) parseWhileExpression() *ast.Loop {
	exp := &ast.Loop{Token: p.curToken}
	p.nextToken() // Consume WHILE

	exp.Condition = p.parseExpression()

	if !p.curTokenIs(lexer.LOOP) {
		p.currentError(lexer.LOOP)
		return nil
	}
	p.nextToken() // Consume LOOP

	exp.Body = p.parseExpression()

	if !p.curTokenIs(lexer.POOL) {
		p.currentError(lexer.POOL)
		return nil
	}
	p.nextToken() // Consume POOL

	return exp
}

func (p *Parser) parseLetExpression() *ast.Let {
	letExp := &ast.Let{Token: p.curToken}

	p.nextToken() // move past 'let'

	// Parse variable name
	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}
	letExp.VarName = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// Parse type annotation
	if !p.curTokenIs(lexer.COLON) {
		p.currentError(lexer.COLON)
		return nil
	}
	p.nextToken()

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}
	letExp.VarType = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	// Parse optional initialization
	if p.curTokenIs(lexer.ASSIGN) {
		p.nextToken()
		letExp.VarInit = p.parseExpression()
	}

	if !p.curTokenIs(lexer.IN) {
		p.currentError(lexer.IN)
		return nil
	}
	p.nextToken()

	letExp.Body = p.parseExpression()

	return letExp
}

// Update parseNewExpression to correctly handle NEW and TYPEID
func (p *Parser) parseNewExpression() *ast.New {
	exp := &ast.New{Token: p.curToken}
	p.nextToken() // Consume NEW

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}
	exp.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken() // Consume TYPEID
	return exp
}

// Update parseIsVoidExpression to consume the ISVOID token
func (p *Parser) parseIsVoidExpression() *ast.IsVoid {
	exp := &ast.IsVoid{Token: p.curToken}
	p.nextToken() // Consume ISVOID
	exp.Expr = p.parseExpression()
	return exp
}

// Update parseNotExpression to use the token's literal
func (p *Parser) parseNotExpression() ast.Expression {
	token := p.curToken
	if !p.expectAndPeek(p.peekToken.Type) {
		return nil
	}

	return &ast.UnaryExpression{
		Token:    token,
		Operator: lexer.NOT,
		Right:    p.parseExpression(),
	}
}

func (p *Parser) parseBlockExpression() *ast.Block {
	block := &ast.Block{Token: p.curToken}
	block.Expressions = []ast.Expression{}

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		expr := p.parseExpression()
		if expr != nil {
			block.Expressions = append(block.Expressions, expr)
		}

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

	return block
}

func (p *Parser) parseBinaryExpression(precedence int) ast.Expression {
	left := p.parsePrefixExpression()
	if left == nil {
		return nil
	}

	for !p.peekTokenIs(lexer.SEMI) && precedence < p.peekPrecedence() {
		operatorToken := p.peekToken
		if !p.expectAndPeek(operatorToken.Type) {
			return nil
		}

		operator := operatorToken.Type
		curPrecedence := p.curPrecedence()

		right := p.parseBinaryExpression(curPrecedence)
		if right == nil {
			return nil
		}

		left = &ast.BinaryExpression{
			Token:    operatorToken,
			Left:     left,
			Operator: operator,
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	switch p.curToken.Type {
	case lexer.IF:
		return p.parseIfExpression()
	case lexer.WHILE:
		return p.parseWhileExpression()
	case lexer.LET:
		return p.parseLetExpression()
	case lexer.LBRACE:
		return p.parseBlockExpression()
	case lexer.NEW:
		return p.parseNewExpression()
	case lexer.ISVOID:
		return p.parseIsVoidExpression()
	case lexer.NOT, lexer.NEG:
		return p.parseUnaryExpression()
	default:
		return p.parsePrimaryExpression()
	}
}

func (p *Parser) parseUnaryExpression() ast.Expression {
	operator := p.curToken.Type // store token type
	tok := p.curToken
	p.nextToken()

	return &ast.UnaryExpression{
		Token:    tok,
		Operator: operator, // operator is lexer.TokenType
		Right:    p.parseBinaryExpression(PREFIX),
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
		if p.peekTokenIs(lexer.ASSIGN) {
			name := &ast.ObjectIdentifier{
				Token: p.curToken,
				Value: p.curToken.Literal,
			}
			assignToken := p.peekToken
			p.nextToken() // consume OBJECTID
			p.nextToken() // consume ASSIGN
			right := p.parseExpression()
			return &ast.Assignment{
				Token: assignToken,
				Left:  name,
				Value: right,
			}
		}
		return &ast.ObjectIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
	case lexer.LPAREN:
		p.nextToken()
		expr := p.parseExpression()
		if !p.curTokenIs(lexer.RPAREN) {
			p.currentError(lexer.RPAREN)
			return nil
		}
		p.nextToken() // consume ')'
		return expr
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token in primary expression: %s", p.curToken.Type))
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

// Update precedences to ensure proper operator precedence
var precedences = map[lexer.TokenType]int{
	lexer.EQ:     EQUALS,
	lexer.LT:     LESSGREATER,
	lexer.LE:     LESSGREATER,
	lexer.PLUS:   SUM,
	lexer.MINUS:  SUM,
	lexer.TIMES:  PRODUCT,
	lexer.DIVIDE: PRODUCT,
	lexer.ASSIGN: ASSIGN,
	lexer.NOT:    PREFIX,
	lexer.NEG:    PREFIX,
	lexer.LPAREN: CALL,
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
