package ast

import (
	"coolz-compiler/lexer"
	"testing"
)

func TestSimpleClass(t *testing.T) {
	// Test the simple Main class from example.cl
	mainClass := &Class{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &TypeIdentifier{Value: "Main"},
		Features: []Feature{
			&Method{
				Name: &ObjectIdentifier{Value: "main"},
				Type: &TypeIdentifier{Value: "Int"},
				Body: &IntegerLiteral{Value: 42},
			},
		},
	}

	if mainClass.Name.Value != "Main" {
		t.Errorf("Expected class name 'Main', got %s", mainClass.Name.Value)
	}

	if len(mainClass.Features) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(mainClass.Features))
	}

	method, ok := mainClass.Features[0].(*Method)
	if !ok {
		t.Fatalf("Expected first feature to be a method")
	}

	if method.Name.Value != "main" {
		t.Errorf("Expected method name 'main', got %s", method.Name.Value)
	}

	if method.Type.Value != "Int" {
		t.Errorf("Expected return type 'Int', got %s", method.Type.Value)
	}
}

func TestListImplementation(t *testing.T) {
	// Test the List class from example_complex.cl
	listClass := &Class{
		Token: lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:  &TypeIdentifier{Value: "List"},
		Features: []Feature{
			&Method{
				Name: &ObjectIdentifier{Value: "isNil"},
				Type: &TypeIdentifier{Value: "Bool"},
				Body: &BooleanLiteral{Value: true},
			},
			&Method{
				Name: &ObjectIdentifier{Value: "head"},
				Type: &TypeIdentifier{Value: "Int"},
				Body: &BlockExpression{
					Expressions: []Expression{
						&DynamicDispatch{
							Object: &Self{},
							Method: &ObjectIdentifier{Value: "abort"},
						},
						&IntegerLiteral{Value: 0},
					},
				},
			},
			&Method{
				Name: &ObjectIdentifier{Value: "cons"},
				Type: &TypeIdentifier{Value: "List"},
				Formals: []*Formal{
					{
						Name: &ObjectIdentifier{Value: "i"},
						Type: &TypeIdentifier{Value: "Int"},
					},
				},
				Body: &DynamicDispatch{
					Object: &NewExpression{
						Type: &TypeIdentifier{Value: "Cons"},
					},
					Method: &ObjectIdentifier{Value: "init"},
					Arguments: []Expression{
						&ObjectIdentifier{Value: "i"},
						&Self{},
					},
				},
			},
		},
	}

	if listClass.Name.Value != "List" {
		t.Errorf("Expected class name 'List', got %s", listClass.Name.Value)
	}

	if len(listClass.Features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(listClass.Features))
	}

	// Test the cons method
	consMethod, ok := listClass.Features[2].(*Method)
	if !ok {
		t.Fatalf("Expected third feature to be a method")
	}

	if len(consMethod.Formals) != 1 {
		t.Errorf("Expected 1 formal parameter, got %d", len(consMethod.Formals))
	}

	// Test the Cons class
	consClass := &Class{
		Token:  lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:   &TypeIdentifier{Value: "Cons"},
		Parent: &TypeIdentifier{Value: "List"},
		Features: []Feature{
			&Attribute{
				Name: &ObjectIdentifier{Value: "car"},
				Type: &TypeIdentifier{Value: "Int"},
			},
			&Attribute{
				Name: &ObjectIdentifier{Value: "cdr"},
				Type: &TypeIdentifier{Value: "List"},
			},
			&Method{
				Name: &ObjectIdentifier{Value: "init"},
				Type: &TypeIdentifier{Value: "List"},
				Formals: []*Formal{
					{
						Name: &ObjectIdentifier{Value: "i"},
						Type: &TypeIdentifier{Value: "Int"},
					},
					{
						Name: &ObjectIdentifier{Value: "rest"},
						Type: &TypeIdentifier{Value: "List"},
					},
				},
				Body: &BlockExpression{
					Expressions: []Expression{
						&Assignment{
							Left:  &ObjectIdentifier{Value: "car"},
							Value: &ObjectIdentifier{Value: "i"},
						},
						&Assignment{
							Left:  &ObjectIdentifier{Value: "cdr"},
							Value: &ObjectIdentifier{Value: "rest"},
						},
						&Self{},
					},
				},
			},
		},
	}

	if consClass.Parent.Value != "List" {
		t.Errorf("Expected parent class 'List', got %s", consClass.Parent.Value)
	}

	if len(consClass.Features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(consClass.Features))
	}
}

