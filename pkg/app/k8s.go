package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	util "github.com/cheriot/kubenav/internal/util"
	"github.com/cheriot/kubenav/pkg/app/relations"

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/describe"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// KubeContextList names of contexts from kubeconfig
func KubeContextList() ([]string, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := rules.Load()
	if err != nil {
		return nil, fmt.Errorf("Unable to rules.Load(): %w", err)
	}

	return util.Keys(config.Contexts), nil
}

type KubeCluster struct {
	name             string
	restClientConfig *restclient.Config
	apiResources     []metav1.APIResource
	scheme           *runtime.Scheme // Could be global since it's go types?
	dynamicClient    dynamic.Interface
}

func NewKubeClusterDefault(ctx context.Context) (*KubeCluster, error) {
	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		panic(err)
	}
	return NewKubeCluster(ctx, config.CurrentContext)
}

func NewKubeCluster(ctx context.Context, kubeCtxName string) (*KubeCluster, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: kubeCtxName})
	restClientConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating restclient: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restClientConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamicClient: %w", err)
	}

	apiResource, err := fetchAllApiResources(restClientConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting api-resources: %w", err)
	}

	// This could be global. It's not context/cluster specific
	scheme := runtime.NewScheme()
	err = schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return nil, fmt.Errorf("NewKubeCluster failed to build scheme: %w", err)
	}

	return &KubeCluster{
		name:             kubeCtxName,
		restClientConfig: restClientConfig,
		apiResources:     apiResource,
		scheme:           scheme,
		dynamicClient:    dynamicClient,
	}, nil
}

func (kc *KubeCluster) KubeNamespaceList(ctx context.Context) ([]string, error) {
	coreclient, err := corev1client.NewForConfig(kc.restClientConfig)
	if err != nil {
		return []string{}, fmt.Errorf("unable to create coreclient for %s: %w", kc.name, err)
	}

	nsList, err := coreclient.Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []string{}, fmt.Errorf("unable to list namespaces for %s: %w", kc.name, err)
	}

	return util.Map(nsList.Items, func(ns corev1.Namespace) string {
		return ns.Name
	}), nil
}

// Commands:
// ctx
// ns
// ctx <name>
// ns <name> (current ctx)
// po (current ctx ns)
// po <name>
// <name> (current ctx ns kind)
// thisiserr

// The view needs:
// responseType {Namespace, Ctx, List, Object, Error}
// ctx
// ns
// kind
// name
// err
type CommandResultType string

const (
	CRTContext = "ctx"
	CRTQuery   = "query"
	CRTObject  = "obj"
	CRTError   = "err"
)

type CommandResult struct {
	CommandResultType `json:"commandResultType"`
	Namespace         string `json:"ns"`
	Kind              string `json:"kind"`
	Query             string `json:"query"`
	Name              string `json:"name"`
	ErrorMsg          string `json:"error"`
}

func ErrorCommandResult(errorMsg string) CommandResult {
	return CommandResult{
		CommandResultType: CRTError,
		ErrorMsg:          errorMsg,
	}
}

func (kc *KubeCluster) Command(ctx context.Context, ns string, query string, cmd string) CommandResult {
	result := CommandResult{
		Namespace: ns,
		Query:     query,
	}

	fields := strings.Fields(strings.TrimSpace(cmd))
	if len(fields) == 0 {
		return ErrorCommandResult("empty command")
	}

	action := fields[0]
	if action == "ctx" || action == "context" {
		// ctx (context selection page)
		// ctx somename (change context)
		result.CommandResultType = CRTContext
		if len(fields) > 1 {
			// TODO validate this is a valid context name
			result.Name = fields[1]
		}
		return result
	}

	actionMatches := findAPIResources(kc.apiResources, action)
	if len(actionMatches) == 0 {
		return ErrorCommandResult(fmt.Sprintf("unkown command or resource '%s' in '%s'", action, cmd))
	}

	apiResource := actionMatches[0]
	result.Kind = apiResource.Kind

	if len(fields) > 1 {
		// ns somename
		result.CommandResultType = CRTObject
		result.Name = fields[1]
		return result
	}

	// po
	// ns
	result.CommandResultType = CRTQuery
	result.Query = action
	return result
}

