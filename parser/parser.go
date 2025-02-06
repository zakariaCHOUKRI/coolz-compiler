package parser

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
	"strconv"
	"strings"
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
	CASE        // case expressions
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
	p.registerPrefix(lexer.LET, p.parseLetExpression)
	p.registerPrefix(lexer.LBRACE, p.parseBlockExpression)
	p.registerPrefix(lexer.CASE, p.parseCaseExpression)
	p.registerPrefix(lexer.SELF, p.parseSelfExpression)
	p.registerPrefix(lexer.RBRACE, p.parseRBrace) // Add RBRACE to prefix parsers

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

// Remove parseDebugCounter variable and debug print statements
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
	p.errors = append(p.errors, fmt.Sprintf("Expected next token to be %s, got %s line %d col %d",
		t.String(), p.peekToken.Type.String(), p.peekToken.Line, p.peekToken.Column))
}

func (p *Parser) currentError(t lexer.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("Expected current token to be %s, got %s line %d col %d",
		t.String(), p.curToken.Type.String(), p.curToken.Line, p.curToken.Column))
}

// Simplify ParseProgram by removing debug prints
func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}

	for p.curToken.Type != lexer.EOF && p.curToken.Type != lexer.ERROR {
		c := p.parseClass()
		if c == nil {
			break
		}
		prog.Classes = append(prog.Classes, c)
	}

	return prog
}

// Fix parseClass to properly handle semicolons
// Simplify parseClass by removing debug prints
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
		p.nextToken() // consume 'inherits'
		if !p.curTokenIs(lexer.TYPEID) {
			p.currentError(lexer.TYPEID)
			return nil
		}
		c.Parent = &ast.TypeIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		p.nextToken()
	}

	if !p.curTokenIs(lexer.LBRACE) {
		p.currentError(lexer.LBRACE)
		return nil
	}
	p.nextToken() // consume '{'

	// Parse features until we hit '}'
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.SEMI) {
			p.nextToken() // Skip standalone semicolons
			continue
		}

		feature := p.parseFeature()
		if feature != nil {
			c.Features = append(c.Features, feature)
			if !p.expectAndPeek(lexer.SEMI) {
				return nil
			}
			p.nextToken() // Move past semicolon
		}
	}

	if !p.curTokenIs(lexer.RBRACE) {
		p.currentError(lexer.RBRACE)
		return nil
	}
	p.nextToken() // consume '}'

	return c
}

func (p *Parser) parseFeature() ast.Feature {
	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}

	startToken := p.curToken
	p.nextToken() // consume the identifier

	if p.curTokenIs(lexer.LPAREN) {
		method := &ast.Method{
			Token: startToken,
			Name:  &ast.ObjectIdentifier{Token: startToken, Value: startToken.Literal},
		}

		if !p.parseMethodBody(method) {
			return nil
		}
		return method
	}

	attr := &ast.Attribute{
		Token: startToken,
		Name:  &ast.ObjectIdentifier{Token: startToken, Value: startToken.Literal},
	}

	if !p.expectCurrent(lexer.COLON) {
		return nil
	}

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}

	attr.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	if p.curTokenIs(lexer.ASSIGN) {
		p.nextToken() // move past '<-'
		attr.Init = p.parseExpression(LOWEST)
	}

	return attr
}

// Update parseMethodBody to handle tokens correctly
func (p *Parser) parseMethodBody(method *ast.Method) bool {
	// Already on LPAREN from parseFeature
	p.nextToken() // consume '('

	if !p.curTokenIs(lexer.RPAREN) {
		formals := []*ast.Formal{}

		// Parse first formal
		formal := p.parseFormal()
		if formal != nil {
			formals = append(formals, formal)

			// Parse additional formals
			for p.curTokenIs(lexer.COMMA) {
				p.nextToken() // consume comma
				formal = p.parseFormal()
				if formal == nil {
					return false
				}
				formals = append(formals, formal)
			}
		}

		method.Formals = formals

		if !p.curTokenIs(lexer.RPAREN) {
			p.currentError(lexer.RPAREN)
			return false
		}
	}

	p.nextToken() // consume ')'

	if !p.expectCurrent(lexer.COLON) {
		return false
	}

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return false
	}

	method.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	if !p.curTokenIs(lexer.LBRACE) {
		p.currentError(lexer.LBRACE)
		return false
	}
	p.nextToken() // consume '{'

	method.Body = p.parseExpression(LOWEST)
	if method.Body == nil {
		return false
	}

	if !p.peekTokenIs(lexer.RBRACE) {
		p.peekError(lexer.RBRACE)
		return false
	}
	p.nextToken() // consume expression
	p.nextToken() // consume '}'

	return true
}

