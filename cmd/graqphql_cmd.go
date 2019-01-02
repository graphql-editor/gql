package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/graphql-go/graphql"

	"github.com/slothking-online/gql/client"
	"github.com/slothking-online/gql/introspection"
	"github.com/spf13/cobra"
)

var (
	CacheTimeout time.Duration = -10 * time.Minute
)

type GraphQLCommandType struct {
	Description string
	Fields      map[string]*FieldCommand
}

type GraphQLTypeCommandMap map[string]GraphQLCommandType

type FieldCommand struct {
	*cobra.Command
	Field    introspection.Field
	Schema   introspection.Schema
	Args     []FieldCommandArgument
	MaxDepth int
	Fields   []string
}

func shortDesc(field introspection.Field, args []FieldCommandArgument) string {
	docString := ": " + field.Type.GoString()
	if len(args) != 0 {
		argsStrings := make([]string, 0, len(args))
		for _, arg := range args {
			argsStrings = append(argsStrings, fmt.Sprintf("%s: %s", arg.Name(), arg.Type()))
		}
		docString = "(" + strings.Join(argsStrings, ", ") + ")" + docString
	}
	return field.Name + docString
}

func NewFieldCommand(
	field introspection.Field,
	schema introspection.Schema,
	RunE func(*cobra.Command, []string) error,
	PreRun func(*cobra.Command, []string),
) *FieldCommand {
	args := GetFieldArguments(field)
	fc := &FieldCommand{
		Field:  field,
		Schema: schema,
		Args:   args,
		Command: &cobra.Command{
			Use:              field.Name,
			Long:             field.Description,
			Short:            shortDesc(field, args),
			RunE:             RunE,
			PersistentPreRun: PreRun,
		},
	}
	for _, arg := range fc.Args {
		Set(fc.Command, arg)
	}
	fc.Command.Flags().IntVar(&fc.MaxDepth, "max-depth", 0, "resolve this field up to max-depth")
	fc.Command.Flags().StringArrayVar(&fc.Fields, "fields", nil, "additional fields to resolve aside from the next one in resolve path")
	return fc
}

func (f *FieldCommand) ArgsString() string {
	if len(f.Args) == 0 {
		return ""
	}
	buf := &bytes.Buffer{}
	sep := "("
	for _, arg := range f.Args {
		// if not changed assume not set
		flag := f.Command.Flags().Lookup(FieldCommandArgName(arg))
		if !flag.Changed {
			continue
		}
		fmt.Fprintf(buf, "%s%s: %s", sep, arg.Name(), arg.String())
		sep = ", "
	}
	if buf.Len() != 0 {
		fmt.Fprint(buf, ")")
	}
	return buf.String()
}

func (f *FieldCommand) solve(sf introspection.Field, depth int, withArgs bool) string {
	sf.Type = sf.Type.GetOfTypeLeaf()
	realType := sf.Type
	if realType.TypeRef() {
		realType = realType.Deref(f.Schema.Types)
	}
	for _, rt := range f.Schema.Types {
		if sf.Type.Name == rt.Name {
			realType = rt
			break
		}
	}
	fields := make([]string, 0, len(realType.Fields))
	for _, field := range realType.Fields {
		// Ignore all fields that have atleast
		// one non null argument
		for _, a := range field.Args {
			if a.Type.NonNull() {
				continue
			}
		}
		field.Type = field.Type.GetOfTypeLeaf()
		switch {
		case field.Type.Enum(), field.Type.Scalar():
			fields = append(fields, field.Name)
		default:
			if depth > 0 {
				solved := f.solve(field, depth-1, false)
				if solved != "" {
					fields = append(fields, solved)
				}
			}
		}
	}
	if len(fields) == 0 {
		return ""
	}
	if withArgs {
		return fmt.Sprintf("%s%s { %s }", sf.Name, f.ArgsString(), strings.Join(fields, " "))
	}
	return fmt.Sprintf("%s { %s }", sf.Name, strings.Join(fields, " "))
}

