package stub

import (
	"github.com/banzaicloud/wildfly-operator/pkg/apis/wildfly/v1alpha1"
	"fmt"
	"reflect"
	"github.com/operator-framework/operator-sdk/pkg/sdk/action"
	"github.com/operator-framework/operator-sdk/pkg/sdk/handler"
	"github.com/operator-framework/operator-sdk/pkg/sdk/types"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/operator-framework/operator-sdk/pkg/sdk/query"
	"k8s.io/apimachinery/pkg/labels"
	"time"
	"strconv"
	"github.com/sirupsen/logrus"
)

func NewHandler() handler.Handler {
	return &Handler{}
}

type Handler struct {
}

const ApplicationConfig = "standalone-full-ha-k8s.xml"
const ApplicationHttpPort int32 = 8080
const ManagementHttpPort int32 = 9990

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	logrus.Infof("Handle: %+v %+v\n", event, event.Object)
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
		if err != nil && !errors.IsAlreadyExists(err) {
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
				return fmt.Errorf("failed to update nodes status: %v", err)
			}
		}

		// Create LoadBalancer
		ser := getService(o)
		err = action.Create(ser)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create service: %v", err)
		}

		if len(o.Status.ExternalAddresses) == 0 {
			err = updateExternalAddresses(o)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func updateExternalAddresses(o *v1alpha1.WildflyAppServer) error {
	o.Status.ExternalAddresses = make(map[string]string)
	for i := 0; i < 30; i++ {
		logrus.Infof("Checking for Loadbalancer ...")

		ser := &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: o.Name,
				Namespace: o.Namespace,
			},
		}
		err := query.Get(ser)
		if err != nil {
			return fmt.Errorf("failed to retrive Loadbalancer status: %v", err)
		}

		if len(ser.Status.LoadBalancer.Ingress) > 0 && len(ser.Status.LoadBalancer.Ingress[0].IP) > 0 {
			logrus.Infof("found at external address: %+v", ser.Status.LoadBalancer.Ingress[0].IP)
			o.Status.ExternalAddresses["application"] = ser.Status.LoadBalancer.Ingress[0].IP + ":" + strconv.Itoa(int(ApplicationHttpPort))
			o.Status.ExternalAddresses["management"] = ser.Status.LoadBalancer.Ingress[0].IP + ":" + strconv.Itoa(int(ManagementHttpPort))

			err := action.Update(o)
			if err != nil {
				return fmt.Errorf("failed to update nodes status: %v", err)
			}

			return nil
		} else {
			logrus.Infof("not found")
		}

		time.Sleep(time.Duration(10 * time.Second))
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
						Args: []string{
							"--server-config=" + ApplicationConfig,
						},
						Ports: []v1.ContainerPort{
							{
								ContainerPort: ApplicationHttpPort,
								Name:          "http",
							},
							{
								ContainerPort: ManagementHttpPort,
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
								"cpu":    *resource.NewMilliQuantity(500, resource.BinarySI),
								"memory": *resource.NewMilliQuantity(512, resource.BinarySI),
							},
						},
					}},
				},
			},
		},
	}
	if len(cr.Spec.ConfigMapName) > 0 && len(cr.Spec.StandaloneConfigKey) > 0 {
		dep.Spec.Template.Spec.Volumes = []v1.Volume{
			{
				Name: "config-volume",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: cr.Spec.ConfigMapName,
						},
						Items: []v1.KeyToPath{
							{
								Key:  cr.Spec.StandaloneConfigKey,
								Path: ApplicationConfig,
							},
						},
					},
				},
			},
		}
		dep.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
			{
				Name:      "config-volume",
				MountPath: "/opt/jboss/wildfly/standalone/configuration/" + ApplicationConfig,
				SubPath:   ApplicationConfig,
			},
		}
	}
	false := true
	dep.Spec.Template.Spec.Containers[0].Env = append(dep.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
		Name: "WILDFLY_ADMIN_USER",
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{Name: cr.Name},
				Key:                  "wildfly-admin-user",
				Optional:             &false,
			},
		},
	})

	dep.Spec.Template.Spec.Containers[0].Env = append(dep.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
		Name: "WILDFLY_ADMIN_PASSWORD",
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{Name: cr.Name},
				Key:                  "wildfly-admin-password",
				Optional:             &false,
			},
		},
	})

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
					Port: ApplicationHttpPort,
				},
				{
					Name: "management",
					Port: ManagementHttpPort,
				},
			},
		},
	}
	setOwnerRefrence(service, cr)
	return service
}
