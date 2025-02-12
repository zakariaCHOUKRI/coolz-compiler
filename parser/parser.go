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
	SUM         // + or -
	PRODUCT     // * or /
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
	p.registerPrefix(lexer.INT_CONST, p.parseIntegerLiteral)
	p.registerPrefix(lexer.OBJECTID, p.parseIdentifier)
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
	p.registerPrefix(lexer.LBRACE, p.parseBlockExpression)

	// Register infix parsers
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.DIVIDE, p.parseInfixExpression)
	p.registerInfix(lexer.TIMES, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.LE, p.parseInfixExpression)
	p.registerInfix(lexer.ASSIGN, p.parseAssignment)
	p.registerInfix(lexer.DOT, p.parseDynamicDispatch) // Register DOT for dynamic dispatch
	p.registerInfix(lexer.AT, p.parseStaticDispatch)   // Register AT for static dispatch

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

func (p *Parser) expectPeek(t lexer.TokenType) bool {
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
	p.debugToken(fmt.Sprintf("parseExpression with precedence %d", precedence))
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.errors = append(p.errors, fmt.Sprintf("No prefix parse function for %v (literal: %s) found at line %d, column %d",
			p.curToken.Type, p.curToken.Literal, p.curToken.Line, p.curToken.Column))
		return nil
	}
	leftExp := prefix()
	if leftExp == nil {
		p.debugToken(fmt.Sprintf("Prefix parse function for %v returned nil", p.curToken.Type))
		return nil
	}
	for {
		if p.curTokenIs(lexer.SEMI) || p.curTokenIs(lexer.ESAC) || p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.DARROW) {
			break
		}
		infix := p.infixParseFns[p.curToken.Type]
		if infix == nil {
			break
		}
		currPrecedence := p.curPrecedence()
		if currPrecedence <= precedence {
			break
		}
		p.debugToken(fmt.Sprintf("Before infix: left=%T, operator=%v", leftExp, p.curToken.Type))
		leftExp = infix(leftExp)
		if leftExp == nil {
			return nil
		}
	}
	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	var exp ast.Expression
	if p.curToken.Literal == "self" {
		exp = &ast.Self{Token: p.curToken}
	} else {
		exp = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	p.nextToken() // Advance past the identifier

	// Check if this is a method call
	if p.curTokenIs(lexer.LPAREN) {
		dispatch := &ast.DynamicDispatch{
			Token:  p.curToken,
			Object: &ast.Self{Token: lexer.Token{Type: lexer.SELF, Literal: "self"}}, // Implicit self
			Method: &ast.ObjectIdentifier{Token: exp.(*ast.ObjectIdentifier).Token, Value: exp.(*ast.ObjectIdentifier).Value},
		}

		p.nextToken() // consume the '('

		// Parse arguments
		var args []ast.Expression

		if !p.curTokenIs(lexer.RPAREN) {
			// Parse first argument
			firstArg := p.parseExpression(LOWEST)
			if firstArg != nil {
				args = append(args, firstArg)
			}

			// Parse remaining arguments
			for p.curTokenIs(lexer.COMMA) {
				p.nextToken() // consume comma
				arg := p.parseExpression(LOWEST)
				if arg != nil {
					args = append(args, arg)
				}
			}
		}

		if !p.expectCurrent(lexer.RPAREN) {
			return nil
		}

		dispatch.Arguments = args
		return dispatch
	}

	return exp
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	p.debugToken("Parsing integer literal")
	num, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("Could not parse %q as integer", p.curToken.Literal))
		return nil
	}
	lit := &ast.IntegerLiteral{Token: p.curToken, Value: num}
	p.nextToken()
	p.debugToken("After parsing integer literal")
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
	p.debugToken("Starting case expression")

	exp := &ast.CaseExpression{Token: p.curToken}
	p.nextToken() // Skip 'case'

	p.debugToken("Before parsing case condition")
	exp.Expr = p.parseExpression(LOWEST)
	if exp.Expr == nil {
		p.errors = append(p.errors, "Failed to parse case condition")
		return nil
	}
	p.debugToken(fmt.Sprintf("After parsing case condition: %T", exp.Expr))

	if !p.expectCurrent(lexer.OF) {
		p.errors = append(p.errors, fmt.Sprintf("Expected 'of' after case expression, got %s", p.curToken.Type))
		return nil
	}

	for !p.curTokenIs(lexer.ESAC) && !p.curTokenIs(lexer.EOF) {
		p.debugToken("Starting case branch")

		branch := &ast.CaseBranch{}

		if p.curToken.Type != lexer.OBJECTID {
			p.errors = append(p.errors, fmt.Sprintf("Expected identifier in case branch, got %s", p.curToken.Type))
			return nil
		}

		branch.Identifier = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()

		if !p.expectCurrent(lexer.COLON) {
			return nil
		}

		if p.curToken.Type != lexer.TYPEID {
			p.errors = append(p.errors, fmt.Sprintf("Expected type in case branch, got %s", p.curToken.Type))
			return nil
		}

		branch.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()

		if !p.expectCurrent(lexer.DARROW) {
			return nil
		}

		p.debugToken("Before parsing branch expression")
		// Use LOWEST precedence to ensure we parse the entire expression
		branchExpr := p.parseExpression(LOWEST)
		p.debugToken(fmt.Sprintf("After parsing branch expression: %T", branchExpr))

		if branchExpr == nil {
			p.errors = append(p.errors, "Failed to parse case branch expression")
			return nil
		}
		branch.Expr = branchExpr

		exp.Branches = append(exp.Branches, branch)

		// Only move to next token if it's a semicolon
		if p.curTokenIs(lexer.SEMI) {
			p.nextToken()
			p.debugToken("After semicolon")
		} else if !p.curTokenIs(lexer.ESAC) {
			p.errors = append(p.errors, fmt.Sprintf("Expected semicolon or esac after branch, got %s", p.curToken.Type))
			return nil
		}
	}

	if !p.expectCurrent(lexer.ESAC) {
		p.errors = append(p.errors, fmt.Sprintf("Expected 'esac' at end of case expression, got %s", p.curToken.Type))
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
	p.debugToken(fmt.Sprintf("Parsing infix expression with operator %s, left=%v",
		p.curToken.Literal, left.TokenLiteral()))

	exp := &ast.BinaryExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken() // Move past the operator

	if p.curToken.Type == lexer.EOF {
		p.errors = append(p.errors, "Unexpected EOF in infix expression")
		return nil
	}

	p.debugToken(fmt.Sprintf("Parsing right side of %s with precedence %d", exp.Operator, precedence))
	exp.Right = p.parseExpression(precedence)
	if exp.Right == nil {
		p.errors = append(p.errors, fmt.Sprintf("Failed to parse right side of %s expression", exp.Operator))
		return nil
	}

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
	lexer.DOT:    CALL,
	lexer.AT:     CALL,
	lexer.ASSIGN: LOWEST,
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

// Add this debugging function
func (p *Parser) debugToken(prefix string) {
	fmt.Printf("DEBUG: %s - Current Token: {Type: %v, Literal: %s, Line: %d, Column: %d}\n",
		prefix, p.curToken.Type, p.curToken.Literal, p.curToken.Line, p.curToken.Column)
}

func (p *Parser) parseSelf() ast.Expression {
	selfNode := &ast.Self{Token: p.curToken}
	p.nextToken() // Advance to the next token after 'self'
	return selfNode
}

func (p *Parser) parseVoidLiteral() ast.Expression {
	return &ast.VoidLiteral{Token: p.curToken}
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	var args []ast.Expression

	// If the list is empty (e.g., `foo()`), consume the end token and return
	if p.peekTokenIs(end) {
		p.nextToken() // Consume the end token
		return args
	}

	// Parse the first argument
	p.nextToken() // Move to the first argument
	args = append(args, p.parseExpression(LOWEST))

	// Parse additional arguments separated by commas
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken() // Consume the comma
		args = append(args, p.parseExpression(LOWEST))
	}

	// Expect the end token (e.g., `)`)
	if !p.expectPeek(end) {
		return nil
	}

	return args
}

