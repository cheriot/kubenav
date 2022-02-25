package main

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RenderApiResources(resources []v1.APIResource) error {
	for _, r := range resources {
		// fmt.Printf("%s %s %s %s %s\n\n", r.Kind, r.ShortNames, r.Categories, r.Group, r.Version)
		fmt.Printf("%+v\n\n", r)
	}

	return nil
}

func RenderGetResource(response *K8sSearchResponse) error {
	fmt.Printf("get view %v\n", response)
	for _, kr := range response.KindResults {
		fmt.Printf("%s.%s\n", kr.Resource.Version, kr.Resource.Kind)
		for _, item := range kr.Results.Items {
			b, err := json.MarshalIndent(item, "", "  ")
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(string(b))
			//for k := range item.Object {
			//	fmt.Printf("%+v\n\n", k)
			//}
		}
	}
	return nil
}
