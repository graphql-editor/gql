package introspection

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/slothking-online/gql/client"

	"github.com/stretchr/testify/mock"

	"github.com/aexol/test_util"

	"github.com/graphql-go/graphql"

	"github.com/stretchr/testify/assert"
)

func TestArgGoString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("Arg1: [String!]", Arg{
		Name: "Arg1",
		Type: Type{
			Kind: graphql.TypeKindList,
			OfType: &Type{
				Kind: graphql.TypeKindNonNull,
				OfType: &Type{
					Name: "String",
					Kind: graphql.TypeKindScalar,
				},
			},
		},
	}.GoString())
}

func TestFieldArgsString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("", Field{}.ArgsString())
	assert.Equal("(Arg1: String)", Field{
		Args: []Arg{
			Arg{
				Name: "Arg1",
				Type: Type{
					Name: "String",
				},
			},
		},
	}.ArgsString())
	assert.Equal("(Arg1: String, Arg2: String)", Field{
		Args: []Arg{
			Arg{
				Name: "Arg1",
				Type: Type{
					Name: "String",
				},
			},
			Arg{
				Name: "Arg2",
				Type: Type{
					Name: "String",
				},
			},
		},
	}.ArgsString())
}

func TestFieldGoString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("Field: String", Field{Name: "Field", Type: Type{Name: "String"}}.GoString())
	assert.Equal("Field(Arg1: String): String", Field{
		Name: "Field",
		Type: Type{
			Name: "String",
		},
		Args: []Arg{
			Arg{
				Name: "Arg1",
				Type: Type{
					Name: "String",
				},
			},
		},
	}.GoString())
}

func TestFieldArgNames(t *testing.T) {
	assert := assert.New(t)
	assert.Equal([]string(nil), Field{}.ArgNames())
	assert.Equal([]string{"arg1", "arg2"}, Field{
		Args: []Arg{
			Arg{
				Name: "arg1",
			},
			Arg{
				Name: "arg2",
			},
		},
	}.ArgNames())
}

type testCaseTypeChecks struct {
	in                                                                           Type
	named, scalar, enum, input, interf, object, union, nonnull, list, valid, ref bool
}

func (tt testCaseTypeChecks) test(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(tt.named, tt.in.Named(), "Named test with type %v", tt.in)
	assert.Equal(tt.scalar, tt.in.Scalar(), "Scalar test with type %v", tt.in)
	assert.Equal(tt.enum, tt.in.Enum(), "Enum test with type %v", tt.in)
	assert.Equal(tt.input, tt.in.Input(), "Input test with type %v", tt.in)
	assert.Equal(tt.interf, tt.in.Interface(), "Interface test with type %v", tt.in)
	assert.Equal(tt.object, tt.in.Object(), "Object test with type %v", tt.in)
	assert.Equal(tt.union, tt.in.Union(), "Union test with type %v", tt.in)
	assert.Equal(tt.nonnull, tt.in.NonNull(), "NonNull test with type %v", tt.in)
	assert.Equal(tt.list, tt.in.List(), "List test with type %v", tt.in)
	assert.Equal(tt.valid, tt.in.Valid(), "Valid test with type %v", tt.in)
	assert.Equal(tt.ref, tt.in.TypeRef(), "ref test with type %v", tt.in)
}

func TestTypeChecks(t *testing.T) {
	data := []testCaseTypeChecks{
		{
			in:    Type{},
			valid: false,
		},
		{
			in:    Type{Kind: "some-named-kind"},
			named: true,
			valid: true,
		},
		{
			in:     Type{Kind: graphql.TypeKindScalar},
			named:  true,
			scalar: true,
			valid:  true,
		},
		{
			in:    Type{Kind: graphql.TypeKindEnum},
			named: true,
			enum:  true,
			valid: true,
		},
		{
			in:    Type{Kind: graphql.TypeKindInputObject},
			named: true,
			input: true,
			valid: true,
			ref:   true,
		},
		{
			in:     Type{Kind: graphql.TypeKindInterface},
			named:  true,
			interf: true,
			valid:  true,
			ref:    true,
		},
		{
			in:     Type{Kind: graphql.TypeKindObject},
			named:  true,
			object: true,
			valid:  true,
			ref:    true,
		},
		{
			in:    Type{Kind: graphql.TypeKindUnion},
			named: true,
			union: true,
			valid: true,
			ref:   true,
		},
		{
			in:      Type{Kind: graphql.TypeKindNonNull},
			nonnull: true,
			valid:   true,
		},
		{
			in:    Type{Kind: graphql.TypeKindList},
			list:  true,
			valid: true,
		},
		{
			in:    Type{Kind: graphql.TypeKindInputObject, Fields: []Field{Field{}}},
			named: true,
			input: true,
			valid: true,
		},
		{
			in:     Type{Kind: graphql.TypeKindInterface, Fields: []Field{Field{}}},
			named:  true,
			interf: true,
			valid:  true,
		},
		{
			in:     Type{Kind: graphql.TypeKindObject, Fields: []Field{Field{}}},
			named:  true,
			object: true,
			valid:  true,
		},
		{
			in:    Type{Kind: graphql.TypeKindUnion, PossibleTypes: []Type{Type{}}},
			named: true,
			union: true,
			valid: true,
		},
		{
			in:    Type{Name: "name"},
			valid: false,
			ref:   true,
		},
	}
	for _, tt := range data {
		tt.test(t)
	}
}

