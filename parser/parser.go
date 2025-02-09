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

// func (p *Parser) expectAndPeek(t lexer.TokenType) bool {
// 	if p.peekTokenIs(t) {
// 		p.nextToken()
// 		return true
// 	}
// 	p.peekError(t)
// 	return false
// }

func (p *Parser) expectCurrent(t lexer.TokenType) bool {
	if p.curTokenIs(t) {
		p.nextToken()
		return true
	}
	p.currentError(t)
	return false
}

// func (p *Parser) peekError(t lexer.TokenType) {
// 	p.errors = append(p.errors, fmt.Sprintf("Expected next token to be %v, got %v line %d col %d", t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column))
// }

func (p *Parser) currentError(t lexer.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("Expected current token to be %v, got %v line %d col %d", t, p.curToken.Type, p.peekToken.Line, p.peekToken.Column))
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}
	for p.curToken.Type == lexer.CLASS {
		class := p.parseClass()
		if class != nil {
			prog.Classes = append(prog.Classes, class)
		} else {
			break
		}
	}
	return prog
}

func (p *Parser) parseClass() *ast.Class {
	token := p.curToken
	p.nextToken()
	name := &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	var parent *ast.TypeIdentifier
	if p.curToken.Type == lexer.INHERITS {
		p.nextToken()
		parent = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}
	if p.curToken.Type == lexer.LBRACE {
		p.nextToken()
	}
	var features []ast.Feature
	for p.curToken.Type == lexer.OBJECTID {
		f := p.parseFeature()
		if f != nil {
			features = append(features, f)
		}
	}
	if p.curToken.Type == lexer.RBRACE {
		p.nextToken()
	}
	if p.curToken.Type == lexer.SEMI {
		p.nextToken()
	}
	return &ast.Class{
		Token:    token,
		Name:     name,
		Parent:   parent,
		Features: features,
	}
}

func (p *Parser) parseFeature() ast.Feature {
	id := &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	if p.curToken.Type == lexer.LPAREN {
		return p.parseMethod(id)
	}
	return p.parseAttribute(id)
}

func (p *Parser) parseMethod(name *ast.ObjectIdentifier) *ast.Method {
	p.nextToken()
	formals := p.parseFormals()
	if p.curToken.Type == lexer.RPAREN {
		p.nextToken()
	}
	if p.curToken.Type == lexer.COLON {
		p.nextToken()
	}
	typ := &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	if p.curToken.Type == lexer.LBRACE {
		p.nextToken()
	}
	body := p.parseExpression(LOWEST)
	if p.curToken.Type == lexer.RBRACE {
		p.nextToken()
	}
	if p.curToken.Type == lexer.SEMI {
		p.nextToken()
	}
	return &ast.Method{
		Name:    name,
		Type:    typ,
		Formals: formals,
		Body:    body,
	}
}

func (p *Parser) parseAttribute(name *ast.ObjectIdentifier) *ast.Attribute {
	var typ *ast.TypeIdentifier
	if p.curToken.Type == lexer.COLON {
		p.nextToken()
		typ = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}
	var init ast.Expression
	if p.curToken.Type == lexer.ASSIGN {
		p.nextToken()
		init = p.parseExpression(LOWEST)
	}
	if p.curToken.Type == lexer.SEMI {
		p.nextToken()
	}
	return &ast.Attribute{
		Name: name,
		Type: typ,
		Init: init,
	}
}

