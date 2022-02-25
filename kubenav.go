package main

import (
	"fmt"

	flags "github.com/jessevdk/go-flags"
)

var globalOptions ApplicationOptions

type GenCommandPositionalArgs struct {
	Kind string `positional-arg-name:"kind" required:"true" description:"name, shortName, or category of resource(s) to query"`
}

type GetCommand struct {
	Namespace      string                   `long:"namespace" short:"n" required:"true" description:"Namespace scope for queries"`
	PositionalArgs GenCommandPositionalArgs `positional-args:"true"`
}

func (c *GetCommand) Execute(args []string) error {
	fmt.Printf("Execute GetCommand %+v %+v %+v\n", globalOptions, c, args)
	stuff, err := GetResource(globalOptions.KubeConfig, c.PositionalArgs.Kind, c.Namespace)
	if err != nil {
		panic(fmt.Sprintf("Unable to execute get command: %s", err.Error()))
	}
	err = RenderGetResource(stuff)
	if err != nil {
		panic(fmt.Sprintf("Unable to complete get command: %s", err.Error()))
	}
	return nil
}

type ApiResourcesCommand struct{}

func (c *ApiResourcesCommand) Execute(_ []string) error {
	fmt.Printf("Execute ApiResourcesCommand\n")
	resources, err := ApiResources(globalOptions.KubeConfig)
	if err != nil {
		panic(fmt.Sprintf("Unable to execute api-resources command: %s", err.Error()))
	}
	err = RenderApiResources(resources)
	if err != nil {
		panic(fmt.Sprintf("Unable to complete api-resources command: %s", err.Error()))
	}
	return nil
}

type ApplicationOptions struct {
	Verbose    int    `long:"verbose" short:"v" description:"Debug level [0,4]"`
	KubeConfig string `long:"kubeconfig" description:"Absolute path to the kubeconfig file"`
}

func BuildParser(appOptions *ApplicationOptions) (*flags.Parser, error) {
	parser := flags.NewParser(appOptions, flags.Default)

	getDesc := "Kind, shortName, or category of resource(s)"
	cmd, err := parser.AddCommand("get", getDesc, getDesc, &GetCommand{})
	if err != nil {
		return nil, err
	}
	cmd.ArgsRequired = true

	apiResourcesDesc := "List all API resource kinds available in the cluster."
	_, err = parser.AddCommand("api-resources", apiResourcesDesc, apiResourcesDesc, &ApiResourcesCommand{})
	if err != nil {
		return nil, err
	}

	parser.CommandHandler = func(commander flags.Commander, args []string) error {
		fmt.Printf("Set log level here. %+v\n", globalOptions)
		return commander.Execute(args)
	}

	return parser, nil
}

func main() {
	globalOptions = ApplicationOptions{}
	cliParser, err := BuildParser(&globalOptions)
	if err != nil {
		panic(err)
	}

	// Let go-flags print errors.
	cliParser.Parse()
}
