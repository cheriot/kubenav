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

func RenderResourceTables(resourceTables []app.ResourceTable) error {
	for _, rt := range resourceTables {
		fmt.Printf("%s/%s.%s\n", rt.APIResource.Group, rt.APIResource.Version, rt.APIResource.Kind)

		colNames := util.Map(rt.Table.ColumnDefinitions, func(cd v1.TableColumnDefinition) string {
			return cd.Name
		})
		fmt.Println(strings.Join(colNames, "\t"))

		for _, tr := range rt.Table.Rows {
			cells := util.Map(tr.Cells, func(cell interface{}) string {
				switch c := cell.(type) {
				case string:
					return c
				default:
					log.Errorf("unknown type from %+v", cell)
					return "<unknowntype>"
				}
			})
			fmt.Println(strings.Join(cells, "\t"))
		}
	}

	return nil
}
