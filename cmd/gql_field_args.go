package cmd

import (
	"fmt"

	"github.com/slothking-online/gql/introspection"

	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

type FieldCommandValue interface {
	// GraphQL-like string of value
	String() string
	// return GraphQL type name of the argument
	Type() string
}

type FieldCommandArgument interface {
	FieldCommandValue
	// argument value
	Value() interface{}
	// argument name
	Name() string
}

// represnets GraphQL field argument of type Int
type FieldCommandIntArgument struct {
	name  string
	value int64
}

func (f *FieldCommandIntArgument) String() string {
	return fmt.Sprintf("%d", f.value)
}

func (f *FieldCommandIntArgument) Type() string {
	return "Int"
}

func (f *FieldCommandIntArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandIntArgument) Name() string {
	return f.name
}

type FieldCommandFloatArgument struct {
	name  string
	value float64
}

func (f *FieldCommandFloatArgument) String() string {
	return fmt.Sprintf("%f", f.value)
}

func (f *FieldCommandFloatArgument) Type() string {
	return "Float"
}

func (f *FieldCommandFloatArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandFloatArgument) Name() string {
	return f.name
}

type FieldCommandStringArgument struct {
	name  string
	value string
}

func (f *FieldCommandStringArgument) String() string {
	return fmt.Sprintf("\"%s\"", f.value)
}

func (f *FieldCommandStringArgument) Type() string {
	return "String"
}

func (f *FieldCommandStringArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandStringArgument) Name() string {
	return f.name
}

type FieldCommandBooleanArgument struct {
	name  string
	value bool
}

func (f *FieldCommandBooleanArgument) String() string {
	return fmt.Sprintf("%t", f.value)
}

func (f *FieldCommandBooleanArgument) Type() string {
	return "Boolean"
}

func (f *FieldCommandBooleanArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandBooleanArgument) Name() string {
	return f.name
}

type FieldCommandIDArgument struct {
	name  string
	value string
}

func (f *FieldCommandIDArgument) String() string {
	return fmt.Sprintf("\"%s\"", f.value)
}

func (f *FieldCommandIDArgument) Type() string {
	return "ID"
}

func (f *FieldCommandIDArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandIDArgument) Name() string {
	return f.name
}

type FieldCommandEnumArgument struct {
	name     string
	value    string
	enumName string
}

func (f *FieldCommandEnumArgument) String() string {
	return f.value
}

func (f *FieldCommandEnumArgument) Type() string {
	return f.enumName
}

func (f *FieldCommandEnumArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandEnumArgument) Name() string {
	return f.name
}

type FieldCommandNonNullArgument struct {
	FieldCommandArgument
}

func (f *FieldCommandNonNullArgument) Type() string {
	return f.FieldCommandArgument.Type() + "!"
}

type nonNullSetter struct {
	f *FieldCommandNonNullArgument
}

func (n *nonNullSetter) Set(cmd *cobra.Command, f FieldCommandArgument) {
	Set(cmd, (*fieldCommandNonNullArgument)(n.f))
	cmd.MarkFlagRequired(FieldCommandArgName(n.f))
}

func (f *FieldCommandNonNullArgument) Value() interface{} {
	return &nonNullSetter{f: f}
}

type fieldCommandNonNullArgument FieldCommandNonNullArgument

func (f *fieldCommandNonNullArgument) Type() string {
	return f.FieldCommandArgument.Type() + "!"
}

type Setter interface {
	Set(*cobra.Command, FieldCommandArgument)
}

type FieldCommandInputArgument struct {
	name string
	// should there be some client side parsing to
	// check for valid input?
	value     string
	inputName string
}

func (f *FieldCommandInputArgument) String() string {
	return f.value
}

func (f *FieldCommandInputArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandInputArgument) Type() string {
	return f.inputName
}

func (f *FieldCommandInputArgument) Name() string {
	return f.name
}

type FieldCommandListArgument struct {
	FieldCommandArgument
	value string
}

func (f *FieldCommandListArgument) String() string {
	return f.value
}

func (f *FieldCommandListArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandListArgument) Type() string {
	return "[" + f.FieldCommandArgument.Type() + "]"
}

type FieldCommandCustomScalarArgument struct {
	name       string
	value      string
	scalarName string
}

func (f *FieldCommandCustomScalarArgument) String() string {
	// Return custom scalar as is, without any changes
	return f.value
}

func (f *FieldCommandCustomScalarArgument) Value() interface{} {
	return &f.value
}

func (f *FieldCommandCustomScalarArgument) Type() string {
	return f.scalarName
}

func (f *FieldCommandCustomScalarArgument) Name() string {
	return f.name
}

func FieldCommandArgName(f FieldCommandArgument) string {
	return fmt.Sprintf("arg-%s", f.Name())
}

func FieldCommandArgType(f FieldCommandArgument) string {
	return fmt.Sprintf("Argument of type %s", f.Type())
}

func Set(cmd *cobra.Command, f FieldCommandArgument) {
	argName := FieldCommandArgName(f)
	usage := FieldCommandArgType(f)
	switch fvt := f.Value().(type) {
	case Setter:
		fvt.Set(cmd, f)
	case pflag.Value:
		cmd.Flags().Var(fvt, argName, usage)
	case *int64:
		cmd.Flags().Int64Var(fvt, argName, 0, usage)
	case *float64:
		cmd.Flags().Float64Var(fvt, argName, 0.0, usage)
	case *string:
		cmd.Flags().StringVar(fvt, argName, "", usage)
	case *bool:
		cmd.Flags().BoolVar(fvt, argName, false, usage)
	}
}

func getFieldCommandArgumentForArg(arg introspection.Arg) FieldCommandArgument {
	switch {
	case arg.Type.Scalar():
		switch arg.Type.Name {
		case "Int":
			return &FieldCommandIntArgument{
				name: arg.Name,
			}
		case "Float":
			return &FieldCommandFloatArgument{
				name: arg.Name,
			}
		case "String":
			return &FieldCommandStringArgument{
				name: arg.Name,
			}
		case "Boolean":
			return &FieldCommandBooleanArgument{
				name: arg.Name,
			}
		case "ID":
			return &FieldCommandIDArgument{
				name: arg.Name,
			}
		default:
			return &FieldCommandCustomScalarArgument{
				name:       arg.Name,
				scalarName: arg.Type.Name,
			}
		}
	case arg.Type.Enum():
		return &FieldCommandEnumArgument{
			name:     arg.Name,
			enumName: arg.Type.Name,
		}
	case arg.Type.Input():
		return &FieldCommandInputArgument{
			name:      arg.Name,
			inputName: arg.Type.Name,
		}
	case arg.Type.NonNull():
		nArg := arg
		// TODO: better error message for panic here
		nArg.Type = *nArg.Type.OfType
		return &fieldCommandNonNullArgument{
			FieldCommandArgument: getFieldCommandArgumentForArg(nArg),
		}
	case arg.Type.List():
		nArg := arg
		// TODO: better error message for panic here
		nArg.Type = *nArg.Type.OfType
		return &FieldCommandListArgument{
			FieldCommandArgument: getFieldCommandArgumentForArg(nArg),
		}
	default:
		panic(fmt.Sprintf("malformed schema, %s cannot be used as input", arg.Type.Kind))
	}
}

func GetFieldArguments(field introspection.Field) []FieldCommandArgument {
	args := make([]FieldCommandArgument, 0, len(field.Args))
	for _, arg := range field.Args {
		switch {
		case arg.Type.NonNull():
			// unpack root non null
			nArg := arg
			nArg.Type = *arg.Type.OfType
			args = append(args, &FieldCommandNonNullArgument{
				FieldCommandArgument: getFieldCommandArgumentForArg(nArg),
			})
		default:
			args = append(args, getFieldCommandArgumentForArg(arg))
		}
	}
	return args
}
