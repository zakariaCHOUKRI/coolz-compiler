package ast

import (
	"encoding/json"
)

// CustomMarshaler is an interface for types that need custom JSON marshaling
type CustomMarshaler interface {
	MarshalJSON() ([]byte, error)
}

// MarshalJSON implements custom JSON marshaling for TypeIdentifier
func (ti *TypeIdentifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  "TypeIdentifier",
		Value: ti.Value,
	})
}

// MarshalJSON implements custom JSON marshaling for ObjectIdentifier
func (oi *ObjectIdentifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  "ObjectIdentifier",
		Value: oi.Value,
	})
}

// MarshalJSON implements custom JSON marshaling for Program
func (p *Program) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string   `json:"type"`
		Classes []*Class `json:"classes"`
	}{
		Type:    "Program",
		Classes: p.Classes,
	})
}

// MarshalJSON implements custom JSON marshaling for Class
func (c *Class) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string    `json:"type"`
		Name     string    `json:"name"`
		Parent   string    `json:"parent,omitempty"`
		Features []Feature `json:"features"`
	}{
		Type:     "Class",
		Name:     c.Name.Value,
		Parent:   stringOrEmpty(c.Parent),
		Features: c.Features,
	})
}

// MarshalJSON implements custom JSON marshaling for Method
func (m *Method) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string     `json:"type"`
		Name    string     `json:"name"`
		Formals []*Formal  `json:"formals"`
		RetType string     `json:"returnType"`
		Body    Expression `json:"body"`
	}{
		Type:    "Method",
		Name:    m.Name.Value,
		Formals: m.Formals,
		RetType: m.Type.Value,
		Body:    m.Body,
	})
}

// Helper function for optional string fields
func stringOrEmpty(ti *TypeIdentifier) string {
	if ti == nil {
		return ""
	}
	return ti.Value
}

// Add MarshalJSON for all expression types
func (il *IntegerLiteral) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Value int64  `json:"value"`
	}{
		Type:  "IntegerLiteral",
		Value: il.Value,
	})
}

func (sl *StringLiteral) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  "StringLiteral",
		Value: sl.Value,
	})
}

// Add similar MarshalJSON methods for all other AST node types...
// For brevity, I'm showing just a few examples. You would need to implement
// this for each type in ast.go