func TestGetOfTypeLeaf(t *testing.T) {
	assert := assert.New(t)
	tt := Type{
		Kind: "some-kind",
		Name: "some-name",
	}
	assert.Equal(tt.GetOfTypeLeaf(), tt)
	assert.Equal(Type{OfType: &tt}.GetOfTypeLeaf(), tt)
	assert.Equal(Type{OfType: &Type{OfType: &tt}}.GetOfTypeLeaf(), tt)
}

func TestTypeDeref(t *testing.T) {
	assert := assert.New(t)
	tt := Type{
		Kind: "some-kind",
		Name: "some-name",
	}
	tt2 := Type{
		Kind: "some-kind2",
		Name: "some-name2",
	}
	tt3 := Type{
		Kind: "some-kind3",
		Name: "some-name3",
	}
	types := []Type{tt, tt2}

	assert.Equal(tt.Deref(types), tt)
	assert.Equal(tt2.Deref(types), tt2)
	assert.Equal(tt3.Deref(types), Type{})
}

type testCaseTypeForPath struct {
	inType  Type
	schema  Schema
	path    []string
	outType Type
	ok      bool
}

func (tt testCaseTypeForPath) test(t *testing.T) {
	assert := assert.New(t)
	outT, ok := tt.inType.TypeForPath(tt.schema, tt.path)
	assert.Equal(tt.outType, outT)
	assert.Equal(tt.ok, ok)
}

func TestTypeForPath(t *testing.T) {
	schema := Schema{
		Types: []Type{
			Type{
				Name: "type1",
				Kind: graphql.TypeKindObject,
				Fields: []Field{
					Field{
						Name: "field",
						Type: Type{
							Kind: graphql.TypeKindObject,
							Name: "type2",
						},
					},
					Field{
						Name: "nnfield",
						Type: Type{
							Kind: graphql.TypeKindNonNull,
							OfType: &Type{
								Kind: graphql.TypeKindObject,
								Name: "type2",
							},
						},
					},
				},
			},
			Type{
				Name: "type2",
				Kind: graphql.TypeKindObject,
				Fields: []Field{
					Field{
						Name: "field",
						Type: Type{
							Kind: graphql.TypeKindObject,
							Name: "type3",
						},
					},
				},
			},
			Type{
				Name: "type3",
				Kind: graphql.TypeKindObject,
				Fields: []Field{
					Field{
						Name: "field",
						Type: Type{
							Name: "String",
							Kind: graphql.TypeKindScalar,
						},
					},
				},
			},
		},
	}
	data := []testCaseTypeForPath{
		{
			inType: Type{
				Name: "type1",
				Kind: graphql.TypeKindObject,
			},
			schema:  schema,
			outType: schema.Types[0],
			ok:      true,
		},
		{
			inType: Type{
				Name: "type1",
				Kind: graphql.TypeKindObject,
			},
			path:    []string{"field"},
			schema:  schema,
			outType: schema.Types[1],
			ok:      true,
		},
		{
			inType: Type{
				Name: "type1",
				Kind: graphql.TypeKindObject,
			},
			path:    []string{"nnfield"},
			schema:  schema,
			outType: schema.Types[1],
			ok:      true,
		},
		{
			inType: Type{
				Name: "type1",
				Kind: graphql.TypeKindObject,
			},
			path:    []string{"fiel"},
			schema:  schema,
			outType: Type{},
			ok:      false,
		},
		{
			inType: Type{
				Name: "type1",
				Kind: graphql.TypeKindObject,
			},
			path:    []string{"field", "field", "field"},
			schema:  schema,
			outType: schema.Types[2].Fields[0].Type,
			ok:      true,
		},
	}
	for _, tt := range data {
		tt.test(t)
	}
}

func TestTypeGoString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("String", Type{Name: "String"}.GoString())
	assert.Equal("String!", Type{
		Kind: graphql.TypeKindNonNull,
		OfType: &Type{
			Name: "String",
		},
	}.GoString())
	assert.Equal("[String]", Type{
		Kind: graphql.TypeKindList,
		OfType: &Type{
			Name: "String",
		},
	}.GoString())
	assert.Equal("[String]!", Type{
		Kind: graphql.TypeKindNonNull,
		OfType: &Type{
			Kind: graphql.TypeKindList,
			OfType: &Type{
				Name: "String",
			},
		},
	}.GoString())
	assert.Equal("[String!]!", Type{
		Kind: graphql.TypeKindNonNull,
		OfType: &Type{
			Kind: graphql.TypeKindList,
			OfType: &Type{
				Kind: graphql.TypeKindNonNull,
				OfType: &Type{
					Name: "String",
				},
			},
		},
	}.GoString())
}

