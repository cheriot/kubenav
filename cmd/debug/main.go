package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/cheriot/kubenav/pkg/app"
	"github.com/cheriot/kubenav/pkg/app/relations"

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

	kc, err := app.NewKubeClusterDefault(context.Background())
	resourceTables, err := kc.Query(context.Background(), c.Namespace, c.PositionalArgs.Kind)

	err = RenderResourceTables(resourceTables)
	if err != nil {
		panic(fmt.Sprintf("Unable to complete get command: %s", err.Error()))
	}
	return nil
}

type RelationsCommand struct{}

func (c *RelationsCommand) Execute(_ []string) error {
	pod := &corev1.Pod{}
	pod.Spec.NodeName = "fakeNodeName"
	// ns := &corev1.Namespace{}
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	fmt.Printf("%+v\n", pod.GetObjectKind().GroupVersionKind())

	hors := relations.RelationsList(pod, schema.GroupKind{Group: "", Kind: "Pod"})

	fmt.Printf("Found %d relations", len(hors))
	for _, hor := range hors {
		fmt.Printf("%+v", hor)
	}

	return nil
}

type ApiResourcesCommand struct{}

func (c *ApiResourcesCommand) Execute(_ []string) error {
	fmt.Printf("Execute ApiResourcesCommand\n")
	resources, err := app.ApiResources(globalOptions.KubeConfig)
	if err != nil {
		panic(fmt.Sprintf("Unable to execute api-resources command: %s", err.Error()))
	}
	err = RenderApiResources(resources)
	if err != nil {
		panic(fmt.Sprintf("Unable to complete api-resources command: %s", err.Error()))
	}
	return nil
}

type DescribeCommand struct {
	Namespace      string                 `long:"namespace" short:"n" required:"true" description:"Namespace scope for queries"`
	PositionalArgs DescribePositionalArgs `positional-args:"true"`
}

type DescribePositionalArgs struct {
	Kind string `positional-arg-name:"kind" required:"true" description:"name, shortName, or category of resource(s) to query"`
	Name string `positional-arg-name:"name" required:"true" description:"name of the instance to describe"`
}

func (c *DescribeCommand) Execute(_ []string) error {
	fmt.Printf("Execute DescribeCommand\n")
	output, err := app.Describe(c.Namespace, c.PositionalArgs.Kind, c.PositionalArgs.Name)
	if err != nil {
		fmt.Printf("error describing %s %s %s %+v: %v", c.Namespace, c.PositionalArgs.Kind, c.PositionalArgs.Name, c.PositionalArgs, err)
	}
	fmt.Println(output)
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

	describeDesc := "Describe the state of an object."
	_, err = parser.AddCommand("describe", describeDesc, describeDesc, &DescribeCommand{})
	if err != nil {
		return nil, err
	}

	relDesc := "Relations of an object."
	_, err = parser.AddCommand("relations", relDesc, relDesc, &RelationsCommand{})
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
