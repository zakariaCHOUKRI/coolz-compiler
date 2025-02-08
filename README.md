# Coolz Compiler

A compiler for the COOL (Classroom Object-Oriented Language) programming language, targeting Windows platforms.

## Prerequisites

1. Go 1.16 or later
   - Download from: https://golang.org/dl/
   - Add Go to your PATH environment variable

2. LLVM 13.0 or later
   - Download the pre-built Windows binaries from: https://releases.llvm.org/download.html
   - Install LLVM to `C:\Program Files\LLVM` (recommended)
   - Add `C:\Program Files\LLVM\bin` to your PATH environment variable

## Building the Compiler

1. Clone or create the project:
```cmd
mkdir c:\Users\zczak\Desktop\coolz-compiler
cd c:\Users\zczak\Desktop\coolz-compiler
```

2. Initialize the Go module:
```cmd
go mod init coolz-compiler
```

3. Install dependencies:
```cmd
go get github.com/llir/llvm
```

4. Build the runtime library:
```cmd
cd runtime
gcc -c cool_runtime.c -o cool_runtime.o
ar rcs libcool_runtime.a cool_runtime.o
cd ..
```

5. Build the compiler:
```cmd
go build -o coolzc.exe
```

## Using the Compiler

1. Create a COOL source file (e.g., `example.cl`):
```cool
class Main {
    main() : Int {
        42
    };
};
```

2. Compile your COOL program:
```cmd
coolzc.exe example.cl
```

The compiler will:
- Create a `build` directory
- Generate LLVM IR (`build/output.ll`)
- Create an object file (`build/output.o`)
- Link with runtime to create executable (`build/output.exe`)

3. Run your compiled program:
```cmd
.\build\output.exe
```

## Example Using IO

Here's an example demonstrating basic IO usage:

```cool
class Main inherits IO {
    main() : Object {
        {
            out_string("Enter a number: ");
            let x : Int <- in_int() in {
                out_string("You entered: ").out_int(x).out_string("\n");
            };
        }
    };
};
```

## Project Structure

```
coolz-compiler/
├── main.go           # Compiler entry point
├── optimizer/        # IR optimization
├── codegen/         # Machine code generation
├── runtime/         # Runtime support library
└── build/           # Output directory
```

## Troubleshooting

1. If you get "llc.exe not found":
   - Verify LLVM is installed
   - Check if `C:\Program Files\LLVM\bin` is in PATH
   - Run `llc --version` to verify llc is accessible

2. If linking fails:
   - Ensure `cool_runtime.lib` exists in the runtime directory
   - Verify clang can be called from command line
   - Check if the runtime library path is correct

3. If compilation fails:
   - Check error messages in console
   - Verify COOL syntax in source file
   - Ensure all Go dependencies are installed

## Optimization Levels

The compiler supports four optimization levels:
- `O0` (NoOptimization): No optimizations
- `O1` (BasicOptimization): Basic optimizations
- `O2` (MediumOptimization): Moderate optimizations (default)
- `O3` (HighOptimization): Aggressive optimizations

## Current Limitations

- Basic integer operations only
- Simple class structures
- No inheritance implementation yet
- Limited standard library support

## Contributing

Feel free to submit issues and enhancement requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
