package main

import (
	"coolz-compiler/builder"
	"coolz-compiler/visualizer"
	"fmt"
	"log"
	"os"
)

func main() {
	// Create a new builder
	b := builder.NewBuilder()

	// Build an example AST for a simple COOL program
	program := b.Class("Main").
		Method("main").
		ReturnType("Object").
		Body(
			b.Block().
				Add(b.Let().
					Bind("x", "Int", b.Int(42)).
					Bind("y", "String", b.String("Hello, World!")).
					In(
						b.Block().
							Add(b.Call(b.String("Output: "), "concat").Args(b.String("42"))).
							Add(b.Binary(
								b.Int(40),
								"+",
								b.Int(2),
							)).
							Build(),
					),
				).
				Build(),
		).
		Build()

	// Convert to JSON
	json, err := visualizer.ToJSON(program)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("JSON representation:")
	fmt.Println(json)

	// Generate DOT graph
	dot := visualizer.ToDOT(program)
	fmt.Println("\nDOT representation:")
	fmt.Println(dot)

	// Save DOT to file
	err = os.WriteFile("ast.dot", []byte(dot), 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nSaved DOT graph to ast.dot")
	fmt.Println("To generate PNG: dot -Tpng ast.dot -o ast.png")

	// Pretty print
	pretty := visualizer.ToPrettyString(program)
	fmt.Println("\nPretty printed AST:")
	fmt.Println(pretty)
}
