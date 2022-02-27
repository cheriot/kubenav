package app

import (
	"context"
	"sync"
)

var kubeClustersLock = sync.RWMutex{}
var kubeClusters = make(map[string]*KubeCluster)

func GetOrMakeKubeCluster(ctx context.Context, kubeCtxName string) (*KubeCluster, error) {
	kubeClustersLock.Lock()
	defer kubeClustersLock.Unlock()

	kc, found := kubeClusters[kubeCtxName]
	if !found {
		var err error
		kc, err = NewKubeCluster(ctx, kubeCtxName)
		if err != nil {
			return nil, err
		}
		kubeClusters[kc.name] = kc
	}

	return kc, nil
}
