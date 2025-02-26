# COOLZ Compiler

<div align="center">

![COOLZ Logo](https://img.shields.io/badge/COOLZ-Compiler-blue)
![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)
![LLVM](https://img.shields.io/badge/LLVM-262D3A?logo=llvm&logoColor=white)

```
  ..|'''.|  ..|''||    ..|''||   '||'      |'''''||  
.|'     '  .|'    ||  .|'    ||   ||           .|'   
||         ||      || ||      ||  ||          ||     
'|.      . '|.     || '|.     ||  ||        .|'      
 ''|....'   ''|...|'   ''|...|'  .||.....| ||......| 

                                     ||   ||                  
  ....    ...   .. .. ..   ... ...  ...   ||    ....  ... ..  
.|   '' .|  '|.  || || ||   ||'  ||  ||   ||  .|...||  ||' '' 
||      ||   ||  || || ||   ||    |  ||   ||  ||       ||     
 '|...'  '|..|' .|| || ||.  ||...'  .||. .||.  '|...' .||.    
                            ||                                
                           ''''                               
```

A modern COOL (Classroom Object Oriented Language) compiler implemented in Go with LLVM IR generation

</div>

---

## 🌟 Features

<details>
<summary><b>📝 Lexical Analysis</b></summary>

- Full COOL token recognition including:
  - Keywords (class, if, then, else, fi, etc.)
  - Identifiers (Type IDs and Object IDs)
  - Literals (integers, strings, booleans)
  - Operators (+, -, *, /, <-, =, <, <=, etc.)
- Support for single-line comments (`--`) and nested multi-line comments (`(* *)`)
- String literal processing with escape sequences
- Line and column number tracking for error reporting
</details>

<details>
<summary><b>🔍 Parser</b></summary>

- Complete COOL syntax support including:
  - Class definitions with inheritance
  - Method and attribute declarations
  - Block expressions
  - Let expressions with multiple bindings
  - Case expressions
  - If-then-else expressions
  - While loops
  - Object dispatch (both dynamic and static)
  - Binary and unary operations
  - Object instantiation (new)
  - Self and void expressions
- Proper operator precedence handling
- Detailed error reporting
- AST generation with full source location information
- Fully functional Pratt parsing
</details>

<details>
<summary><b>🔎 Semantic Analysis</b></summary>

- Full type checking system including:
  - Class hierarchy validation
  - Method override checking
  - Type conformance verification
  - Self and SELF_TYPE handling
- Symbol table management with:
  - Proper scoping rules
  - Method and attribute tracking
  - Inheritance-aware symbol lookup
- Main class and main() method validation
- Multiple error detection and reporting
- Cycle detection in inheritance graphs
</details>

<details>
<summary><b>⚙️ Code Generation</b></summary>

- LLVM IR generation with:
  - Full class layout generation
  - Virtual method tables
  - Proper object instantiation
  - Garbage collection support (planned)
- Built-in class support:
  - Object with abort(), type_name(), and copy() methods
  - IO with in_string(), in_int(), out_string(), and out_int() methods
  - String with length(), concat(), and substr() methods
  - Int and Bool with proper operations
- Expression generation for all COOL features
</details>

## 🚀 Getting Started

### Prerequisites

- Go (Golang)
- Clang

### Building

```sh
go build -o coolz main.go
```

### Usage

Generate LLVM IR code:
```sh
./coolz -o output.ll input.cl
```

Compile to executable (Unix-like systems):
```sh
clang -o name output.ll
```

Compile to executable (Windows with VS dependencies):
```sh
clang-cl output.ll /Fe:name.exe /MD /link /subsystem:console libucrt.lib libcmt.lib legacy_stdio_definitions.lib advapi32.lib shell32.lib user32.lib kernel32.lib msvcrt.lib
```

---

<div align="center">
Made with ❤️ (and a LOOOT of effort) by Zakaria CHOUKRI
</div>