func (f *FieldCommand) BuildQuery() string {
	if f.Field.Type.Enum() || f.Field.Type.Scalar() {
		return f.Field.Name
	}
	var solved string
	// Keep increasing depth until we get atleast some kind
	// of query
	nd := f.MaxDepth
	for solved == "" {
		solved = f.solve(f.Field, nd, true)
		nd++
	}
	return solved
}

type GraphQLCommandConfig struct {
	// requied: name of the field this command resolves
	Field introspection.Field
	// required: query builder
	QueryBuilder *QueryBuilder
	// optional: path requested from this field
	Path []string
	// required: schema executed by command
	Schema introspection.Schema
}

type QueryBuilder struct {
	query     string
	variables map[string]interface{}
}

type GraphQLCommand struct {
	Config       GraphQLCommandConfig
	FieldCommand *FieldCommand
	QueryBuilder *QueryBuilder
}

func (qb *QueryBuilder) Wrap(name string, extra ...string) {
	if qb.query == "" {
		qb.query = name
	} else {
		qb.query = fmt.Sprintf("%s { %s %s }", name, strings.Join(extra, " "), qb.query)
	}
}

func (qb *QueryBuilder) Query() string {
	return qb.query
}

func (qb *QueryBuilder) Set(name string, value interface{}) {

}

func (qb *QueryBuilder) Variables() map[string]interface{} {
	return qb.variables
}

// appends field query parent to a query
func (g *GraphQLCommand) FieldPreRun(c *cobra.Command, args []string) {
	// FieldCommand cannot be nil here
	// that's why, rather than checking,
	// panic outright as this is a programming
	// error
	if g.FieldCommand == nil {
		panic("FieldCommand cannot be nil in FieldPreRun")
	}
	if g.FieldCommand.Command == c {
		// Leaf field
		g.QueryBuilder.Wrap(g.FieldCommand.BuildQuery())
	} else {
		g.QueryBuilder.Wrap(g.FieldCommand.Field.Name + g.FieldCommand.ArgsString())
	}

	// Traverse parent preruns, to build full query.
	var parentPreRun func(*cobra.Command, []string)
	for p := g.FieldCommand.Command.Parent(); p != nil; p = p.Parent() {
		if p.PersistentPreRun != nil {
			parentPreRun = p.PersistentPreRun
			break
		}
	}
	if parentPreRun != nil {
		parentPreRun(c, args)
	}
}

// run command
func (g *GraphQLCommand) RunE(c *cobra.Command, args []string) error {
	if err := g.checkValid(); err != nil {
		return err
	}
	// just run parent command, as the real implementation
	// of run is defined by root query
	return g.FieldCommand.Command.Parent().RunE(c, args)
}

func (g *GraphQLCommand) isSimple() bool {
	t := g.Config.Field.Type
	return t.Scalar() || t.Enum()
}

func (g *GraphQLCommand) checkValid() error {
	if g.isSimple() && len(g.Config.Path) != 0 {
		return errors.New("enum and scalar leaf fields do not accept any more arguments")
	}
	return nil
}

func (g *GraphQLCommand) BuildSubCommands() error {
	var thisPath string
	var path []string

	if len(g.Config.Path) > 0 {
		thisPath = g.Config.Path[0]
		if len(g.Config.Path) > 1 {
			path = g.Config.Path[1:]
		}
	}
	var t *introspection.Type
	for _, tt := range g.Config.Schema.Types {
		if tt.Name == g.Config.Field.Type.Name {
			t = &tt
			break
		}
	}
	if t == nil {
		return fmt.Errorf("type %s not found", g.Config.Field.Type.Name)
	}
	if g.isSimple() {
		return nil
	}
	for _, field := range t.Fields {
		config := g.Config
		config.Field = field
		config.Path = path
		cmd := NewGraphQLCommand(config)
		g.FieldCommand.AddCommand(cmd.FieldCommand.Command)
		if thisPath == config.Field.Name {
			if err := (&cmd).BuildSubCommands(); err != nil {
				return err
			}
		}
	}
	return nil
}

