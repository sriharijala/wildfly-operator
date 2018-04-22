package stub

import (
	"reflect"
	"github.com/banzaicloud/wildfly-operator/pkg/apis/wildfly/v1alpha1"
	"github.com/coreos/operator-sdk/pkg/sdk/action"
	"github.com/coreos/operator-sdk/pkg/sdk/handler"
	"github.com/coreos/operator-sdk/pkg/sdk/types"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	"k8s.io/apimachinery/pkg/api/resource"
	"fmt"
	"github.com/coreos/operator-sdk/pkg/sdk/query"
	"k8s.io/apimachinery/pkg/labels"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func NewHandler() handler.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	fmt.Printf("Handle: %+v %+v\n", event, event.Object)
	switch o := event.Object.(type) {
	case *v1alpha1.WildflyAppServer:

		// Ignore the delete event since the garbage collector will clean up all secondary resources for the CR
		// All secondary resources must have the CR set as their OwnerReference for this to be the case
		if event.Deleted {
			return nil
		}

		// Create the deployment if it doesn't exist
		dep := getDeployment(o)
		err := action.Create(dep)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create deployment: %v", err)
		}

		// Ensure the deployment size is the same as the spec
		err = query.Get(dep)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %v", err)
		}
		size := o.Spec.NodeCount
		if *dep.Spec.Replicas != size {
			dep.Spec.Replicas = &size
			err = action.Update(dep)
			if err != nil {
				return fmt.Errorf("failed to update deployment: %v", err)
			}
		}

		// Update the WildflyAppServer status with the pod names
		podList := podList()
		labelSelector := labels.SelectorFromSet(getLabels(o)).String()
		listOps := &metav1.ListOptions{LabelSelector: labelSelector}
		err = query.List(o.Namespace, podList, query.WithListOptions(listOps))
		if err != nil {
			return fmt.Errorf("failed to list pods: %v", err)
		}
		podNames := getPodNames(podList.Items)
		if !reflect.DeepEqual(podNames, o.Status.Nodes) {
			o.Status.Nodes = podNames
			err := action.Update(o)
			if err != nil {
				return fmt.Errorf("failed to update infinispan status: %v", err)
			}
		}

		// Create the Service for Infinispan
		ser := getService(o)
		err = action.Create(ser)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create service: %v", err)
		}
	}
	return nil
}

// podList returns a v1.PodList object
func podList() *v1.PodList {
	return &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []v1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func getDeployment(cr *v1alpha1.WildflyAppServer) *appsv1.Deployment {
	labelMap := getLabels(cr)
	labelSelector := labels.SelectorFromSet(labelMap).String()
	nodeCount := cr.Spec.NodeCount
	image := cr.Spec.Image
	appName := cr.Name
	maxUnavailable := intstr.FromInt(1)
	maxSurge := intstr.FromInt(1)

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &nodeCount,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
					MaxSurge:       &maxSurge,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labelMap,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelMap,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Image: image,
						Name:  appName,
						Env: []v1.EnvVar{
							{
								Name: "KUBERNETES_NAMESPACE",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							{
								Name:  "KUBERNETES_LABELS",
								Value: labelSelector,
							},
						},
						Ports: []v1.ContainerPort{
							{
								ContainerPort: 8080,
								Name:          "http",
							},
							{
								ContainerPort: 9990,
								Name:          "management",
							},
							{
								ContainerPort: 7600,
								Name:          "jgroups-tcp",
							},
							{
								ContainerPort: 57600,
								Name:          "jgroups-tcp-fd",
							},
						},
						LivenessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/",
									Port: intstr.FromString("http"),
								},
							},
							InitialDelaySeconds: 120,
							TimeoutSeconds:      5,
							PeriodSeconds:       60,
							SuccessThreshold:    1,
							FailureThreshold:    6,
						},
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/",
									Port: intstr.FromString("http"),
								},
							},
							InitialDelaySeconds: 30,
							TimeoutSeconds:      3,
							PeriodSeconds:       5,
							SuccessThreshold:    2,
							FailureThreshold:    6,
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								"cpu":    *resource.NewMilliQuantity(300, resource.BinarySI),
								"memory": *resource.NewMilliQuantity(512, resource.BinarySI),
							},
						},
					}},
				},
			},
		},
	}
	setOwnerRefrence(dep, cr)
	return dep
}

func setOwnerRefrence(obj metav1.Object, cr *v1alpha1.WildflyAppServer) {
	ownerRef := *metav1.NewControllerRef(cr, schema.GroupVersionKind{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Kind:    "WildflyAppServer",
	})
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

func getLabels(r *v1alpha1.WildflyAppServer) map[string]string {
	labels := make(map[string]string)
	labels["appName"] = r.Name
	if r.Labels != nil {
		for labelKey, labelValue := range r.Labels {
			labels[labelKey] = labelValue
		}
	}
	return labels
}

func getLabelsAsString(labels map[string]string) string {
	var sb strings.Builder
	for labelKey, labelValue := range labels {
		sb.WriteString(labelKey)
		sb.WriteString("=")
		sb.WriteString(labelValue)
	}
	return sb.String()
}

func getService(cr *v1alpha1.WildflyAppServer) *v1.Service {
	ls := getLabels(cr)
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeLoadBalancer,
			Selector: ls,
			Ports: []v1.ServicePort{
				{
					Name: "http",
					Port: 8080,
				},
				{
					Name: "management",
					Port: 9990,
				},
			},
		},
	}
	setOwnerRefrence(service, cr)
	return service
}
