package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	util "github.com/cheriot/kubenav/internal/util"

	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	//"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func KubeContextList() ([]string, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := rules.Load()
	if err != nil {
		return nil, fmt.Errorf("Unable to rules.Load(): %w", err)
	}

	return util.Keys(config.Contexts), nil
	//	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
	//		clientcmd.NewDefaultClientConfigLoadingRules(),
	//		&clientcmd.ConfigOverrides{},
	//	)
}

func KubeNamespaceList(ctx context.Context, kubeCtxName string) ([]string, error) {

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: kubeCtxName})
	restClientConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to build a restclient.Config: %w", err)
	}

	coreclient, err := corev1client.NewForConfig(restClientConfig)
	nsList, err := coreclient.Namespaces().List(ctx, metav1.ListOptions{})
	return util.Map(nsList.Items, func(ns corev1.Namespace) string { return ns.Name }), nil
}

func ApiResources(kubeconfigOverride string) ([]metav1.APIResource, error) {

	config, err := readConfig(kubeconfigOverride)
	if err != nil {
		return nil, err
	}
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	groups, resourceLists, err := client.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}

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
	fmt.Printf("notPreferred %+v\n\n", notPreferred)

	var apiResources []metav1.APIResource
	for _, rls := range resourceLists {
		if !notPreferred[rls.GroupVersion] {
			for _, r := range rls.APIResources {
				group, version, err := splitGroupVersion(rls.GroupVersion)
				if err != nil {
					// TODO: log error
					fmt.Printf("skipping %+v\n\n", r)
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

func GetResource(kubeconfigOverride string, inputName string, namespace string) (*K8sSearchResponse, error) {
	config, err := readConfig(kubeconfigOverride)
	if err != nil {
		return nil, err
	}

	// TODO pass in the config
	resources, err := ApiResources(kubeconfigOverride)
	if err != nil {
		return nil, err
	}
	matches := make([]metav1.APIResource, 0)
	for _, r := range resources {
		if hasAnyLower(inputName, r.Categories, r.ShortNames, []string{r.Name, r.Kind}) {
			matches = append(matches, r)
		}
	}

	fmt.Printf("now go find %v\n\n", matches)

	iface, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	response := K8sSearchResponse{
		Query:       inputName,
		Namespace:   namespace,
		KindResults: make([]K8sSearchKindResult, 0, len(matches)),
	}
	for _, r := range matches {
		unstructured, err := iface.Resource(toGVR(r)).Namespace(namespace).List(context.TODO(), metav1.ListOptions{Limit: 1})
		if err != nil {
			return nil, err
		}

		result := K8sSearchKindResult{
			Resource: r,
			Results:  unstructured,
		}

		response.KindResults = append(response.KindResults, result)
	}

	return &response, nil
}

type K8sSearchResponse struct {
	Query       string
	Namespace   string
	KindResults []K8sSearchKindResult
}

type K8sSearchKindResult struct {
	Resource metav1.APIResource
	// UnstructuredList.Object map[apiVersion:v1 kind:PodList metadata:map[resourceVersion:279690]]
	Results *unstructured.UnstructuredList
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
