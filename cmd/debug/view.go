package main

import (
	"fmt"
	"strings"

	"github.com/cheriot/kubenav/internal/app"
	"github.com/cheriot/kubenav/internal/util"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RenderApiResources(resources []v1.APIResource) error {
	for _, r := range resources {
		// fmt.Printf("%s %s %s %s %s\n\n", r.Kind, r.ShortNames, r.Categories, r.Group, r.Version)
		fmt.Printf("%+v\n\n", r)
	}

	return nil
}

func RenderGetResource(response *app.K8sSearchResponse) error {
	columns := util.Map(response.Table.ColumnDefinitions, func(cd v1.TableColumnDefinition) string {
		return cd.Name
	})
	fmt.Println(strings.Join(columns, "\t"))

	for _, tr := range response.Table.Rows {
		cells := util.Map(tr.Cells, func(cell interface{}) string {
			switch c := cell.(type) {
			case string:
				return c
			}
			log.Errorf("unknown type from %+v", cell)
			return "<unknowntype>"
		})
		fmt.Println(strings.Join(cells, "\t"))
	}

	// for _, kr := range response.KindResults {
	// 	fmt.Printf("%s.%s\n", kr.Resource.Version, kr.Resource.Kind)
	// 	for _, item := range kr.Results.Items {
	// 		b, err := json.MarshalIndent(item, "", "  ")
	// 		if err != nil {
	// 			fmt.Println(err.Error())
	// 		}

	// 		fmt.Printf(string(b)) // Make Printf
	// 	}
	// }
	return nil
}