type ResourceTable struct {
	APIResource   metav1.APIResource `json:"apiResource"`
	Table         *metav1.Table      `json:"table"`
	IsError       bool               `json:"isError"`
	TableRowNames []string           `json:"tableRowNames"`
}

func (kc *KubeCluster) Query(ctx context.Context, nsName string, query string) ([]ResourceTable, error) {
	log.Infof("Query for %s", query)
	matches := findAPIResources(kc.apiResources, query)
	log.Infof("matches %v", util.Map(matches, func(ar metav1.APIResource) string { return ar.Kind }))

	results := util.Map(matches, func(r metav1.APIResource) ResourceTable {
		table, err := kc.listResource(ctx, r, nsName)
		if err != nil {
			log.Errorf("listResource error for resource %+v: %w", r, err)
			table = PrintError(err)
		}

		// Get a list of metadata.name for the object represented by each row. Ideally this would come from
		// Table.Rows[]Object but I'm not sure how to specify the includeObject policy or decode the RawExtension
		// instance.
		nameIdx := -1
		for i, cd := range table.ColumnDefinitions {
			if strings.ToLower(cd.Name) == "name" && cd.Type == "string" {
				nameIdx = i
			}
		}
		rowNames := make([]string, len(table.Rows))
		for i, row := range table.Rows {
			if nameIdx > -1 {
				rowNames[i] = row.Cells[nameIdx].(string)
			} else {
				rowNames[i] = ""
			}
		}

		return ResourceTable{
			APIResource:   r,
			Table:         table,
			IsError:       err != nil,
			TableRowNames: rowNames,
		}
	})

	// Maintain order of the results, but move empty tables to the end
	nonEmpty, empty := util.Partition(results, func(r ResourceTable) bool {
		return len(r.Table.Rows) > 0
	})
	orderedResults := append(nonEmpty, empty...)

	return orderedResults, nil
}

func findAPIResources(apiResources []metav1.APIResource, identifier string) []metav1.APIResource {
	isMatch := func(r metav1.APIResource) bool {
		names := []string{
			strings.ToLower(r.Name),
			strings.ToLower(r.Kind),
			strings.ToLower(r.Group),
			strings.ToLower(r.SingularName),
		}
		names = append(names, r.Categories...)
		names = append(names, r.ShortNames...)
		return util.Contains(names, strings.ToLower(identifier))
	}

	return util.Filter(apiResources, isMatch)
}

type KubeObject struct {
	Relations []relations.HasOneDestination `json:"relations"`
	Describe  string                        `json:"describe"`
	Yaml      string                        `json:"yaml"`
	Errors    []error                       `json:"errors"`
}

func (kc *KubeCluster) GetResource(ctx context.Context, nsName string, kind string, resourceName string) (*KubeObject, error) {
	errors := make([]error, 0)
	matches := findAPIResources(kc.apiResources, kind)
	if len(matches) == 0 {
		return nil, fmt.Errorf("unable to find an api resource: %s", kind)
	}

	apiResource := matches[0]
	if len(matches) > 1 {
		errors = append(errors, fmt.Errorf("found more APIResource matches than expected, %d, for GetKubeObject %+v", len(matches), matches))
	}

	unstructured, err := kc.getResource(ctx, apiResource, nsName, resourceName)
	if err != nil {
		return nil, fmt.Errorf("unable to GetKubeObject: %w", err)
	}

	yamlStr, err := renderYaml(unstructured)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to serialize yaml: %w", err))
	}

	describeStr, err := kc.Describe(ctx, nsName, kind, resourceName)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to describe: %w", err))
	}

	rs := make([]relations.HasOneDestination, 0)
	if kc.scheme.IsGroupRegistered(apiResource.Group) {
		gvk := toGVK(apiResource)
		obj, err := kc.scheme.New(gvk)
		if err != nil {
			errors = append(errors, fmt.Errorf("unable to instantiate %s when yamling %s %s %s", gvk, nsName, kind, resourceName))
		} else {

			err = kc.scheme.Convert(unstructured, obj, nil)
			if err != nil {
				errors = append(errors, fmt.Errorf("unable to convert: %w", err))
			}
			rs = relations.RelationsList(obj, toGK(apiResource))
		}
	}

	return &KubeObject{
		Relations: rs,
		Yaml:      yamlStr,
		Describe:  describeStr,
		Errors:    errors,
	}, nil
}

