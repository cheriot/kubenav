package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	util "github.com/cheriot/kubenav/internal/util"

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
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
	scheme           *runtime.Scheme // Could be global
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

	apiResource, err := apiResources(restClientConfig)
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

type ResourceTable struct {
	APIResource metav1.APIResource `json:"apiResource"`
	Table *metav1.Table `json:"table"`
	IsError bool `json:"isError"`
}

func (kc *KubeCluster) Query(ctx context.Context, nsName string, query string) ([]ResourceTable, error) {
	log.Infof("Query for %s", query)
	isMatch := func(r metav1.APIResource) bool {
		names := []string{
			strings.ToLower(r.Name),
			strings.ToLower(r.Kind),
			strings.ToLower(r.Group),
			strings.ToLower(r.SingularName),
		}
		names = append(names, r.Categories...)
		return util.Contains(names, strings.ToLower(query))
	}

	matches := util.Filter(kc.apiResources, isMatch)
	log.Infof("matches %v", util.Map(matches, func(ar metav1.APIResource) string { return ar.Kind }))

	results := util.Map(matches, func(r metav1.APIResource) ResourceTable {
		table, err := kc.ListResource(r, nsName)
		if err != nil {
			log.Errorf("ListResource error for resource %+v: %w", r, err)
			table = PrintError(err)
		}
		return ResourceTable{
			APIResource: r,
			Table:       table,
		}
	})

	nonempty := util.Filter(results, func(rt ResourceTable) bool {
		return len(rt.Table.Rows) > 0
	})

	return nonempty, nil
}

func ApiResources(kubeconfigOverride string) ([]metav1.APIResource, error) {
	config, err := readConfig(kubeconfigOverride)
	if err != nil {
		return nil, err
	}
	return apiResources(config)
}

func apiResources(restClientConfig *restclient.Config) ([]metav1.APIResource, error) {
	// Should this use runtime.Scheme or RESTMapper???
	// https://iximiuz.com/en/posts/kubernetes-api-structure-and-terminology/
	// https://iximiuz.com/en/posts/kubernetes-api-go-types-and-common-machinery/

	client, err := discovery.NewDiscoveryClientForConfig(restClientConfig)
	if err != nil {
		return nil, err
	}
	groups, resourceLists, err := client.ServerGroupsAndResources()
	if err != nil {
		return nil, err
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

	// for _, r := range apiResources {
	// 	fmt.Printf("%+v\n", r)
	// }
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
	return "", "", fmt.Errorf("Unexpected GroupVersion format %s", groupVersion)
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

func (kc *KubeCluster) ListResource(r metav1.APIResource, namespace string) (*metav1.Table, error) {
	var uList *unstructured.UnstructuredList
	var err error
	if r.Namespaced {
		uList, err = kc.dynamicClient.Resource(toGVR(r)).Namespace(namespace).List(context.TODO(), metav1.ListOptions{Limit: 1})
	} else {
		uList, err = kc.dynamicClient.Resource(toGVR(r)).List(context.TODO(), metav1.ListOptions{Limit: 1})
	}
	if err != nil {
		return nil, fmt.Errorf("dynamicClient list failed for %+v: %w", r, err)
	}

	return PrintList(kc.scheme, r, uList)
}

func hasAnyLower(t0 string, args ...[]string) bool {
	target := strings.ToLower(t0)
	for _, ts := range args {
		for _, t := range ts {
			if strings.ToLower(t) == target {
				return true
			}
		}
	}
	return false
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