// Update parseFormal to handle tokens correctly
func (p *Parser) parseFormal() *ast.Formal {
	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}

	formal := &ast.Formal{
		Name: &ast.ObjectIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
	}

	p.nextToken() // consume identifier

	if !p.curTokenIs(lexer.COLON) {
		p.currentError(lexer.COLON)
		return nil
	}
	p.nextToken() // consume colon

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}

	formal.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	p.nextToken() // consume type
	return formal
}

// Update parseFormals to correctly handle multiple formals
func (p *Parser) parseFormals() []*ast.Formal {
	var formals []*ast.Formal

	// Parse first formal
	formal := p.parseFormal()
	if formal == nil {
		return nil
	}
	formals = append(formals, formal)

	// Parse additional formals after commas
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		formal = p.parseFormal()
		if formal == nil {
			return nil
		}
		formals = append(formals, formal)
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
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()
	if leftExp == nil {
		return nil
	}

	// Add a safety counter to prevent infinite loops
	loopCount := 0
	maxLoops := 100

	for !p.curTokenIs(lexer.SEMI) && !p.curTokenIs(lexer.EOF) && precedence < p.peekPrecedence() {
		loopCount++
		if loopCount > maxLoops {
			return leftExp
		}

		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			break
		}

		p.nextToken()
		newExp := infix(leftExp)
		if newExp == nil {
			return leftExp // Return what we have so far instead of nil
		}
		leftExp = newExp
	}

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
	p.nextToken() // advance to avoid infinite loop
}

func (p *Parser) parseIdentifier() ast.Expression {
	if p.peekTokenIs(lexer.LPAREN) {
		return p.parseDispatchExpression(&ast.ObjectIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})
	}
	return &ast.ObjectIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseSelfExpression() ast.Expression {
	return &ast.SelfExpression{Token: p.curToken}
}

func (p *Parser) parseDispatchExpression(method *ast.ObjectIdentifier) ast.Expression {
	exp := &ast.DispatchExpression{
		Token:  p.curToken,
		Method: method,
	}

	if !p.expectAndPeek(lexer.LPAREN) {
		return nil
	}

	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	// consume closing parenthesis
	p.nextToken()
	return exp
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	num, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: num}
}

// Update parseStringLiteral to handle escaping
func (p *Parser) parseStringLiteral() ast.Expression {
	value := p.unescapeString(p.curToken.Literal)
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: value,
	}
}

func (p *Parser) unescapeString(s string) string {
	var result strings.Builder
	escaped := false

	// Remove surrounding quotes
	s = s[1 : len(s)-1]

	for _, ch := range s {
		if escaped {
			switch ch {
			case 'n':
				result.WriteRune('\n')
			case 't':
				result.WriteRune('\t')
			case '"':
				result.WriteRune('"')
			case '\\':
				result.WriteRune('\\')
			default:
				result.WriteRune(ch)
			}
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else {
			result.WriteRune(ch)
		}
	}

	return result.String()
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
	// If it's 'not', parse the entire next expression at LOWEST precedence
	if p.curTokenIs(lexer.NOT) {
		p.nextToken()
		exp.Right = p.parseExpression(LOWEST)
		return exp
	}
	// ...existing code...
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

// Fix parseDotExpression to properly handle method calls with arguments
func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	exp := &ast.MethodCallExpression{
		Token:  p.curToken,
		Object: left,
	}

	if !p.expectAndPeek(lexer.OBJECTID) {
		return nil
	}

	exp.Method = &ast.ObjectIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectAndPeek(lexer.LPAREN) {
		return nil
	}

	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	p.nextToken() // consume ')'

	return exp
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	var exps []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return exps
	}

	p.nextToken()
	exps = append(exps, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		exps = append(exps, p.parseExpression(LOWEST))
	}

	if !p.expectAndPeek(end) {
		return nil
	}

	return exps
}