func TestSchemaTypeForPath(t *testing.T) {
	assert := assert.New(t)
	tt := Type{
		Kind: graphql.TypeKindObject,
		Name: "some-type",
		Fields: []Field{
			Field{
				Name: "field",
				Type: Type{
					Kind: graphql.TypeKindScalar,
					Name: "some-simple-type",
				},
			},
		},
	}
	schema := Schema{
		Types: []Type{
			tt,
		},
		QueryType: Type{
			Name: "some-type",
			Kind: graphql.TypeKindObject,
		},
		MutationType: Type{
			Name: "some-type",
			Kind: graphql.TypeKindObject,
		},
		SubscriptionType: Type{
			Name: "some-type",
			Kind: graphql.TypeKindObject,
		},
	}
	tp, ok := schema.TypeForPath([]string{"query", "field"})
	assert.Equal(tt.Fields[0].Type, tp)
	assert.True(ok)
	tp, ok = schema.TypeForPath([]string{"mutation", "field"})
	assert.Equal(tt.Fields[0].Type, tp)
	assert.True(ok)
	tp, ok = schema.TypeForPath([]string{"subscription", "field"})
	assert.Equal(tt.Fields[0].Type, tp)
	assert.True(ok)
}

func TestSchemaFieldForPath(t *testing.T) {
	assert := assert.New(t)
	tt := Type{
		Kind: graphql.TypeKindObject,
		Name: "some-type",
		Fields: []Field{
			Field{
				Name: "field",
				Type: Type{
					Kind: graphql.TypeKindScalar,
					Name: "some-simple-type",
				},
			},
		},
	}
	schema := Schema{
		Types: []Type{
			tt,
		},
		QueryType: Type{
			Name: "some-type",
			Kind: graphql.TypeKindObject,
		},
		MutationType: Type{
			Name: "some-type",
			Kind: graphql.TypeKindObject,
		},
		SubscriptionType: Type{
			Name: "some-type",
			Kind: graphql.TypeKindObject,
		},
	}
	f, ok := schema.FieldForPath([]string{"query", "field"})
	assert.Equal(tt.Fields[0], f)
	assert.True(ok)
	f, ok = schema.FieldForPath([]string{"mutation", "field"})
	assert.Equal(tt.Fields[0], f)
	assert.True(ok)
	f, ok = schema.FieldForPath([]string{"subscription", "field"})
	assert.Equal(tt.Fields[0], f)
	assert.True(ok)
	_, ok = schema.FieldForPath([]string{"query"})
	assert.False(ok)
	_, ok = schema.FieldForPath([]string{"mutation"})
	assert.False(ok)
	_, ok = schema.FieldForPath([]string{"subscription"})
	assert.False(ok)
}

func TestGetTypeInfo(t *testing.T) {
	expectedType := Type{
		Name: "some-name",
		Kind: "some-kind",
	}
	jsonType, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"__type": expectedType,
		},
	})
	assert := assert.New(t)
	transport := &test_util.MockRoundTripper{}
	transport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		body, _ := ioutil.ReadAll(req.Body)
		expectedBody, _ := json.Marshal(client.Raw{
			Query: typeInfoQuery,
			Variables: map[string]interface{}{
				"typeName": "some-name",
			},
		})
		return assert.JSONEq(string(body), string(expectedBody))
	})).Return(&http.Response{
		Status:     "OK",
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(jsonType)),
	}, nil)
	gqlType, err := GetTypeInfo(&client.Client{
		Endpoint: "http://example.com/graphql",
		Client: http.Client{
			Transport: transport,
		},
	}, "some-name", nil)
	assert.Equal(expectedType, gqlType)
	assert.NoError(err)
}

func TestGetSchemaTypes(t *testing.T) {
	expectedSchema := Schema{
		Types: []Type{
			Type{
				Name: "some-name",
				Kind: "some-kind",
			},
		},
		QueryType: Type{
			Name: "some-name",
			Kind: "some-kind",
		},
	}
	jsonSchema, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"__schema": expectedSchema,
		},
	})
	assert := assert.New(t)
	transport := &test_util.MockRoundTripper{}
	transport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		body, _ := ioutil.ReadAll(req.Body)
		expectedBody, _ := json.Marshal(client.Raw{
			Query: schemaInfoQuery,
		})
		return assert.JSONEq(string(body), string(expectedBody))
	})).Return(&http.Response{
		Status:     "OK",
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(jsonSchema)),
	}, nil)
	schema, err := GetSchemaTypes(&client.Client{
		Endpoint: "http://example.com/graphql",
		Client: http.Client{
			Transport: transport,
		},
	}, nil)
	assert.Equal(expectedSchema, schema)
	assert.NoError(err)
}
