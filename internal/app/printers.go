package app

import (
	"fmt"
	"time"

	printers "github.com/cheriot/kubenav/internal/copyofk8sprinters"
	"github.com/cheriot/kubenav/internal/copyofk8sprinters/internalversion"
	//k8sprinters "github.com/cheriot/kubenav/internal/copyofk8sprinters/internalversion"
	util "github.com/cheriot/kubenav/internal/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"

	// AddToScheme
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apiserverinternalv1alpha1 "k8s.io/api/apiserverinternal/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	authenticationv1 "k8s.io/api/authentication/v1"
	authenticationv1beta1 "k8s.io/api/authentication/v1beta1"
	authorizationv1 "k8s.io/api/authorization/v1"
	authorizationv1beta1 "k8s.io/api/authorization/v1beta1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	certificatesv1 "k8s.io/api/certificates/v1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	coordinationv1 "k8s.io/api/coordination/v1"
	coordinationv1beta1 "k8s.io/api/coordination/v1beta1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	eventsv1 "k8s.io/api/events/v1"
	eventsv1beta1 "k8s.io/api/events/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	flowcontrolv1alpha1 "k8s.io/api/flowcontrol/v1alpha1"
	flowcontrolv1beta1 "k8s.io/api/flowcontrol/v1beta1"
	flowcontrolv1beta2 "k8s.io/api/flowcontrol/v1beta2"
	imagepolicy "k8s.io/api/imagepolicy/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	nodev1 "k8s.io/api/node/v1"
	nodev1alpha1 "k8s.io/api/node/v1alpha1"
	nodev1beta1 "k8s.io/api/node/v1beta1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1alpha1 "k8s.io/api/rbac/v1alpha1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	schedulingv1alpha1 "k8s.io/api/scheduling/v1alpha1"
	schedulingv1beta1 "k8s.io/api/scheduling/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	storagev1alpha1 "k8s.io/api/storage/v1alpha1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
)

var schemeBuilder = runtime.SchemeBuilder{
	appsv1beta1.AddToScheme,
	autoscalingv2beta1.AddToScheme,
	batchv1.AddToScheme,
	batchv1beta1.AddToScheme,
	certificatesv1beta1.AddToScheme,
	coordinationv1.AddToScheme,
	rbacv1beta1.AddToScheme,
	admissionv1.AddToScheme,
	admissionv1beta1.AddToScheme,
	admissionregistrationv1.AddToScheme,
	admissionregistrationv1beta1.AddToScheme,
	apiserverinternalv1alpha1.AddToScheme,
	appsv1.AddToScheme,
	appsv1beta1.AddToScheme,
	appsv1beta2.AddToScheme,
	authenticationv1.AddToScheme,
	authenticationv1beta1.AddToScheme,
	authorizationv1.AddToScheme,
	authorizationv1beta1.AddToScheme,
	autoscalingv2.AddToScheme,
	autoscalingv2beta2.AddToScheme,
	autoscalingv2beta1.AddToScheme,
	autoscalingv1.AddToScheme,
	batchv1.AddToScheme,
	batchv1beta1.AddToScheme,
	certificatesv1.AddToScheme,
	certificatesv1beta1.AddToScheme,
	coordinationv1.AddToScheme,
	coordinationv1beta1.AddToScheme,
	corev1.AddToScheme,
	discoveryv1beta1.AddToScheme,
	discoveryv1.AddToScheme,
	eventsv1.AddToScheme,
	eventsv1beta1.AddToScheme,
	extensionsv1beta1.AddToScheme,
	flowcontrolv1beta2.AddToScheme,
	flowcontrolv1beta1.AddToScheme,
	flowcontrolv1alpha1.AddToScheme,
	imagepolicy.AddToScheme,
	networkingv1.AddToScheme,
	networkingv1beta1.AddToScheme,
	nodev1.AddToScheme,
	nodev1beta1.AddToScheme,
	nodev1alpha1.AddToScheme,
	policyv1.AddToScheme,
	policyv1beta1.AddToScheme,
	rbacv1.AddToScheme,
	rbacv1beta1.AddToScheme,
	rbacv1alpha1.AddToScheme,
	schedulingv1.AddToScheme,
	schedulingv1beta1.AddToScheme,
	schedulingv1alpha1.AddToScheme,
	storagev1.AddToScheme,
	storagev1beta1.AddToScheme,
	storagev1alpha1.AddToScheme,
}

func PrintList(scheme *runtime.Scheme, ar metav1.APIResource, uList *unstructured.UnstructuredList) (*metav1.Table, error) {
	isRegistered := scheme.IsVersionRegistered(toGV(ar))
	if isRegistered {
		table, err := printRegistered(scheme, ar, uList)
		if err != nil {
			return nil, fmt.Errorf("printRegistered error: %w", err)
		}
		return table, nil
	}

	// fallback to [name, age] if no printer
	return printUnstructured(uList)
}

func PrintError(err error) *metav1.Table {
	return &metav1.Table{
		ColumnDefinitions: []metav1.TableColumnDefinition{{Name: "Error"}},
		Rows: []metav1.TableRow{{Cells: []interface{}{err.Error()}}},
	}
}

func printRegistered(scheme *runtime.Scheme, ar metav1.APIResource, uList *unstructured.UnstructuredList) (*metav1.Table, error) {
	gvk := uList.GetObjectKind().GroupVersionKind()

	typedInstance, err := scheme.New(gvk)
	if err != nil {
		return nil, fmt.Errorf("failed New %+v: %w", gvk, err)
	}

	// Populate typedInstance with the data from uList
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(uList.UnstructuredContent(), typedInstance); err != nil {
		return nil, fmt.Errorf("unable to convert unstructured object to %v: %w", gvk, err)
	}

	tableGenerator := printers.NewTableGenerator().With(internalversion.AddHandlers)
	table, err := tableGenerator.GenerateTable(typedInstance, printers.GenerateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to GenerateTable for %v: %w", gvk, err)
	}

	table.Rows = util.Map(table.Rows, func(r metav1.TableRow) metav1.TableRow {
		// This contains the entire Pod. Don't send it down to the client.
		r.Object = runtime.RawExtension{}
		return r
	})

	return table, nil
}

func printUnstructured(uList *unstructured.UnstructuredList) (*metav1.Table, error) {
	columns := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	}

	rows := util.Map(uList.Items, func(item unstructured.Unstructured) metav1.TableRow {
		return metav1.TableRow{
			Cells: []interface{}{
				item.GetName(),
				translateTimestampSince(item.GetCreationTimestamp()),
			},
		}
	})

	return &metav1.Table{
		ColumnDefinitions: columns,
		Rows: rows,
	}, nil
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}