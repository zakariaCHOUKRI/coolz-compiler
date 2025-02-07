package codegen

import (
	"coolz-compiler/codegen/testutil"
	"strings"
	"testing"
)

func TestGenerateIntConstant(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected string
	}{
		{
			name:     "positive integer",
			value:    42,
			expected: "li $a0, 42",
		},
		{
			name:     "negative integer",
			value:    -10,
			expected: "li $a0, -10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := NewCodeGenerator()
			expr := &IntConstant{Value: tt.value}
			cg.generateIntConstant(expr)
			testutil.AssemblyMatcher(t, cg.assembly.String(), tt.expected)
		})
	}
}

func TestGenerateIf(t *testing.T) {
	cg := NewCodeGenerator()
	ifExpr := &If{
		Condition:  &BoolConstant{Value: true},
		ThenBranch: &IntConstant{Value: 1},
		ElseBranch: &IntConstant{Value: 0},
	}

	cg.generateIf(ifExpr)

	expected := `
		li $a0, 1
		beqz $a0, L1
		li $a0, 1
		j L2
	L1:
		li $a0, 0
	L2:
	`

	testutil.AssemblyMatcher(t, cg.assembly.String(), expected)
}

func TestGenerateWhile(t *testing.T) {
	cg := NewCodeGenerator()
	whileExpr := &While{
		Condition: &BoolConstant{Value: true},
		Body: &Block{
			Expressions: []Expression{
				&IntConstant{Value: 1},
				&IntConstant{Value: 2},
			},
		},
	}

	cg.generateWhile(whileExpr)

	expected := `
        L1:
        li $a0, 1
        beqz $a0, L2
        li $a0, 1
        li $a0, 2
        j L1
        L2:
    `

	testutil.AssemblyMatcher(t, cg.assembly.String(), expected)
}

func TestGenerateStringConstant(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{
			name:  "simple string",
			value: "Hello",
			want:  []string{"str1: .asciiz \"Hello\"", "\tla $a0, str1"},
		},
		{
			name:  "string with escapes",
			value: "Hello\nWorld",
			want:  []string{"str1: .asciiz \"Hello\\nWorld\"", "\tla $a0, str1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := NewCodeGenerator()
			expr := &StringConstant{Value: tt.value}
			result := cg.Generate(expr)

			for _, expected := range tt.want {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected assembly to contain %q, but didn't find it", expected)
				}
			}
		})
	}
}

func TestGenerateBlock(t *testing.T) {
	cg := NewCodeGenerator()
	block := &Block{
		Expressions: []Expression{
			&IntConstant{Value: 1},
			&StringConstant{Value: "test"},
			&BoolConstant{Value: true},
		},
	}

	// Use Generate instead of generateBlock directly
	result := cg.Generate(block)

	expected := []string{
		".data",
		"str1: .asciiz \"test\"",
		".text",
		"li $a0, 1",
		"la $a0, str1",
		"li $a0, 1",
	}

	// Check each expected line exists in the result
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected assembly to contain %q, but didn't find it", exp)
		}
	}
}
