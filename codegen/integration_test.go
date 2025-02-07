package codegen

import (
	"coolz-compiler/codegen/testutil"
	"testing"
)

func TestSimpleProgram(t *testing.T) {
	// Create a simple program AST
	program := &Program{
		Classes: []*Class{
			{
				Name:   "Main",
				Parent: "Object",
				Features: []Feature{
					&Method{
						Name: "main",
						Body: &Block{
							Expressions: []Expression{
								&IntConstant{Value: 42},
							},
						},
					},
				},
			},
		},
	}

	// Generate code
	cg := NewCodeGenerator()
	result := cg.Generate(program)

	// Expected assembly (simplified)
	expected := `
		.data
		_vtable_Main:
		.text
		.globl main
		main:
			sw $fp, 0($sp)
			sw $ra, -4($sp)
			sw $s0, -8($sp)
			move $fp, $sp
			addiu $sp, $sp, -12
			li $a0, 42
			lw $fp, 0($sp)
			lw $ra, -4($sp)
			lw $s0, -8($sp)
			addiu $sp, $sp, 12
			jr $ra
	`

	testutil.AssemblyMatcher(t, result, expected)
}

func TestComplexProgram(t *testing.T) {
	program := &Program{
		Classes: []*Class{
			{
				Name:   "Main",
				Parent: "Object",
				Features: []Feature{
					&Method{
						Name: "main",
						Body: &Block{
							Expressions: []Expression{
								&If{
									Condition: &BoolConstant{Value: true},
									ThenBranch: &Block{
										Expressions: []Expression{
											&StringConstant{Value: "True branch"},
											&IntConstant{Value: 1},
										},
									},
									ElseBranch: &IntConstant{Value: 0},
								},
								&While{
									Condition: &BoolConstant{Value: true},
									Body: &Block{
										Expressions: []Expression{
											&IntConstant{Value: 42},
											&Dispatch{
												Object:     nil, // self
												MethodName: "print",
												Arguments: []Expression{
													&StringConstant{Value: "Loop"},
												},
											},
										},
									},
								},
							},
						},
					},
					&Method{
						Name: "print",
						Formals: []*Formal{
							{Name: "msg", Type: "String"},
						},
						ReturnType: "Object",
						Body:       &StringConstant{Value: ""},
					},
				},
			},
		},
	}

	cg := NewCodeGenerator()
	result := cg.Generate(program)

	expected := `
		.data
		_vtable_Main:
		str1: .asciiz "True branch"
		str2: .asciiz "Loop"
		str3: .asciiz ""
		.text
		.globl main
		main:
			sw $fp, 0($sp)
			sw $ra, -4($sp)
			sw $s0, -8($sp)
			move $fp, $sp
			addiu $sp, $sp, -12
			li $a0, 1
			beqz $a0, L1
			la $a0, str1
			li $a0, 1
			j L2
		L1:
			li $a0, 0
		L2:
		L3:
			li $a0, 1
			beqz $a0, L4
			li $a0, 42
			sw $a0, 0($sp)
			addiu $sp, $sp, -4
			la $a0, str2
			jal print
			addiu $sp, $sp, 4
			lw $a0, 0($sp)
			j L3
		L4:
			lw $fp, 0($sp)
			lw $ra, -4($sp)
			lw $s0, -8($sp)
			addiu $sp, $sp, 12
			jr $ra
		print:
			sw $fp, 0($sp)
			sw $ra, -4($sp)
			sw $s0, -8($sp)
			move $fp, $sp
			addiu $sp, $sp, -12
			la $a0, str3
			lw $fp, 0($sp)
			lw $ra, -4($sp)
			lw $s0, -8($sp)
			addiu $sp, $sp, 12
			jr $ra
	`

	testutil.AssemblyMatcher(t, result, expected)
}

func TestMethodDispatch(t *testing.T) {
	program := &Program{
		Classes: []*Class{
			{
				Name:   "Main",
				Parent: "Object",
				Features: []Feature{
					&Method{
						Name: "main",
						Body: &Dispatch{
							Object:     nil,
							MethodName: "foo",
							Arguments: []Expression{
								&IntConstant{Value: 1},
								&StringConstant{Value: "test"},
							},
						},
					},
				},
			},
		},
	}

	cg := NewCodeGenerator()
	result := cg.Generate(program)

	expected := `
		.data
		_vtable_Main:
		str1: .asciiz "test"
		.text
		.globl main
		main:
			sw $fp, 0($sp)
			sw $ra, -4($sp)
			sw $s0, -8($sp)
			move $fp, $sp
			addiu $sp, $sp, -12
			sw $a0, 0($sp)
			addiu $sp, $sp, -4
			li $a0, 1
			sw $a0, 0($sp)
			addiu $sp, $sp, -4
			la $a0, str1
			jal foo
			addiu $sp, $sp, 8
			lw $a0, 0($sp)
			addiu $sp, $sp, 4
			lw $fp, 0($sp)
			lw $ra, -4($sp)
			lw $s0, -8($sp)
			addiu $sp, $sp, 12
			jr $ra
	`

	testutil.AssemblyMatcher(t, result, expected)
}
