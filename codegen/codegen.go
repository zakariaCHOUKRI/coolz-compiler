package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/llir/llvm/ir"
)

// Target represents the target architecture and platform
type Target struct {
	Architecture string
	Platform     string
	CPU          string
	Features     string
}

// CodeGenerator handles the generation of machine code
type CodeGenerator struct {
	target    Target
	outputDir string
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator(target Target, outputDir string) *CodeGenerator {
	return &CodeGenerator{
		target:    target,
		outputDir: outputDir,
	}
}

// Generate converts LLVM IR to machine code and produces an executable
func (g *CodeGenerator) Generate(module *ir.Module) error {
	// Ensure output directory exists
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Write LLVM IR to a file
	irFile := filepath.Join(g.outputDir, "output.ll")
	if err := g.writeIR(module, irFile); err != nil {
		return fmt.Errorf("failed to write IR: %v", err)
	}

	// Convert IR to object file using llc
	objFile := filepath.Join(g.outputDir, "output.o")
	if err := g.generateObject(irFile, objFile); err != nil {
		return fmt.Errorf("failed to generate object file: %v", err)
	}

	// Link object file to create executable
	exeFile := filepath.Join(g.outputDir, "output.exe") // Use .exe extension on Windows
	if err := g.linkExecutable(objFile, exeFile); err != nil {
		return fmt.Errorf("failed to link executable: %v", err)
	}

	return nil
}

// writeIR writes the LLVM IR to a file
func (g *CodeGenerator) writeIR(module *ir.Module, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(module.String())
	return err
}

// generateObject converts LLVM IR to an object file using llc
func (g *CodeGenerator) generateObject(irFile, objFile string) error {
	llcPath := g.findLLCPath()
	if llcPath == "" {
		return fmt.Errorf("llc not found in PATH or LLVM directory")
	}

	args := []string{
		"-filetype=obj",
		"-o", objFile,
		irFile,
	}

	if g.target.CPU != "" {
		args = append(args, "-mcpu="+g.target.CPU)
	}

	cmd := exec.Command(llcPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// linkExecutable links the object file to create the final executable
func (g *CodeGenerator) linkExecutable(objFile, exeFile string) error {
	clangPath := g.findClangPath()
	if clangPath == "" {
		return fmt.Errorf("clang not found in PATH or LLVM directory")
	}

	runtimeObj := filepath.Join(g.findRuntimePath(), "cool_runtime.o") // Use .o file directly) // Changed from cool_runtime.lib

	args := []string{
		objFile,
		runtimeObj, // Link with .o file
		"-o", exeFile,
	}

	cmd := exec.Command(clangPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *CodeGenerator) findLLCPath() string {
	// Check common LLVM installation paths on Windows
	paths := []string{
		"C:\\Program Files\\LLVM\\bin\\llc.exe",
		"C:\\Program Files (x86)\\LLVM\\bin\\llc.exe",
	}

	// Check if llc is in PATH
	if path, err := exec.LookPath("llc.exe"); err == nil {
		return path
	}

	// Check common installation paths
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func (g *CodeGenerator) findClangPath() string {
	// Similar to findLLCPath but for clang
	paths := []string{
		"C:\\Program Files\\LLVM\\bin\\clang.exe",
		"C:\\Program Files (x86)\\LLVM\\bin\\clang.exe",
	}

	if path, err := exec.LookPath("clang.exe"); err == nil {
		return path
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func (g *CodeGenerator) findRuntimePath() string {
	// Look for runtime library in the runtime directory first, then standard locations
	paths := []string{
		"runtime",                             // Local runtime directory
		filepath.Join(g.outputDir, "runtime"), // Build directory runtime
		"C:\\Program Files\\LLVM\\lib",        // LLVM standard locations
		"C:\\Program Files (x86)\\LLVM\\lib",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// DefaultTarget returns a default target configuration
func DefaultTarget() Target {
	return Target{
		Architecture: "x86_64",
		Platform:     "windows", // Changed from linux to windows
		CPU:          "x86-64",
		Features:     "",
	}
}