type GraphQLRootCommands struct {
	Config       GraphQLRootConfig
	Query        GraphQLCommand
	Mutation     GraphQLCommand
	Subscription GraphQLCommand
	Schema       introspection.Schema
	QueryBuilder *QueryBuilder
}

func NewGraphQLCommand(config GraphQLCommandConfig) GraphQLCommand {
	config.Field.Type = config.Field.Type.GetOfTypeLeaf()
	cmd := GraphQLCommand{
		Config:       config,
		QueryBuilder: config.QueryBuilder,
	}
	cmd.FieldCommand = NewFieldCommand(
		config.Field,
		config.Schema,
		cmd.RunE,
		cmd.FieldPreRun,
	)
	return cmd
}

func NewGraphQLRootCommands(config GraphQLRootConfig) (GraphQLRootCommands, error) {
	if config.QueryBuilder == nil {
		config.QueryBuilder = &QueryBuilder{}
	}
	gqlRoot := GraphQLRootCommands{
		Config:       config,
		QueryBuilder: config.QueryBuilder,
	}
	var err error
	if config.Schema != nil && !config.ForceRemote {
		err = gqlRoot.newCommandFromSchema()
	}
	if config.Schema == nil || config.ForceRemote || err != nil {
		err = gqlRoot.newCommandFromRemote()
		if err != nil {
			return GraphQLRootCommands{}, err
		}
	}
	return gqlRoot, nil
}

// create command from user supplied schema
func (g *GraphQLRootCommands) newCommandFromSchema() error {
	return nil
}

// replace all non alphanumeric characters with "-"
func endpointCacheFn(cfg GraphQLRootConfig) string {
	return regexp.MustCompile("[^a-zA-Z0-9]").ReplaceAllString(cfg.Endpoint, "-")
}

// get default system cache path
func (g *GraphQLRootCommands) cacheFilePath() (string, error) {
	cDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cDir = filepath.Join(cDir, "gql")
	if _, err := os.Stat(cDir); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(cDir, os.ModeDir|os.FileMode(0740))
		}
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(cDir, endpointCacheFn(g.Config)), nil
}

// check if there's cached introspection
// that's not older than CacheTimeout
func (g *GraphQLRootCommands) loadFromCache() (s introspection.Schema, ok bool) {
	// TODO: replace json cache with something
	// more binary, maybe github.com/davecgh/go-xdr?
	// or maybe even go as far as keeping whole cache and
	// whole introspection set in something like SQLite?
	if noCache {
		return
	}
	cfn, err := g.cacheFilePath()
	if err != nil {
		return
	}
	st, err := os.Stat(cfn)
	if err != nil {
		return
	}
	if st.ModTime().Before(time.Now().Add(CacheTimeout)) {
		return
	}
	b, err := ioutil.ReadFile(cfn)
	if err != nil {
		return
	}
	if err := json.Unmarshal(b, &s); err == nil {
		ok = true
	}
	return
}

// cache introspection for reuse
func (g *GraphQLRootCommands) saveCache(schema introspection.Schema) {
	// TODO: log some kind of warning on errors?
	cfn, err := g.cacheFilePath()
	if err != nil {
		return
	}
	b, err := json.Marshal(schema)
	if err != nil {
		return
	}
	ioutil.WriteFile(cfn, b, os.FileMode(0740))
}

// do introspection on remote endpoint and pull schema from
// upstream
func (g *GraphQLRootCommands) newCommandFromRemote() error {
	if schema, ok := g.loadFromCache(); ok {
		return g.newCommandFromIntrospection(schema)
	}
	cli := client.New(client.Config{
		Endpoint: g.Config.Endpoint,
	})
	httpHeader := make(http.Header)
	for k, v := range g.Config.Header {
		httpHeader.Add(k, v)
	}
	schema, err := introspection.GetSchemaTypes(cli, httpHeader)
	if err != nil {
		return err
	}
	err = g.newCommandFromIntrospection(schema)
	if err == nil {
		g.saveCache(schema)
	}
	return err
}

