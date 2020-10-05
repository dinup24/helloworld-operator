/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/dinup24/helloworld-operator/apis/apps/v1alpha1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.helloworld.ibm.com,resources=helloworldren,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.helloworld.ibm.com,resources=helloworldren/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;delete

func (r *HelloWorldReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("helloworld", req.NamespacedName)

	instance := &appsv1alpha1.HelloWorld{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	result, err := r.reconcileDeployment(ctx, instance)
	if result.Requeue == true || err != nil {
		return result, err
	}

	result, err = r.reconcileService(ctx, instance)
	if result.Requeue == true || err != nil {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *HelloWorldReconciler) reconcileDeployment(ctx context.Context, instance *appsv1alpha1.HelloWorld) (ctrl.Result, error) {
	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := deploymentForHelloWorld(instance)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)

		err = r.Create(ctx, dep)

		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := instance.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func (r *HelloWorldReconciler) reconcileService(ctx context.Context, instance *appsv1alpha1.HelloWorld) (ctrl.Result, error) {
	// Check if the service already exists, if not create a new one
	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		ser := serviceForHelloWorld(instance)
		log.Info("Creating a new Service", "Service.Namespace", ser.Namespace, "Service.Name", ser.Name)

		err = r.Create(ctx, ser)

		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", ser.Namespace, "Service.Name", ser.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.HelloWorld{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func int32Ptr(i int32) *int32 { return &i }

func serviceForHelloWorld(helloworld *appsv1alpha1.HelloWorld) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helloworld.Name,
			Namespace: helloworld.Namespace,
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "helloworld",
			},
			Ports: []corev1.ServicePort{
				{
					Port: 9000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 9000,
					},
					Protocol: "TCP",
				},
			},
		},
	}

	//ctrl.SetControllerReference(m, dep, r.Scheme)

	return service
}

func deploymentForHelloWorld(helloworld *appsv1alpha1.HelloWorld) *appsv1.Deployment {
	image := "dinup24/bedrock-helloworld:v1"
	if helloworld.Spec.Image != "" {
		image = helloworld.Spec.Image
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helloworld.Name,
			Namespace: helloworld.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "helloworld",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "helloworld",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "helloworld",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9000,
									Protocol:      "TCP",
								},
							},
						},
					},
				},
			},
		},
	}

	//ctrl.SetControllerReference(m, dep, r.Scheme)

	return deployment
}