func (kc *KubeCluster) Describe(ctx context.Context, nsName string, kind string, resourceName string) (string, error) {
	matches := findAPIResources(kc.apiResources, kind)

	for _, apiResource := range matches {

		describer, found := describe.DescriberFor(toGK(apiResource), kc.restClientConfig)
		if !found {
			var restMapper *meta.DefaultRESTMapper
			if kc.scheme.IsGroupRegistered(apiResource.Group) {
				restMapper = meta.NewDefaultRESTMapper(kc.scheme.PrioritizedVersionsAllGroups())
			} else {
				restMapper = meta.NewDefaultRESTMapper([]schema.GroupVersion{toGV(apiResource)})
				scope := meta.RESTScopeNamespace
				if !apiResource.Namespaced {
					scope = meta.RESTScopeRoot
				}

				restMapper.Add(toGVK(apiResource), scope)
			}

			restMapping, err := restMapper.RESTMapping(toGK(apiResource))
			if err != nil {
				return "", fmt.Errorf("no RESTMapping for %v: %w", toGK(apiResource), err)
			}

			describer, found = describe.GenericDescriberFor(restMapping, kc.restClientConfig)
			if !found {
				return "", fmt.Errorf("unable to create a GenericDescriberFor %v", apiResource)
			}
		}

		return describer.Describe(nsName, resourceName, describe.DescriberSettings{ShowEvents: true, ChunkSize: 5})
	}

	return "", fmt.Errorf("no resources found matching %s", kind)
}

func renderYaml(unstructured *unstructured.Unstructured) (string, error) {
	// None of this managedFields nonsense
	if untyped, ok := unstructured.Object["metadata"]; ok {
		if md, ok := untyped.(map[string]interface{}); ok {
			delete(md, "managedFields")
		}
	}

	bs, err := yaml.Marshal(&unstructured.Object)
	if err != nil {
		return "", fmt.Errorf("unable to marshal unstructured: %w", err)
	}

	return string(bs), nil
}

func (kc *KubeCluster) Yaml(ctx context.Context, nsName string, kind string, resourceName string) (string, error) {
	matches := findAPIResources(kc.apiResources, kind)

	for _, apiResource := range matches {
		gvk := toGVK(apiResource)

		unst, err := kc.getResource(ctx, apiResource, nsName, resourceName)
		if err != nil {
			return "", fmt.Errorf("unable to getResource %v %s %s", gvk, nsName, resourceName)
		}

		return renderYaml(unst)
	}

	return "", fmt.Errorf("no resources found matching %s", kind)
}

func ApiResources(kubeconfigOverride string) ([]metav1.APIResource, error) {
	config, err := readConfig(kubeconfigOverride)
	if err != nil {
		return nil, err
	}
	return fetchAllApiResources(config)
}

func Describe(ns string, kind string, name string) (string, error) {
	ctxs, err := KubeContextList()
	if err != nil {
		return "", err
	}

	kc, err := GetOrMakeKubeCluster(context.TODO(), ctxs[0])
	if err != nil {
		return "", err
	}

	return kc.Describe(context.TODO(), ns, kind, name)
}