func (p *Parser) parseBlockExpression() ast.Expression {
	block := &ast.BlockExpression{Token: p.curToken}
	p.nextToken() // Consume '{'

	var exprs []ast.Expression
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			break
		}
		exprs = append(exprs, expr)

		// Require semicolon after each expression except the last
		if !p.peekTokenIs(lexer.RBRACE) && !p.expectPeek(lexer.SEMI) {
			return nil
		}
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	block.Expressions = exprs
	return block
}

func (p *Parser) parseDynamicDispatch(left ast.Expression) ast.Expression {
	dd := &ast.DynamicDispatch{
		Token:  p.curToken,
		Object: left,
	}
	p.nextToken() // Consume the DOT

	// Parse method name (must be OBJECTID)
	if !p.curTokenIs(lexer.OBJECTID) {
		p.errors = append(p.errors, fmt.Sprintf("expected method name after '.', got %s", p.curToken.Type))
		return nil
	}
	dd.Method = &ast.ObjectIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	p.nextToken() // Consume the method name

	// Parse arguments inside parentheses
	if !p.expectCurrent(lexer.LPAREN) {
		return nil
	}

	// Parse argument list
	var args []ast.Expression

	// Handle empty argument list
	if p.curTokenIs(lexer.RPAREN) {
		p.nextToken() // consume )
		dd.Arguments = args
		return dd
	}

	// Parse first argument
	exp := p.parseExpression(LOWEST)
	if exp != nil {
		args = append(args, exp)
	}

	// Parse additional arguments
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		exp = p.parseExpression(LOWEST)
		if exp != nil {
			args = append(args, exp)
		}
	}

	if !p.expectCurrent(lexer.RPAREN) {
		return nil
	}

	dd.Arguments = args
	return dd
}

