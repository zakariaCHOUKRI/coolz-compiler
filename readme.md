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

A COOL (Classroom Object Oriented Language) compiler implemented in Go with LLVM IR generation

</div>

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

Compile to executable;
```sh
clang -o name output.ll
```

The command `clang -o name output.ll` works on my linux (kali) machine and i think it should work on most machines (i haven't tried on mac but i think it should work) but doesn't work on my windows machine (Windows 10 + visual studio dependencies) so instead I use this one, so in case it doesn't work for you too try this one, and if it also does not work then try to fix clang (the problem is not from the compiler because it generates a ir code that is machine independent that the user has to transform to machine code depending on their machine using clang or other)
```sh
clang-cl output.ll /Fe:name.exe /MD /link /subsystem:console libucrt.lib libcmt.lib legacy_stdio_definitions.lib advapi32.lib shell32.lib user32.lib kernel32.lib msvcrt.lib
```

## 🌟 Features

### 📝 Lexical Analysis

- Full COOL token recognition including:
  - Keywords (class, if, then, else, fi, etc.)
  - Identifiers (Type IDs and Object IDs)
  - Literals (integers, strings, booleans)
  - Operators (+, -, *, /, <-, =, <, <=, etc.)
- Support for single-line comments (`--`) and nested multi-line comments (`(* *)`)
- String literal processing with escape sequences
- Line and column number tracking for error reporting

### 🔍 Parser

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

### 🔎 Semantic Analysis

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

### ⚙️ Code Generation

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

## 🔌 Extensions

### Module System

The COOLZ compiler implements a simple but effective module system through a preprocessing approach. This extension allows for better code organization and reusability while maintaining compatibility with the standard COOL language specification.

#### Syntax
```cool
import modulename;  // imports modulename.cl from the same directory
```

#### How it Works

1. **Preprocessing Stage**: Before lexical analysis, the compiler processes import statements recursively
2. **File Resolution**: Imported files are looked up in the same directory as the importing file
3. **Content Merging**: The preprocessor replaces each import statement with the content of the referenced file
4. **Main Class Handling**: The Main class from imported modules is automatically excluded to prevent multiple entry points
5. **Circular Import Detection**: The system detects and prevents circular dependencies between modules


#### Example Usage

math.cl:
```cool
class Main inherits IO{
    main(): Object {
        {
            out_string("Math module loaded.\n");
        }
    };
};

class GCD inherits IO{
    gcd(a: Int, b: Int) : Int{
        {
            if a = b then
                a
            else if a < b then 
                    gcd(b, a)
                 else gcd(b, a-b)
                 fi
            fi;
        }
    };
};

```

main.cl:
```cool
import math;

class Main inherits IO{
    main(): Object {
        let solver: GCD <- new GCD
        in {
            out_string("gcd(15, 35) = ");
            out_int(solver.gcd(15, 35));
            out_string("\n");

            out_string("gcd(7, 49) = ");
            out_int(solver.gcd(7, 49));
            out_string("\n");

            out_string("gcd(7, 39) = ");
            out_int(solver.gcd(7, 39));
            out_string("\n");
        }
    };
};

```

## 📝 Tests
This project was done using TDD (test driven development) and every layer of it was tested and stress tested to avoid any problems.
You can look at the tests for each layer in the corresponding directories.
And you can run them yourself by using the command:
```bash
go test <directory_name> [-v]
```
with directory name the name of the compiler layer and the -v flag if you want verbose details about the tests.

## 📚 Examples

### [example_strings.cl](examples/example_strings.cl)
Demonstrates string operations including `length()`, `substr()`, and `concat()`.

### [example_operators.cl](examples/example_operators.cl)
Shows usage of various operators and tests Pratt parsing.

### [example_loop.cl](examples/example_loop.cl)
Illustrates loop constructs with `while` loops.

### [example_io.cl](examples/example_io.cl)
Demonstrates input and output operations.

### [example_ifelse.cl](examples/example_ifelse.cl)
Shows conditional statements with `if-then-else`.

### [example_function.cl](examples/example_function.cl)
Illustrates function definitions and recursive calls.

### [example_classes.cl](examples/example_classes.cl)
Demonstrates class inheritance, method overriding, and polymorphism. As well as the functions `copy()`, `type_name()`, and `abort()`.

### [example_chaining.cl](examples/example_chaining.cl)
Shows method chaining with string operations.

### [example_polymorphism.cl](examples/example_polymorphism.cl)
Demonstrates method overriding and polymorphism with various animal classes.

### [example_module.cl](examples/example_module.cl)
Shows the usage of the module system with the `import` statement.

### [example_primes.cl](examples/example_primes.cl)
Shows the usage of multiple cool language constructs (loops, conditions, methods) to generate prime numbers.

---

<div align="center">
Made with ❤️ (and a LOOOOT of effort) by Zakaria CHOUKRI
</div>
