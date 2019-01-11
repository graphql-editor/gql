# GraphQL CLI

## About

Without going too much into details, I was lately working on some integration of one of our GraphQL projects and had to write a simple internal tooling script for some simple tasks.

Now, I love curl as much as a next guy, it is my first `goto` tool when I need to do some debugging with anything related to REST/GraphQL/HTTP(S) in general but stil, writing
```sh
$ curl -X POST -g -H "Authorization: bearer ${GITHUB_TOKEN}" -H "Content-Type: application/json" https://api.github.com/graphql -d '{"query": "query {viewer {issues(first: 1) {nodes{title}}}}"}
```
is a bit of a mouthful. Especially those brackets, I actually made a typo with backets while writing this example.

So my first though was, well I'll just write a simple CLI for that project, but with GraphQL and it's introspection, you can quite easily make an "one shoe fits all" kind of tool. And that's what this project tries to be.

Now, this project is very much WIP. Any and all contributions are welcome as long as a contributor remembers one thing, the aim of this tool is not to be a robust and flexible GraphQL client that creats specialized and optimized GraphQL queries. Nor is it a tool meant to help with writing/deployment/whatnot of GraphQL schema for project. It's goal is to make a simple and convenient CLI for most schemas that already exist somewhere upstream. It's meant to be used mostly while scripting and debugging. So usage>performance.

It's quite small (~3000 lines) and comes with (almost finished) completion to boot.

![sample](https://user-images.githubusercontent.com/11337563/50778798-1f8dd680-129f-11e9-89a7-7ac9805ca584.gif)

## Requirements

### Build

* Go version 1.11 or higher

## Installation

### Latest release

```sh
$ sudo curl https://raw.githubusercontent.com/slothking-online/gql/master/getgql | sudo sh
```

### From source

```
go get -u github.com/slothking-online/gql
```

### Without root

```sh
$ curl https://raw.githubusercontent.com/slothking-online/gql/master/getgql | PREFIX=$HOME/bin sh
```

## License

MIT

## How It Works

The main idea is that there's "main" path to the field along which the tool is build the query. That is a bit against GraphQL ideology, as it limits the user's flexibility. Making it harder/more costly to get more complex data, but again, this tool is meant to be mostly developer tool for stuff like simple scripting,debugging and experimenting. That's why I decided it's better to just stick to simplicity rather than robustness and flexibility.

Most of functionality is self documenting (if possible pulling docs from schema). If there were no errors, unrelated to GraphQL, the data field of response will be written as JSON to stdout, while errors field to stderr.

For example
```sh
$ gql query --endpoint https://countries.trevorblades.com/ --help
```

will return

```
<snip>

Usage:
  gql query [flags]
  gql query [command]

Available Commands:
  continent   continent(code: String): Continent
  continents  continents: Continent
  countries   countries: Country
  country     country(code: String): Country
  language    language(code: String): Language
  languages   languages: Language

Flags:
      --endpoint string      graphql endpoint
      --fields stringArray   additional fields to resolve aside from the next one in resolve path
      --format string        go template response formatting
      --header Header        set header to be passed in a http request, can be set multiple times (default {})
  -h, --help                 help for query
      --max-depth int        resolve this field up to max-depth
      --no-cache             do not cache schema introspection result

Use "gql query [command] --help" for more information about a command.
```

So if we want to pull currency in country, for example in united states, we would do something like this:

```
$ gql query --endpoint https://countries.trevorblades.com/ country --arg-code US currency
```

which would give us
```
{
    "country": {
        "currency": "USD,USN,USS"
    }
}
```

This is not exactly useful in scriptting, as parsing json in bash is not exactly fun (unless you're using `jq`, but that's another dependency to install). We can refine it using Go Templates

```
gql query --endpoint https://countries.trevorblades.com/ country --arg-code US currency --format '{{.country.currency}}'
```

and then we get
```
USD,USN,USS
```

## Docs

WIP

## Contribute

1.  Fork this repo
2.  Create your feature branch: git checkout -b feature-name
3.  Commit your changes: git commit -am 'Add some feature'
4.  Push to the branch: git push origin my-new-feature
5.  Submit a pull request

## TODO

* Tests for all of commands
* `--fields` option on each node in path to include additional fields in response
* Missing GraphQL features such as depracation and subscription (using websockets)
* Some serious refactoring on cmd package with goal of embedding the command in custom tools.
* Documentation for subscription and mutation
* Cache improvements, right now it's just a simple serialized json
* ~~Completion improvements (right now completion, won't event hint query,mutation and subscription until endpoint path is provided)~~
* ~~ZSH completion~~
* Profiles, typing --endpoint with url each time is annoying