func (p *Parser) parseStaticDispatch(left ast.Expression) ast.Expression {
	sd := &ast.StaticDispatch{
		Token:  p.curToken,
		Object: left,
	}

	// Consume the @ token
	p.nextToken() // Changed from expectCurrent to nextToken

	// Parse the type identifier
	if !p.curTokenIs(lexer.TYPEID) {
		p.errors = append(p.errors, fmt.Sprintf("expected type identifier after '@', got %s", p.curToken.Type))
		return nil
	}
	sd.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	p.nextToken()

	// Expect and consume DOT after type
	if !p.curTokenIs(lexer.DOT) {
		p.errors = append(p.errors, fmt.Sprintf("expected '.' after type, got %s", p.curToken.Type))
		return nil
	}
	p.nextToken()

	// Parse method name (must be OBJECTID)
	if !p.curTokenIs(lexer.OBJECTID) {
		p.errors = append(p.errors, fmt.Sprintf("expected method name after '.', got %s", p.curToken.Type))
		return nil
	}
	sd.Method = &ast.ObjectIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	p.nextToken()

	// Parse arguments inside parentheses
	if !p.curTokenIs(lexer.LPAREN) {
		p.errors = append(p.errors, fmt.Sprintf("expected '(' after method name, got %s", p.curToken.Type))
		return nil
	}

	// Parse the argument list
	var args []ast.Expression

	p.nextToken() // consume (

	// Handle empty argument list
	if p.curTokenIs(lexer.RPAREN) {
		p.nextToken() // consume )
		sd.Arguments = args
		return sd
	}

	// Parse first argument
	exp := p.parseExpression(LOWEST)
	if exp != nil {
		args = append(args, exp)
	}

	// Parse additional arguments
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		exp = p.parseExpression(LOWEST)
		if exp != nil {
			args = append(args, exp)
		}
	}

	if !p.expectCurrent(lexer.RPAREN) {
		return nil
	}

	sd.Arguments = args
	return sd
}
