package relations

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cheriot/kubenav/internal/util"

	// rbacv1 "k8s.io/api/rbac/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Optimize queries with fieldSelector
// kubectl get pods --all-namespaces -o wide --field-selector spec.nodeName=<node>
// https://stackoverflow.com/questions/39231880/kubernetes-api-get-pods-on-specific-nodes

// hasOne (origin has a name):
//   * pod has one node
//   * [cluster]rolebinding roleRef
//   * ingress ingressClass
// hasList (origin has a list of names):
//   * owner references
//   * [cluster]rolebinding subjects
// hasSearch (label selectors)
//   * service matches pods
//   * netpols
//   * reverse of hasOne and hasList
// transitive relations
//   * deployment to pods
//   * ingress to pods

// Operations
// 1. current object -> list of {GroupKind, (search params | object [namespace] name)}
// 2. search page query params -> list of resources

var relations = BuildRelations()

func BuildRelations() []HasReferenceRelation {
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	rbacv1.AddToScheme(scheme)

	podGK := objectKind(&corev1.Pod{}, scheme)
	nodeGK := objectKind(&corev1.Node{}, scheme)
	rsGK := objectKind(&appsv1.ReplicaSet{}, scheme)
	// deploymentGK := objectKind(&appsv1.Deployment{}, scheme)
	rGK := objectKind(&rbacv1.Role{}, scheme)
	rbGK := objectKind(&rbacv1.RoleBinding{}, scheme)
	crbGK := objectKind(&rbacv1.ClusterRoleBinding{}, scheme)
	crGK := objectKind(&rbacv1.ClusterRole{}, scheme)

	var podHasOneNode = HasReferenceRelation{
		Origin: podGK,
		Destinations: func(origin runtime.Object) ([]RelationDestination, bool) {
			pod := origin.(*corev1.Pod)
			if pod.Spec.NodeName == "" {
				return nil, false
			}
			return []RelationDestination{{
				GroupKind: nodeGK,
				Name:      pod.Spec.NodeName,
			}}, true
		},
	}

	var podHasOwners = HasReferenceRelation{
		// https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
		// ownerReferences:
		//   - apiVersion: apps/v1
		//     blockOwnerDeletion: true
		//     controller: true
		//     kind: ReplicaSet
		//     name: frontend
		//     uid: f391f6db-bb9b-4c09-ae74-6a1f77f3d5cf
		Origin: podGK,
		Destinations: func(origin runtime.Object) ([]RelationDestination, bool) {
			pod, isPod := origin.(*corev1.Pod)
			if !isPod {
				return nil, false
			}
			rds := ownerDestinations(pod.GetNamespace(), pod.GetOwnerReferences())
			return rds, len(rds) != 0
		},
	}

	var rsHasOwners = HasReferenceRelation{
		Origin: rsGK,
		Destinations: func(origin runtime.Object) ([]RelationDestination, bool) {
			rs, isRs := origin.(*appsv1.ReplicaSet)
			if !isRs {
				return nil, false
			}
			rds := ownerDestinations(rs.GetNamespace(), rs.GetOwnerReferences())
			return rds, len(rds) != 0
		},
	}

	var roleBindingHasOneRole = HasReferenceRelation{
		Origin: rbGK,
		Destinations: func(origin runtime.Object) ([]RelationDestination, bool) {
			rb := origin.(*rbacv1.RoleBinding)
			if rb.RoleRef.Kind == "" || rb.RoleRef.Name == "" {
				return nil, false
			}

			// Can reference either a Role in the current namespace or a ClusterRole
			if rb.RoleRef.Kind == rGK.Kind {
				return []RelationDestination{{
					GroupKind: rGK,
					Namespace: rb.Namespace,
					Name:      rb.RoleRef.Name,
				}}, true
			}

			return []RelationDestination{{
				GroupKind: crGK,
				Namespace: "", // cluster scope
				Name:      rb.RoleRef.Name,
			}}, true
		},
	}

	var roleBindingHasSubjects = HasReferenceRelation{
		Origin: rbGK,
		Destinations: func(origin runtime.Object) ([]RelationDestination, bool) {
			rb := origin.(*rbacv1.RoleBinding)

			rds := subjectDestinations(rb.Subjects)
			return rds, len(rds) > 0
		},
	}

	var clusterRoleBindingHasOneRole = HasReferenceRelation{
		Origin: crbGK,
		Destinations: func(origin runtime.Object) ([]RelationDestination, bool) {
			crb := origin.(*rbacv1.ClusterRoleBinding)
			if crb.RoleRef.Kind == "" || crb.RoleRef.Name == "" {
				return nil, false
			}

			return []RelationDestination{{
				GroupKind: crGK,
				Namespace: "", // cluster scope
				Name:      crb.RoleRef.Name,
			}}, true
		},
	}

	var clusterRoleBindingHasSubjects = HasReferenceRelation{
		Origin: crbGK,
		Destinations: func(origin runtime.Object) ([]RelationDestination, bool) {
			rb := origin.(*rbacv1.ClusterRoleBinding)

			rds := subjectDestinations(rb.Subjects)
			return rds, len(rds) > 0
		},
	}

	return []HasReferenceRelation{
		podHasOneNode,
		podHasOwners,
		rsHasOwners,
		roleBindingHasOneRole,
		roleBindingHasSubjects,
		clusterRoleBindingHasOneRole,
		clusterRoleBindingHasSubjects,
	}
}

