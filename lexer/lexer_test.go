package lexer

import (
	"strings"
	"testing"
)

func TestNextToken(t *testing.T) {
	tests := []struct {
		input             string
		expectedTokenType []TokenType
		expectedLiteral   []string
	}{
		{
			"class Main {};",
			[]TokenType{CLASS, TYPEID, LBRACE, RBRACE, SEMI, EOF},
			[]string{"class", "Main", "{", "}", ";", ""},
		},
		{
			"x <- true;-- One line comment\nx <- false;",
			[]TokenType{OBJECTID, ASSIGN, BOOL_CONST, SEMI, OBJECTID, ASSIGN, BOOL_CONST, SEMI, EOF},
			[]string{"x", "<-", "true", ";", "x", "<-", "false", ";", ""},
		},
		{
			"_a <- 0; b   <- _a <= \"1\\n\";",
			[]TokenType{OBJECTID, ASSIGN, INT_CONST, SEMI, OBJECTID, ASSIGN, OBJECTID, LE, STR_CONST, SEMI, EOF},
			[]string{"_a", "<-", "0", ";", "b", "<-", "_a", "<=", "1\n", ";", ""},
		},
		{
			"{true\n1\n\"some string\"\n}",
			[]TokenType{LBRACE, BOOL_CONST, INT_CONST, STR_CONST, RBRACE, EOF},
			[]string{"{", "true", "1", "some string", "}", ""},
		},
		{
			"{true\n1\n\"some string\"}",
			[]TokenType{LBRACE, BOOL_CONST, INT_CONST, STR_CONST, RBRACE, EOF},
			[]string{"{", "true", "1", "some string", "}", ""},
		},
		{
			"let a:A in true",
			[]TokenType{LET, OBJECTID, COLON, TYPEID, IN, BOOL_CONST, EOF},
			[]string{"let", "a", ":", "A", "in", "true", ""},
		},
		{
			"case a of b:B => false esac",
			[]TokenType{CASE, OBJECTID, OF, OBJECTID, COLON, TYPEID, DARROW, BOOL_CONST, ESAC, EOF},
			[]string{"case", "a", "of", "b", ":", "B", "=>", "false", "esac", ""},
		},
		{
			"(* This is a multi-line comment *) class Main {};",
			[]TokenType{CLASS, TYPEID, LBRACE, RBRACE, SEMI, EOF},
			[]string{"class", "Main", "{", "}", ";", ""},
		},
		{
			"(* Nested (* multi-line *) comment *) class Main {};",
			[]TokenType{CLASS, TYPEID, LBRACE, RBRACE, SEMI, EOF},
			[]string{"class", "Main", "{", "}", ";", ""},
		},
		{
			`class List {
				(* Define operations on empty lists. *)
				isNil() : Bool { true };
				head()  : Int { { abort(); 0; } };
				tail()  : List { { abort(); self; } };
				cons(i : Int) : List { (new Cons).init(i, self) };
			};`,
			[]TokenType{CLASS, TYPEID, LBRACE, OBJECTID, LPAREN, RPAREN, COLON, TYPEID, LBRACE, BOOL_CONST, RBRACE, SEMI, OBJECTID, LPAREN, RPAREN, COLON, TYPEID, LBRACE, LBRACE, OBJECTID, LPAREN, RPAREN, SEMI, INT_CONST, SEMI, RBRACE, RBRACE, SEMI, OBJECTID, LPAREN, RPAREN, COLON, TYPEID, LBRACE, LBRACE, OBJECTID, LPAREN, RPAREN, SEMI, OBJECTID, SEMI, RBRACE, RBRACE, SEMI, OBJECTID, LPAREN, OBJECTID, COLON, TYPEID, RPAREN, COLON, TYPEID, LBRACE, LPAREN, NEW, TYPEID, RPAREN, DOT, OBJECTID, LPAREN, OBJECTID, COMMA, OBJECTID, RPAREN, RBRACE, SEMI, RBRACE, SEMI, EOF},
			[]string{"class", "List", "{", "isNil", "(", ")", ":", "Bool", "{", "true", "}", ";", "head", "(", ")", ":", "Int", "{", "{", "abort", "(", ")", ";", "0", ";", "}", "}", ";", "tail", "(", ")", ":", "List", "{", "{", "abort", "(", ")", ";", "self", ";", "}", "}", ";", "cons", "(", "i", ":", "Int", ")", ":", "List", "{", "(", "new", "Cons", ")", ".", "init", "(", "i", ",", "self", ")", "}", ";", "}", ";", ""},
		},
		{
			`class Main inherits IO {
				mylist : List;
				print_list(l : List) : Object {
					if l.isNil() then out_string("\n")
					else {
						out_int(l.head());
						out_string(" ");
						print_list(l.tail());
					} fi
				};
				main() : Object {
					{
						mylist <- new List.cons(1).cons(2).cons(3).cons(4).cons(5);
						while (not mylist.isNil()) loop {
							print_list(mylist);
							mylist <- mylist.tail();
						} pool;
					}
				};
			};`,
			[]TokenType{CLASS, TYPEID, INHERITS, TYPEID, LBRACE, OBJECTID, COLON, TYPEID, SEMI, OBJECTID, LPAREN, OBJECTID, COLON, TYPEID, RPAREN, COLON, TYPEID, LBRACE, IF, OBJECTID, DOT, OBJECTID, LPAREN, RPAREN, THEN, OBJECTID, LPAREN, STR_CONST, RPAREN, ELSE, LBRACE, OBJECTID, LPAREN, OBJECTID, DOT, OBJECTID, LPAREN, RPAREN, RPAREN, SEMI, OBJECTID, LPAREN, STR_CONST, RPAREN, SEMI, OBJECTID, LPAREN, OBJECTID, DOT, OBJECTID, LPAREN, RPAREN, RPAREN, SEMI, RBRACE, FI, RBRACE, SEMI, OBJECTID, LPAREN, RPAREN, COLON, TYPEID, LBRACE, LBRACE, OBJECTID, ASSIGN, NEW, TYPEID, DOT, OBJECTID, LPAREN, INT_CONST, RPAREN, DOT, OBJECTID, LPAREN, INT_CONST, RPAREN, DOT, OBJECTID, LPAREN, INT_CONST, RPAREN, DOT, OBJECTID, LPAREN, INT_CONST, RPAREN, DOT, OBJECTID, LPAREN, INT_CONST, RPAREN, SEMI, WHILE, LPAREN, NOT, OBJECTID, DOT, OBJECTID, LPAREN, RPAREN, RPAREN, LOOP, LBRACE, OBJECTID, LPAREN, OBJECTID, RPAREN, SEMI, OBJECTID, ASSIGN, OBJECTID, DOT, OBJECTID, LPAREN, RPAREN, SEMI, RBRACE, POOL, SEMI, RBRACE, RBRACE, SEMI, RBRACE, SEMI, EOF},
			[]string{"class", "Main", "inherits", "IO", "{", "mylist", ":", "List", ";", "print_list", "(", "l", ":", "List", ")", ":", "Object", "{", "if", "l", ".", "isNil", "(", ")", "then", "out_string", "(", "\n", ")", "else", "{", "out_int", "(", "l", ".", "head", "(", ")", ")", ";", "out_string", "(", " ", ")", ";", "print_list", "(", "l", ".", "tail", "(", ")", ")", ";", "}", "fi", "}", ";", "main", "(", ")", ":", "Object", "{", "{", "mylist", "<-", "new", "List", ".", "cons", "(", "1", ")", ".", "cons", "(", "2", ")", ".", "cons", "(", "3", ")", ".", "cons", "(", "4", ")", ".", "cons", "(", "5", ")", ";", "while", "(", "not", "mylist", ".", "isNil", "(", ")", ")", "loop", "{", "print_list", "(", "mylist", ")", ";", "mylist", "<-", "mylist", ".", "tail", "(", ")", ";", "}", "pool", ";", "}", "}", ";", "}", ";", ""},
		},
	}

	for _, tt := range tests {
		l := NewLexer(strings.NewReader(tt.input))

		// Add debug information
		var tokens []Token
		for i, expTType := range tt.expectedTokenType {
			tok := l.NextToken()
			tokens = append(tokens, tok)

			if tok.Type != expTType {
				// Print the token sequence up to the error
				t.Logf("\nToken sequence up to error:")
				for j, token := range tokens {
					t.Logf("%d: Type=%v, Literal='%s', Expected Type=%v",
						j, token.Type, token.Literal, tt.expectedTokenType[j])
				}

				t.Fatalf("\n[%q]:\nError at token %d\nGot: Type=%v, Literal='%s'\nExpected: Type=%v, Literal='%s'",
					tt.input, i, tok.Type, tok.Literal, expTType, tt.expectedLiteral[i])
			}

			if tok.Literal != tt.expectedLiteral[i] {
				t.Fatalf("[%q]: Wrong literal at token %d. expected=%q, got=%q",
					tt.input, i, tt.expectedLiteral[i], tok.Literal)
			}
		}
	}
}
