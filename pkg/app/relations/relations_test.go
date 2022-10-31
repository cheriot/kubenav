package relations_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cheriot/kubenav/pkg/app/relations"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Relations", func() {

	Expect("foo", "bar")
	Context("ReplicaSet with ownerRef", func() {
		rs := &appsv1.ReplicaSet{}
		rs.GetObjectMeta().SetOwnerReferences([]metav1.OwnerReference{{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "hello-server",
			UID:        "456996f4-f01e-4aee-b214-6c7984d786fc",
		}})
		// scheme := runtime.NewScheme()
		// corev1.AddToScheme(scheme)
		// pod := &corev1.Pod{}

		It("has a relation", func() {
			rds := relations.RelationsList(rs, schema.GroupKind{Group: "apps", Kind: "ReplicaSet"})
			Expect(len(rds)).To(Equal(1))

			rd := rds[0]
			Expect(rd.Kind).To(Equal("Deployment"))
			Expect(rd.Name).To(Equal("hello-server"))
		})
	})

	Context("Pod with ownerRef", func() {
		pod := &corev1.Pod{}
		pod.GetObjectMeta().SetOwnerReferences([]metav1.OwnerReference{{
			APIVersion: "apps/v1",
			Kind:       "ReplicaSet",
			Name:       "hello-server-rs",
			UID:        "456996f4-f01e-4aee-b214-6c7984d786fc",
		}})

		It("has a relation", func() {
			rds := relations.RelationsList(pod, schema.GroupKind{Group: "", Kind: "Pod"})
			Expect(len(rds)).To(Equal(1))

			rd := rds[0]
			Expect(rd.Kind).To(Equal("ReplicaSet"))
			Expect(rd.Name).To(Equal("hello-server-rs"))
		})
	})

	Context("Pod with node", func() {
		pod := &corev1.Pod{}
		pod.Spec.NodeName = "node-1-asdf"

		It("has a relation", func() {
			rds := relations.RelationsList(pod, schema.GroupKind{Group: "", Kind: "Pod"})
			Expect(len(rds)).To(Equal(1))

			rd := rds[0]
			Expect(rd.Kind).To(Equal("Node"))
			Expect(rd.Name).To(Equal(pod.Spec.NodeName))
		})
	})

})
