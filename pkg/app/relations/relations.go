package relations

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// rbacv1 "k8s.io/api/rbac/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Optimize queries with fieldSelector
// kubectl get pods --all-namespaces -o wide --field-selector spec.nodeName=<node>
// https://stackoverflow.com/questions/39231880/kubernetes-api-get-pods-on-specific-nodes

// * transitive relations
//   * deployment to pods
//   * ingress to pods

// Operations
// 1. current object -> list of {GroupKind, (search params | object [namespace] name)}
// 2. search page query params -> list of resources
type Relatable interface {
	runtime.Object
}

var relations = BuildRelations()

func BuildRelations() []HasOneRelation {
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)

	podGK := objectKind(&corev1.Pod{}, scheme)
	nodeGK := objectKind(&corev1.Node{}, scheme)
	rsGK := objectKind(&appsv1.ReplicaSet{}, scheme)
	deploymentGK := objectKind(&appsv1.Deployment{}, scheme)

	var podHasOneNode = HasOneRelation{
		Origin:      podGK,
		Destination: nodeGK,
		IsApplicable: func(origin runtime.Object) bool {
			pod := origin.(*corev1.Pod)
			return pod.Spec.NodeName != ""
		},
		IdentifyDestination: func(origin runtime.Object) HasOneDestination {
			pod := origin.(*corev1.Pod)
			return HasOneDestination{
				GroupKind: nodeGK,
				Name:      pod.Spec.NodeName,
			}
		},
	}

	var podHasOneReplicaSet = HasOneRelation{
		// https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
		// ownerReferences:
		//   - apiVersion: apps/v1
		//     blockOwnerDeletion: true
		//     controller: true
		//     kind: ReplicaSet
		//     name: frontend
		//     uid: f391f6db-bb9b-4c09-ae74-6a1f77f3d5cf
		Origin:      podGK,
		Destination: rsGK,
		IsApplicable: func(origin runtime.Object) bool {
			pod := origin.(*corev1.Pod)

			return nil != ownerByGK(pod.OwnerReferences, rsGK)
		},
		IdentifyDestination: func(origin runtime.Object) HasOneDestination {
			pod := origin.(*corev1.Pod)

			or := ownerByGK(pod.OwnerReferences, rsGK)
			return HasOneDestination{
				GroupKind: rsGK,
				Namespace: pod.Namespace,
				Name:      or.Name,
			}
		},
	}

	var rsHasOneDeployment = HasOneRelation{
		Origin:      rsGK,
		Destination: deploymentGK,
		IsApplicable: func(origin runtime.Object) bool {
			rs := origin.(*appsv1.ReplicaSet)

			return nil != ownerByGK(rs.OwnerReferences, deploymentGK)
		},
		IdentifyDestination: func(origin runtime.Object) HasOneDestination {
			rs := origin.(*appsv1.ReplicaSet)
			or := ownerByGK(rs.OwnerReferences, deploymentGK)

			return HasOneDestination{
				GroupKind: deploymentGK,
				Namespace: rs.Namespace,
				Name:      or.Name,
			}
		},
	}

	return []HasOneRelation{podHasOneNode, podHasOneReplicaSet, rsHasOneDeployment}
}

func objectKind(obj runtime.Object, scheme *runtime.Scheme) schema.GroupKind {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil || len(gvks) > 1 {
		panic(fmt.Sprintf("unexpected ObjectKinds: %+v for %+v", gvks, obj))
	}
	return gvks[0].GroupKind()
}

func RelationsList(origin Relatable, originGK schema.GroupKind) []HasOneDestination {
	destinations := make([]HasOneDestination, 0)
	for _, hor := range relations {
		if hor.Origin == originGK && hor.IsApplicable(origin) {
			destinations = append(destinations, hor.IdentifyDestination(origin))
		}
	}

	return destinations
}

func ReverseHasOneRelation(destination runtime.Object, ns string, sr HasOneRelation, possibleOrigins []runtime.Object) []runtime.Object {
	d := HasOneDestination{
		GroupKind: objGK(destination),
		Namespace: ns,
	}

	matches := make([]runtime.Object, 0)
	for _, possibleOrigin := range possibleOrigins {
		if sr.IsApplicable(possibleOrigin) && d == sr.IdentifyDestination(possibleOrigin) {
			matches = append(matches, possibleOrigin)
		}
	}
	return matches
}

// * *:1 relations
//   * ownerRef
//   * ingress to backend
//   * pod to node (spec.nodeName)
//   * [cluster] role binding -> service account
//   * [cluster] role binding -> role

// The Origin object  contains an identifier for the one and only Destination object it has this relationship with.
type HasOneRelation struct {
	Origin      schema.GroupKind `json:"origin"`
	Destination schema.GroupKind `json:"destination"`
	// given the origin object, does it have this relation? Ex does a Service have a .spec.selector. Does a pod have an ownerReference?
	IsApplicable        func(runtime.Object) bool              `json:"-"`
	IdentifyDestination func(runtime.Object) HasOneDestination `json:"-"`
}

type HasOneDestination struct {
	schema.GroupKind `json:"groupKind"`
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
}

// * 1:* direct relations
//   * pod to service
//     * via selector
//     * without selector (manually managed endpoint(s))
//     * ExternalName with cluster local names
//   * netpol to pods
//   * pods to netpol
//   * [cluster] role -> role bindings

// Complex one to many relationships where the results will be viewed as an advanced search page prepopulated with criteria.
type HasManyRelations struct {
	Origin      schema.GroupKind
	Destination schema.GroupKind
	// Does this relationship exist? ie Show this relationship on Origin's page?
	// ie a Service without selector has no known relationship to pods
	IsApplicable func(origin runtime.Object) bool
	// On the search page for Destination, use these query params
	QueryParams func(origin runtime.Object) map[string]string
}

var podNode = HasOneRelation{
	Origin:      schema.GroupKind{Group: "", Kind: "Pod"},
	Destination: schema.GroupKind{Group: "", Kind: "Node"},
	IsApplicable: func(origin runtime.Object) bool {
		pod := origin.(*corev1.Pod)
		return pod.Spec.NodeName != ""
	},
	IdentifyDestination: func(origin runtime.Object) HasOneDestination {
		pod := origin.(*corev1.Pod)
		return HasOneDestination{
			GroupKind: objGK(&corev1.Node{}),
			Name:      pod.Spec.NodeName,
		}
	},
}

func ownerByGK(ownerReferences []metav1.OwnerReference, gk schema.GroupKind) *metav1.OwnerReference {
	for _, or := range ownerReferences {
		group := strings.Split(or.APIVersion, "/")
		if len(group) > 0 && group[0] == gk.Group && or.Kind == gk.Kind {
			return &or
		}
	}
	return nil
}

func objGK(obj runtime.Object) schema.GroupKind {
	return obj.GetObjectKind().GroupVersionKind().GroupKind()
}

type RelationTyped[T runtime.Object, U runtime.Object] struct {
	Origin        T
	Destination   U
	IsApplicable  func(T) bool
	ExtractParams func(T) map[string]string
}