func (p *Parser) parseLetExpression() ast.Expression {
	exp := &ast.LetExpression{Token: p.curToken}
	p.nextToken() // consume 'let'

	// Parse first binding
	binding := p.parseLetBinding()
	if binding == nil {
		return nil
	}
	exp.Bindings = append(exp.Bindings, binding)

	// Parse additional bindings
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next identifier
		binding = p.parseLetBinding()
		if binding == nil {
			return nil
		}
		exp.Bindings = append(exp.Bindings, binding)
	}

	if !p.expectAndPeek(lexer.IN) {
		p.peekError(lexer.IN)
		return nil
	}

	p.nextToken() // move past 'in'
	exp.In = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseLetBinding() *ast.LetBinding {
	binding := &ast.LetBinding{}

	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}

	binding.Identifier = &ast.ObjectIdentifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectAndPeek(lexer.COLON) {
		return nil
	}

	if !p.expectAndPeek(lexer.TYPEID) {
		return nil
	}

	binding.Type = &ast.TypeIdentifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // move to '<-'
		p.nextToken() // move past '<-'
		binding.Init = p.parseExpression(LOWEST)
	}

	return binding
}

func (p *Parser) parseBlockExpression() ast.Expression {
	block := &ast.BlockExpression{Token: p.curToken}
	p.nextToken() // move past '{'

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		exp := p.parseExpression(LOWEST)
		if exp != nil {
			block.Expressions = append(block.Expressions, exp)
		}

		if !p.expectAndPeek(lexer.SEMI) {
			return nil
		}
		p.nextToken() // move past semicolon
	}

	if !p.curTokenIs(lexer.RBRACE) {
		p.currentError(lexer.RBRACE)
		return nil
	}

	p.nextToken() // move past '}'
	return block
}

func (p *Parser) parseCaseExpression() ast.Expression {
	exp := &ast.CaseExpression{Token: p.curToken}
	p.nextToken()

	exp.Subject = p.parseExpression(LOWEST)
	if exp.Subject == nil {
		return nil
	}

	if !p.expectAndPeek(lexer.OF) {
		return nil
	}
	p.nextToken()

	for !p.curTokenIs(lexer.ESAC) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.SEMI) {
			p.nextToken()
			continue
		}

		branch := p.parseCaseBranch()
		if branch == nil {
			return nil
		}
		exp.Branches = append(exp.Branches, branch)

		if !p.curTokenIs(lexer.ESAC) && !p.peekTokenIs(lexer.ESAC) {
			if !p.peekTokenIs(lexer.SEMI) {
				return nil
			}
			p.nextToken()
			p.nextToken()
		}
	}

	if !p.curTokenIs(lexer.ESAC) {
		p.currentError(lexer.ESAC)
		return nil
	}
	p.nextToken()

	return exp
}

func (p *Parser) parseCaseBranch() *ast.CaseBranch {
	if !p.curTokenIs(lexer.OBJECTID) {
		p.currentError(lexer.OBJECTID)
		return nil
	}

	branch := &ast.CaseBranch{
		Variable: &ast.ObjectIdentifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
	}
	p.nextToken()

	if !p.curTokenIs(lexer.COLON) {
		p.currentError(lexer.COLON)
		return nil
	}
	p.nextToken()

	if !p.curTokenIs(lexer.TYPEID) {
		p.currentError(lexer.TYPEID)
		return nil
	}

	branch.Type = &ast.TypeIdentifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	p.nextToken()

	if !p.curTokenIs(lexer.DARROW) {
		p.currentError(lexer.DARROW)
		return nil
	}
	p.nextToken()

	branch.Expression = p.parseExpression(LOWEST)
	if branch.Expression == nil {
		return nil
	}

	return branch
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

// Add parseRBrace method
func (p *Parser) parseRBrace() ast.Expression {
	// This is a no-op parser just to prevent the "no prefix parse function" error
	return nil
}

// Update the precedences map to include TIMES
var precedences = map[lexer.TokenType]int{
	lexer.ASSIGN: ASSIGN,
	lexer.EQ:     EQUALS,
	lexer.LT:     LESSGREATER,
	lexer.LE:     LESSGREATER,
	lexer.PLUS:   SUM,
	lexer.MINUS:  SUM,
	lexer.DIVIDE: PRODUCT,
	lexer.TIMES:  PRODUCT, // Add TIMES operator
	lexer.DOT:    DOT,     // Add DOT operator
	lexer.NOT:    LOWEST,  // Lower "not" so it affects the entire expression
	lexer.RBRACE: LOWEST,  // Add RBRACE operator
}
