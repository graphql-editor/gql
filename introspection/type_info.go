// Package introspection adds very basic remote GraphQL
// introspection funcionality
package introspection

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"

	"github.com/slothking-online/gql/client"
)

// Arg is graphql field argument
type Arg struct {
	// Name is a schema defined argument name
	Name string `json:"name,omitempty"`
	// Description is a schema defined argument description
	Description string `json:"description,omitempty"`
	// Type is a type or type reference defined by schema
	Type         Type   `json:"type,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
}

// GoString formats argument as it would apear in schema
func (a Arg) GoString() string {
	return fmt.Sprintf("%s: %s", a.Name, a.Type.GoString())
}

// Field is graphql field
type Field struct {
	// Args is a list of arguments for a field defined by schema
	Args []Arg `json:"args,omitempty"`
	// Name is a field name defined by schema
	Name string `json:"name,omitempty"`
	// Type is a type or type reference defined by schema
	Type Type `json:"type,omitempty"`
	// Description is a schema defined argument description
	Description string `json:"description,omitempty"`
}

// ArgsString formats field arguments as they would apear in schema
func (f Field) ArgsString() string {
	if len(f.Args) == 0 {
		return ""
	}
	buf := &bytes.Buffer{}
	sep := "("
	for _, a := range f.Args {
		fmt.Fprintf(buf, "%s%s", sep, a.GoString()) // nolint: errcheck
		sep = ", "
	}
	fmt.Fprint(buf, ")") // nolint: errcheck
	return buf.String()
}

// GoString formats field as it would apear in schema
func (f Field) GoString() string {
	return fmt.Sprintf("%s%s: %s", f.Name, f.ArgsString(), f.Type.GoString())
}

// ArgNames is an array of arguments
// that this field accepts
func (f Field) ArgNames() []string {
	if len(f.Args) == 0 {
		return nil
	}
	args := make([]string, 0, len(f.Args))
	for _, a := range f.Args {
		args = append(args, a.Name)
	}
	return args
}

// Type is graphql type definition
type Type struct {
	// Name is a type or type reference name
	Name string `json:"name,omitempty"`
	// Fields is a list of fields for this type, defined by schema
	// Object and Interface only
	Fields []Field `json:"fields,omitempty"`
	// Kind is a type kind, refer to https://godoc.org/github.com/graphql-go/graphql#pkg-constants
	// for valid list of type kinds
	Kind string `json:"kind,omitempty"`
	// Description is a schema defined argument description
	Description string `json:"description,omitempty"`
	// PossibleTypes is a list of possible types defined by schema
	// only valid for Union and Interface kidns
	PossibleTypes []Type `json:"possibleTypes,omitempty"`
	// OfType is a type reference which this type wraps
	// only valid for NonNull and List type kinds
	OfType *Type `json:"ofType,omitempty"`
}

// Named returns true if type is named, as in, not NonNull or List
func (t Type) Named() bool {
	return t.Valid() && (t.OfType == nil && !t.NonNull() && !t.List())
}

// Scalar returns true if type is of scalar type
func (t Type) Scalar() bool {
	return t.Kind == graphql.TypeKindScalar
}

// Enum returns true if type is an Enum
func (t Type) Enum() bool {
	return t.Kind == graphql.TypeKindEnum
}

// Input returns true if type is an Input
func (t Type) Input() bool {
	return t.Kind == graphql.TypeKindInputObject
}

// Interface returns true if type is an Interface
func (t Type) Interface() bool {
	return t.Kind == graphql.TypeKindInterface
}

// Object returns true if type is an Object
func (t Type) Object() bool {
	return t.Kind == graphql.TypeKindObject
}

// Union returns true type is an Union
func (t Type) Union() bool {
	return t.Kind == graphql.TypeKindUnion
}

// NonNull returns true if type is NonNull
func (t Type) NonNull() bool {
	return t.Kind == graphql.TypeKindNonNull
}

// List returns true if type is a List
func (t Type) List() bool {
	return t.Kind == graphql.TypeKindList
}

// Valid returns true if type is valid type
func (t Type) Valid() bool {
	return t.Kind != ""
}

// GetOfTypeLeaf follows solves all NonNull/List types
// until it reaches named type
func (t Type) GetOfTypeLeaf() Type {
	tt := t
	for tt.OfType != nil {
		tt = *tt.OfType
	}
	return tt
}

// TypeRef returns true if type is a type reference
func (t Type) TypeRef() bool {
	// Inputs/Interfaces/Objects without fields, Unions without possible
	// types and types without kind but with name are assumed to be a
	// reference
	return ((t.Input() || t.Interface() || t.Object()) && len(t.Fields) == 0) ||
		(t.Union() && len(t.PossibleTypes) == 0) ||
		(!t.Valid() && t.Name != "")

}

// Deref finds a type matching this type reference
// in a list of types
func (t Type) Deref(types []Type) Type {
	for _, tt := range types {
		if t.Name == tt.Name {
			return tt
		}
	}
	return Type{}
}

// TypeForPath seeks a type in schema matching resolve path.
// First element of path must match one of type fields
func (t Type) TypeForPath(schema Schema, path []string) (tt Type, ok bool) {
	t = t.GetOfTypeLeaf()
	if t.TypeRef() {
		t = t.Deref(schema.Types)
	}
	// Exact match found
	// return self
	if len(path) == 0 {
		tt = t
		ok = true
		return
	}
	for _, field := range t.Fields {
		if field.Name == path[0] {
			// One of the field matches requested path
			// return it
			ft, fok := field.Type.TypeForPath(schema, path[1:])
			if fok {
				tt = ft
				ok = fok
				return
			}
		}
	}
	// no match found
	return
}

// GoString formats type to a string as it would appear
// in schema
func (t Type) GoString() string {
	switch {
	case t.NonNull():
		return t.OfType.GoString() + "!"
	case t.List():
		return "[" + t.OfType.GoString() + "]"
	default:
		return t.Name
	}
}

func ofTypeDepthFunc(depth int) string {
	if depth == 0 {
		return "ofType {kind name}"
	}
	depth--
	return fmt.Sprintf("ofType {kind name %s }", ofTypeDepthFunc(depth))
}

var (
	fragmentFullType = `fragment FullType on __Type {
    kind
    name
    description
    fields(includeDeprecated: true) {
        name
        description
        args {
        ...InputValue
        }
        type {
        ...TypeRef
        }
        isDeprecated
        deprecationReason
    }
    inputFields {
        ...InputValue
    }
    interfaces {
        ...TypeRef
    }
    enumValues(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
    }
    possibleTypes {
        ...TypeRef
    }
}`
	fragmentInputValue = `fragment InputValue on __InputValue {
    name
    description
    type { ...TypeRef }
    defaultValue
}`
	fragmentTypeRef = `fragment TypeRef on __Type {
    kind
    name
    ` + ofTypeDepthFunc(7) + ` 
}`
	introspectionQuery = `query IntrospectionQuery {
    __schema {
        queryType { name }
        mutationType { name }
        subscriptionType { name }
        types {
        ...FullType
        }
        directives {
        name
        description
        locations
        args {
            ...InputValue
        }
        }
    }
}`

	typeInfoQuery = `
query op($typeName: String!){
    __type(name: $typeName) { ...FullType }
}
` + fragmentFullType + "\n" + fragmentInputValue + "\n" + fragmentTypeRef
	schemaInfoQuery = introspectionQuery + "\n" + fragmentFullType + "\n" + fragmentInputValue + "\n" + fragmentTypeRef
)

// GetTypeInfo fetches type information from remote endpoint
func GetTypeInfo(cli *client.Client, typeName string, header http.Header) (Type, error) {
	r := client.Raw{
		Query:  typeInfoQuery,
		Header: header,
		Variables: map[string]interface{}{
			"typeName": typeName,
		},
	}
	out := struct {
		Type Type `json:"__type,omitempty"`
	}{}
	if _, err := cli.Raw(r, &out); err != nil {
		return Type{}, err
	}
	return out.Type, nil
}

// Schema is a representation of remote schema
// returned by schema introspection
type Schema struct {
	// Types defined in schema
	Types []Type `json:"types,omitempty"`
	// QueryType is a reference to a type of query root operation
	QueryType Type `json:"queryType"`
	// MutationType is a reference to a type of mutation root operation
	MutationType Type `json:"mutationType,omitempty"`
	// SubscriptionType is a reference to a type of subscription root operation
	SubscriptionType Type `json:"subscriptionType,omitempty"`
}

// TypeForPath finds a type to which path would resolve to
func (s Schema) TypeForPath(path []string) (t Type, ok bool) {
	if len(path) == 0 {
		return
	}
	switch path[0] {
	case "query":
		t = s.QueryType
	case "mutation":
		t = s.MutationType
	case "subscription":
		t = s.SubscriptionType
	default:
		return
	}
	return t.TypeForPath(s, path[1:])
}

// FieldForPath finds a field to which path would resolve to
func (s Schema) FieldForPath(path []string) (f Field, ok bool) {
	if len(path) <= 1 {
		return
	}
	t, ok := s.TypeForPath(path[:len(path)-1])
	if !ok {
		return
	}
	for _, field := range t.Fields {
		if field.Name == path[len(path)-1] {
			f = field
			ok = true
			break
		}
	}
	return
}

// GetSchemaTypes runs introspection query on remote endpoint returning schema
func GetSchemaTypes(cli *client.Client, header http.Header) (Schema, error) {
	r := client.Raw{
		Query:  schemaInfoQuery,
		Header: header,
	}
	out := struct {
		Schema Schema `json:"__schema"`
	}{}
	if _, err := cli.Raw(r, &out); err != nil {
		return Schema{}, err
	}
	return out.Schema, nil
}