func TestIOExample(t *testing.T) {
	// Test the IO example from example_io.cl
	mainClass := &Class{
		Token:  lexer.Token{Type: lexer.CLASS, Literal: "class"},
		Name:   &TypeIdentifier{Value: "Main"},
		Parent: &TypeIdentifier{Value: "IO"},
		Features: []Feature{
			&Method{
				Name: &ObjectIdentifier{Value: "main"},
				Type: &TypeIdentifier{Value: "Object"},
				Body: &BlockExpression{
					Expressions: []Expression{
						&DynamicDispatch{
							Object: &Self{},
							Method: &ObjectIdentifier{Value: "out_string"},
							Arguments: []Expression{
								&StringLiteral{Value: "Hello. Please enter a number: "},
							},
						},
						&LetExpression{
							Bindings: []*LetBinding{
								{
									Identifier: &ObjectIdentifier{Value: "num"},
									Type:       &TypeIdentifier{Value: "Int"},
									Init: &DynamicDispatch{
										Object: &Self{},
										Method: &ObjectIdentifier{Value: "in_int"},
									},
								},
							},
							In: &BlockExpression{
								Expressions: []Expression{
									&DynamicDispatch{
										Object: &DynamicDispatch{
											Object: &DynamicDispatch{
												Object: &Self{},
												Method: &ObjectIdentifier{Value: "out_string"},
												Arguments: []Expression{
													&StringLiteral{Value: "You entered: "},
												},
											},
											Method: &ObjectIdentifier{Value: "out_int"},
											Arguments: []Expression{
												&ObjectIdentifier{Value: "num"},
											},
										},
										Method: &ObjectIdentifier{Value: "out_string"},
										Arguments: []Expression{
											&StringLiteral{Value: "\n"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if mainClass.Parent.Value != "IO" {
		t.Errorf("Expected parent class 'IO', got %s", mainClass.Parent.Value)
	}

	mainMethod, ok := mainClass.Features[0].(*Method)
	if !ok {
		t.Fatalf("Expected first feature to be a method")
	}

	block, ok := mainMethod.Body.(*BlockExpression)
	if !ok {
		t.Fatalf("Expected method body to be a block expression")
	}

	if len(block.Expressions) != 2 {
		t.Errorf("Expected 2 expressions in main method, got %d", len(block.Expressions))
	}
}

func TestTypeIdentifiers(t *testing.T) {
	tests := []struct {
		input    *TypeIdentifier
		expected string
	}{
		{&TypeIdentifier{Value: "Int"}, "Int"},
		{&TypeIdentifier{Value: "String"}, "String"},
		{&TypeIdentifier{Value: "Bool"}, "Bool"},
		{&TypeIdentifier{Value: "SELF_TYPE"}, "SELF_TYPE"},
	}

	for _, tt := range tests {
		if tt.input.Value != tt.expected {
			t.Errorf("Expected type %s, got %s", tt.expected, tt.input.Value)
		}
	}
}

func TestObjectIdentifiers(t *testing.T) {
	tests := []struct {
		input    *ObjectIdentifier
		expected string
	}{
		{&ObjectIdentifier{Value: "x"}, "x"},
		{&ObjectIdentifier{Value: "self"}, "self"},
		{&ObjectIdentifier{Value: "main"}, "main"},
	}

	for _, tt := range tests {
		if tt.input.Value != tt.expected {
			t.Errorf("Expected identifier %s, got %s", tt.expected, tt.input.Value)
		}
	}
}