func subjectDestinations(subjects []rbacv1.Subject) []RelationDestination {
	rds := make([]RelationDestination, 0)
	for _, s := range subjects {
		rds = append(rds, RelationDestination{
			GroupKind: schema.GroupKind{Group: s.APIGroup, Kind: s.Kind},
			Namespace: s.Namespace,
			Name:      s.Name,
		})
	}

	return rds
}

func objectKind(obj runtime.Object, scheme *runtime.Scheme) schema.GroupKind {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil || len(gvks) > 1 {
		panic(fmt.Sprintf("unexpected ObjectKinds: %+v for %+v", gvks, obj))
	}
	return gvks[0].GroupKind()
}

func RelationsList(origin runtime.Object, originGK schema.GroupKind) []RelationDestination {
	destinations := make([]RelationDestination, 0)
	for _, hor := range relations {
		if hor.Origin == originGK {
			rds, ok := hor.Destinations(origin)
			if ok {
				destinations = append(destinations, rds...)
			}
		}
	}

	return destinations
}

func ReverseHasOneRelation(destination runtime.Object, ns string, sr HasReferenceRelation, possibleOrigins []runtime.Object) []runtime.Object {
	d := RelationDestination{
		GroupKind: objGK(destination),
		Namespace: ns,
	}

	matches := make([]runtime.Object, 0)
	for _, possibleOrigin := range possibleOrigins {
		rds, ok := sr.Destinations(possibleOrigin)
		util.Filter(rds, func(rd RelationDestination) bool {
			// TODO will rd have an accurate Namespace?
			return rd.GroupKind == d.GroupKind && rd.Name == d.Name
		})
		if ok {
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
type HasReferenceRelation struct {
	Origin schema.GroupKind `json:"origin"`
	// given the origin object, does it have this relation? Ex does a Service have a .spec.selector. Does a pod have an ownerReference?
	// IsApplicable        func(runtime.Object) bool                `json:"-"`
	// IdentifyDestination func(runtime.Object) RelationDestination `json:"-"`
	Destinations func(runtime.Object) ([]RelationDestination, bool) `json:"-"`
}

type RelationDestination struct {
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

func ownerDestinations(originNamespace string, ownerReferences []metav1.OwnerReference) []RelationDestination {
	rrs := make([]RelationDestination, 0)
	for _, or := range ownerReferences {
		group := strings.Split(or.APIVersion, "/")
		if len(group) > 0 {
			rrs = append(rrs, RelationDestination{
				GroupKind: schema.GroupKind{Group: group[0], Kind: or.Kind},
				Namespace: originNamespace, // TODO Empty for cluster scope objects
				Name:      or.Name,
			})
		}
	}
	return rrs
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