// all commands on graphql start here
type GraphQLRootConfig struct {
	// required: remote graphql endpoint
	Endpoint string
	// any additional headers that should
	// be attached to the request
	Header Header
	// ignore local schema data and use upstream
	// endpoint, optional
	ForceRemote bool
	// optional: query builder used by command
	QueryBuilder *QueryBuilder
	// required: path being resolved by a command
	Path []string
	// optional: local schema
	Schema *graphql.Schema
}

func (g *GraphQLRootCommands) rootCmd(
	schema introspection.Schema,
	field introspection.Field,
	path []string,
) GraphQLCommand {
	cmd := NewGraphQLCommand(GraphQLCommandConfig{
		Field:        field,
		Path:         path,
		QueryBuilder: g.Config.QueryBuilder,
		Schema:       schema,
	})
	cmd.FieldCommand.Command.RunE = g.RunE
	return cmd
}

// analyze schema introspection result and build command from it
func (g *GraphQLRootCommands) newCommandFromIntrospection(schema introspection.Schema) error {
	var thisPath string
	var path []string
	if len(g.Config.Path) > 0 {
		thisPath = g.Config.Path[0]
		if len(g.Config.Path) > 1 {
			path = g.Config.Path[1:]
		}
	}
	if schema.QueryType.Name != "" {
		g.Query = g.rootCmd(
			schema,
			introspection.Field{
				Type: schema.QueryType,
				Name: "query",
				Description: `Simplified query with completion useful for scripting and command line administration.

Executes a query operation on graphql endpoint returning a value of a field if it is a scalar or an enum. If the field returns an object, interface or union type then all scalar and enum fields of that type are returned. If a field returns a list type, the same rules apply to each value in a list.

If field takes arguments user can set them by setting an option AFTER the field that accepts an argument but BEFORE the next field in resolve path.

By default, _ONLY_ top level scalar and enum fields of an object are returned. This is done so that client can be kept relativly simple while dealing with recursive queries. This can be overriden by setting max-depth value to a positive integer value, in which case client will resolve non-scalar fields up to defined depth. Depth is relative to the leaf value of requested path. Be careful since each field is resolved with it's own set of args, so --max-depth option on a field higher in resolve path will not apply to the following fields.
`,
			},
			path,
		)
		g.Query.FieldCommand.Short = "Quick graphql query operation"
	}

	if schema.MutationType.Name != "" {
		g.Mutation = g.rootCmd(
			schema,
			introspection.Field{
				Type:        schema.MutationType,
				Name:        "mutation",
				Description: `TODO`,
			},
			path,
		)
		g.Mutation.FieldCommand.Short = "Quick graphql mutation operation"
	}

	if schema.SubscriptionType.Name != "" {
		g.Subscription = g.rootCmd(
			schema,
			introspection.Field{
				Type:        schema.SubscriptionType,
				Name:        "subscription",
				Description: `TODO`,
			},
			path,
		)
		g.Subscription.FieldCommand.Short = "Quick graphql subscription operation"
	}
	// Handle only known thisPath cases, ignoring
	// anyhing else.
	// Other values might not necesserily
	// be an error, so just silently ignore them.
	switch thisPath {
	case "query":
		g.Query.BuildSubCommands()
	case "mutation":
		g.Mutation.BuildSubCommands()
	case "subscription":
		g.Subscription.BuildSubCommands()
	}
	g.Schema = schema
	return nil
}

// common run function for all GraphQL commands
func (g *GraphQLRootCommands) RunE(c *cobra.Command, args []string) error {
	httpHeader := make(http.Header)
	for k, v := range g.Config.Header {
		httpHeader.Add(k, v)
	}
	cli := client.New(client.Config{
		Endpoint: g.Config.Endpoint,
	})
	r := client.Raw{
		Query:     g.QueryBuilder.Query(),
		Variables: g.QueryBuilder.Variables(),
		Header:    httpHeader,
	}
	execute(cli, r, nil)
	return nil
}
