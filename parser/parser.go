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
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
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
	p.registerPrefix(lexer.CASE, p.parseCaseExpression)
	p.registerPrefix(lexer.LET, p.parseLetExpression)
	p.registerPrefix(lexer.SELF, p.parseSelf)
	p.registerPrefix(lexer.VOID, p.parseVoidLiteral)

	// Register infix parsers
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.DIVIDE, p.parseInfixExpression)
	p.registerInfix(lexer.TIMES, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.LE, p.parseInfixExpression)
	p.registerInfix(lexer.ASSIGN, p.parseAssignment)
	p.registerInfix(lexer.DOT, p.parseDynamicDispatch)
	p.registerInfix(lexer.AT, p.parseStaticDispatch)

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

	// Parse formals
	if !p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		method.Formals = p.parseFormals()
	}

	if !p.expectAndPeek(lexer.RPAREN) || !p.expectAndPeek(lexer.COLON) || !p.expectAndPeek(lexer.TYPEID) {
		return nil
	}

	method.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectAndPeek(lexer.LBRACE) {
		return nil
	}

	// Parse method body
	p.nextToken()
	method.Body = p.parseExpression(LOWEST)

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

	if !p.expectCurrent(lexer.OBJECTID) || !p.expectCurrent(lexer.COLON) {
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

	// Parse optional initialization
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // consume ':'
		p.nextToken() // consume '='
		p.nextToken() // move to init expr
		attribute.Init = p.parseExpression(LOWEST)
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
		// p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMI) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	num, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: num}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Literal == "true"}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectCurrent(lexer.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseCaseExpression() ast.Expression {
	exp := &ast.CaseExpression{Token: p.curToken}

	if !p.expectCurrent(lexer.CASE) {
		return nil
	}

	exp.Expr = p.parseExpression(LOWEST)

	if !p.expectCurrent(lexer.OF) {
		return nil
	}

	for !p.curTokenIs(lexer.ESAC) {
		branch := &ast.CaseBranch{
			Identifier: &ast.ObjectIdentifier{Token: p.curToken},
			Type:       &ast.TypeIdentifier{Token: p.curToken},
		}

		if !p.expectCurrent(lexer.OBJECTID) || !p.expectCurrent(lexer.COLON) || !p.expectCurrent(lexer.TYPEID) {
			return nil
		}

		branch.Type.Value = p.curToken.Literal
		p.nextToken() // consume '=>'
		branch.Expr = p.parseExpression(LOWEST)
		exp.Branches = append(exp.Branches, branch)

		if !p.curTokenIs(lexer.SEMI) {
			break
		}
		p.nextToken()
	}

	if !p.expectCurrent(lexer.ESAC) {
		return nil
	}
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

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.curToken}

	if !p.expectCurrent(lexer.IF) {
		return nil
	}

	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectCurrent(lexer.THEN) {
		return nil
	}

	exp.Consequence = p.parseExpression(LOWEST)

	if !p.expectCurrent(lexer.ELSE) {
		return nil
	}

	exp.Alternative = p.parseExpression(LOWEST)

	if !p.expectCurrent(lexer.FI) {
		return nil
	}

	return exp
}

func (p *Parser) parseWhileExpression() ast.Expression {
	exp := &ast.WhileExpression{Token: p.curToken}

	if !p.expectCurrent(lexer.WHILE) {
		return nil
	}

	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectCurrent(lexer.LOOP) {
		return nil
	}

	exp.Body = p.parseExpression(LOWEST)

	if !p.expectCurrent(lexer.POOL) {
		return nil
	}

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
	lexer.EQ:     EQUALS,
	lexer.LT:     LESSGREATER,
	lexer.LE:     LESSGREATER,
	lexer.PLUS:   SUM,
	lexer.MINUS:  SUM,
	lexer.DIVIDE: PRODUCT,
	lexer.TIMES:  PRODUCT,
	lexer.LPAREN: CALL,
	lexer.ASSIGN: LOWEST,
	lexer.DOT:    CALL,
	lexer.AT:     CALL,
}

func (p *Parser) parseLetExpression() ast.Expression {
	exp := &ast.LetExpression{Token: p.curToken}

	if !p.expectCurrent(lexer.LET) {
		return nil
	}

	// Parse bindings
	for {
		binding := &ast.LetBinding{
			Identifier: &ast.ObjectIdentifier{Token: p.curToken},
		}

		if !p.expectCurrent(lexer.OBJECTID) || !p.expectCurrent(lexer.COLON) || !p.expectCurrent(lexer.TYPEID) {
			return nil
		}

		binding.Type = &ast.TypeIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}

		// Parse optional initialization
		if p.peekTokenIs(lexer.ASSIGN) {
			p.nextToken() // consume ':'
			p.nextToken() // consume '='
			p.nextToken() // move to init expr
			binding.Init = p.parseExpression(LOWEST)
		}

		exp.Bindings = append(exp.Bindings, binding)

		// Check for more bindings
		if !p.peekTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume ','
		p.nextToken() // move to next binding
	}

	if !p.expectCurrent(lexer.IN) {
		return nil
	}

	p.nextToken() // move to body expr
	exp.In = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseAssignment(left ast.Expression) ast.Expression {
	exp := &ast.Assignment{
		Token: p.curToken,
		Left:  left, // The left-hand side of the assignment
	}

	if !p.expectCurrent(lexer.ASSIGN) {
		return nil
	}

	p.nextToken() // Move to the value expression
	exp.Value = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseDynamicDispatch(left ast.Expression) ast.Expression {
	exp := &ast.DynamicDispatch{
		Token:  p.curToken,
		Object: left,
	}

	if !p.expectCurrent(lexer.DOT) {
		return nil
	}

	p.nextToken() // move to method name
	exp.Method = &ast.ObjectIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectCurrent(lexer.OBJECTID) || !p.expectCurrent(lexer.LPAREN) {
		return nil
	}

	// Parse arguments
	if !p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	}

	if !p.expectCurrent(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseStaticDispatch(left ast.Expression) ast.Expression {
	exp := &ast.StaticDispatch{
		Token:  p.curToken,
		Object: left,
	}

	if !p.expectCurrent(lexer.AT) {
		return nil
	}

	p.nextToken() // move to type
	exp.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectCurrent(lexer.TYPEID) || !p.expectCurrent(lexer.DOT) {
		return nil
	}

	p.nextToken() // move to method name
	exp.Method = &ast.ObjectIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectCurrent(lexer.OBJECTID) || !p.expectCurrent(lexer.LPAREN) {
		return nil
	}

	// Parse arguments
	if !p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	}

	if !p.expectCurrent(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseSelf() ast.Expression {
	return &ast.Self{Token: p.curToken}
}

func (p *Parser) parseVoidLiteral() ast.Expression {
	return &ast.VoidLiteral{Token: p.curToken}
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	var args []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume ','
		p.nextToken() // move to next expr
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectCurrent(end) {
		return nil
	}

	return args
}