func fetchAllApiResources(restClientConfig *restclient.Config) ([]metav1.APIResource, error) {
	// Should this use runtime.Scheme or RESTMapper???
	// https://iximiuz.com/en/posts/kubernetes-api-structure-and-terminology/
	// https://iximiuz.com/en/posts/kubernetes-api-go-types-and-common-machinery/
	log.Infof("fetchAllApiResources **Expensive**")

	client, err := discovery.NewDiscoveryClientForConfig(restClientConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create new discovery client for config: %w", err)
	}
	groups, resourceLists, err := client.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("unable to get server groups and resources: %w", err)
	}
	// APIVersion == group/version

	// ignore deprecated GroupVersions for now
	notPreferred := make(map[string]bool)
	for _, g := range groups {
		if len(g.Versions) > 1 {
			for _, v := range g.Versions {
				if v != g.PreferredVersion {
					notPreferred[v.GroupVersion] = true
				}
			}
		}
	}

	var apiResources []metav1.APIResource
	for _, rls := range resourceLists {
		if !notPreferred[rls.GroupVersion] {
			for _, r := range rls.APIResources {
				group, version, err := splitGroupVersion(rls.GroupVersion)
				if err != nil {
					log.Errorf("error splitting GroupVersion on %+v: %w", rls, err)
					continue
				}
				r.Group = group
				r.Version = version
				if !isSubresource(r) {
					apiResources = append(apiResources, r)
				}
			}
		}
	}

	return apiResources, nil
}

func splitGroupVersion(groupVersion string) (string, string, error) {
	parts := strings.Split(groupVersion, "/")
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	if len(parts) == 1 {
		// core resources like Pod are just v1 with no group.
		return "", groupVersion, nil
	}
	return "", "", fmt.Errorf("unexpected GroupVersion format %s", groupVersion)
}

func readConfig(kubeconfigOverride string) (*restclient.Config, error) {
	kubeconfig := kubeconfigOverride
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func isSubresource(r metav1.APIResource) bool {
	return strings.Contains(r.Name, "/")
}

const LIST_LIMIT = 1000

func (kc *KubeCluster) listResource(ctx context.Context, r metav1.APIResource, namespace string) (*metav1.Table, error) {
	var uList *unstructured.UnstructuredList
	var err error
	if r.Namespaced {
		uList, err = kc.dynamicClient.Resource(toGVR(r)).Namespace(namespace).List(ctx, metav1.ListOptions{Limit: LIST_LIMIT})
	} else {
		uList, err = kc.dynamicClient.Resource(toGVR(r)).List(ctx, metav1.ListOptions{Limit: LIST_LIMIT})
	}
	if err != nil {
		return nil, fmt.Errorf("dynamicClient list failed for %+v: %w", r, err)
	}

	return PrintList(kc.scheme, r, uList)
}

func (kc *KubeCluster) getResource(ctx context.Context, r metav1.APIResource, namespace string, name string) (*unstructured.Unstructured, error) {
	namespacable := kc.dynamicClient.Resource(toGVR(r))

	var ri dynamic.ResourceInterface
	if r.Namespaced {
		if namespace == "" {
			return nil, fmt.Errorf("namespaced resource, but an empty namespace name: %s '%s'", toGVR(r), namespace)
		}
		ri = namespacable.Namespace(namespace)
	} else {
		ri = namespacable
	}

	return ri.Get(ctx, name, metav1.GetOptions{})
}

func toGVR(r metav1.APIResource) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    r.Group,
		Version:  r.Version,
		Resource: r.Name,
	}
}

func toGVK(r metav1.APIResource) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.Group,
		Version: r.Version,
		Kind:    r.Kind,
	}
}

func toGV(r metav1.APIResource) schema.GroupVersion {
	return schema.GroupVersion{
		Group:   r.Group,
		Version: r.Version,
	}
}

func toGK(r metav1.APIResource) schema.GroupKind {
	return schema.GroupKind{
		Group: r.Group,
		Kind:  r.Kind,
	}
}
