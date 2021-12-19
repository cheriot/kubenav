package main

import (
	"fmt"
	"path/filepath"
	"strings"

	//"k8s.io/client-go/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func ApiResources(kubeconfigOverride string) ([]v1.APIResource, error) {

	config, err := readConfig(kubeconfigOverride)
	if err != nil {
		return nil, err
	}

	client := discovery.NewDiscoveryClientForConfigOrDie(config)
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

	apiResources := make([]v1.APIResource, 0, 64)
	for _, rls := range resourceLists {
		if !notPreferred[rls.GroupVersion] {
			for _, r := range rls.APIResources {
				group, version, err := splitGroupVersion(rls.GroupVersion)
				if err != nil {
					// TODO: log error
					continue
				}
				r.Group = group
				r.Version = version
				if !isSubresource(r) {
					apiResources = append(apiResources, r)
					fmt.Printf("%s %s %s %s %s\n\n", r.Kind, r.ShortNames, r.Categories, r.Group, r.Version)
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

func isSubresource(r v1.APIResource) bool {
	return strings.Contains(r.Name, "/")
}