func (p *Parser) parseFormals() []*ast.Formal {
	var formals []*ast.Formal
	for p.curToken.Type == lexer.OBJECTID {
		n := &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		if p.curToken.Type == lexer.COLON {
			p.nextToken()
		}
		t := &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		formals = append(formals, &ast.Formal{Name: n, Type: t})
		if p.curToken.Type == lexer.COMMA {
			p.nextToken()
			continue
		} else {
			break
		}
	}
	return formals
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
		return nil
	}
	leftExp := prefix()
	for precedence < p.peekPrecedence() {
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
	ident := &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	return ident
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	num, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		return nil
	}
	lit := &ast.IntegerLiteral{Token: p.curToken, Value: num}
	p.nextToken()
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	lit := &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	return lit
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	lit := &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Literal == "true"}
	p.nextToken()
	return lit
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
	p.nextToken() // skip 'case'
	exp.Expr = p.parseExpression(LOWEST)
	if !p.expectCurrent(lexer.OF) {
		return nil
	}
	for !p.curTokenIs(lexer.ESAC) {
		branch := &ast.CaseBranch{}
		idTok := p.curToken
		if !p.expectCurrent(lexer.OBJECTID) {
			return nil
		}
		branch.Identifier = &ast.ObjectIdentifier{Token: idTok, Value: idTok.Literal}
		if !p.expectCurrent(lexer.COLON) {
			return nil
		}
		typeTok := p.curToken
		if !p.expectCurrent(lexer.TYPEID) {
			return nil
		}
		branch.Type = &ast.TypeIdentifier{Token: typeTok, Value: typeTok.Literal}
		if !p.expectCurrent(lexer.DARROW) {
			return nil
		}
		branch.Expr = p.parseExpression(LOWEST)
		if p.curTokenIs(lexer.SEMI) {
			p.nextToken()
		}
		exp.Branches = append(exp.Branches, branch)
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
	for {
		binding := &ast.LetBinding{}
		idTok := p.curToken
		if !p.expectCurrent(lexer.OBJECTID) {
			return nil
		}
		binding.Identifier = &ast.ObjectIdentifier{Token: idTok, Value: idTok.Literal}
		if !p.expectCurrent(lexer.COLON) {
			return nil
		}
		typeTok := p.curToken
		if !p.expectCurrent(lexer.TYPEID) {
			return nil
		}
		binding.Type = &ast.TypeIdentifier{Token: typeTok, Value: typeTok.Literal}
		if p.curTokenIs(lexer.ASSIGN) {
			p.nextToken()
			binding.Init = p.parseExpression(LOWEST)
		}
		exp.Bindings = append(exp.Bindings, binding)
		if !p.peekTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken()
		p.nextToken()
	}
	if !p.expectCurrent(lexer.IN) {
		return nil
	}
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
	exp := &ast.DynamicDispatch{Token: p.curToken, Object: left}
	p.nextToken() // skip '.'
	methodTok := p.curToken
	if methodTok.Type != lexer.OBJECTID {
		return nil
	}
	exp.Method = &ast.ObjectIdentifier{Token: methodTok, Value: methodTok.Literal}
	p.nextToken()
	if p.curTokenIs(lexer.LPAREN) {
		p.nextToken()
		if !p.peekTokenIs(lexer.RPAREN) {
			exp.Arguments = p.parseExpressionList(lexer.RPAREN)
		}
		if !p.expectCurrent(lexer.RPAREN) {
			return nil
		}
	}
	return exp
}

func (p *Parser) parseStaticDispatch(left ast.Expression) ast.Expression {
	exp := &ast.StaticDispatch{Token: p.curToken, Object: left}
	p.nextToken() // skip '@'
	typeTok := p.curToken
	if typeTok.Type != lexer.TYPEID {
		return nil
	}
	exp.Type = &ast.TypeIdentifier{Token: typeTok, Value: typeTok.Literal}
	p.nextToken()
	if !p.expectCurrent(lexer.DOT) {
		return nil
	}
	methodTok := p.curToken
	if methodTok.Type != lexer.OBJECTID {
		return nil
	}
	exp.Method = &ast.ObjectIdentifier{Token: methodTok, Value: methodTok.Literal}
	p.nextToken()
	if p.curTokenIs(lexer.LPAREN) {
		p.nextToken()
		if !p.peekTokenIs(lexer.RPAREN) {
			exp.Arguments = p.parseExpressionList(lexer.RPAREN)
		}
		if !p.expectCurrent(lexer.RPAREN) {
			return nil
		}
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
