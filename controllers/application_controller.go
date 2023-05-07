/*
Copyright 2023 kpw.

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
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/kpw123/application-operator/api/v1"
)

const GenericRequeueDuration = 1 * time.Minute

var CounterReconcileApplication int64

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.init0.cn,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.init0.cn,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.init0.cn,resources=applications/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO(user): your logic here
	<-time.NewTicker(100 * time.Millisecond).C
	log := log.FromContext(ctx)

	CounterReconcileApplication += 1
	log.Info("Starting a reconcile", "number", CounterReconcileApplication)

	app := &appsv1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Application not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get the Application,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err

	}

	var result ctrl.Result
	var err error

	result, err = r.reconcileDeployment(ctx, app)
	if err != nil {
		log.Error(err, "Failed to reconcile Deployment.")
		return result, err
	}

	result, err = r.reconcileService(ctx, app)
	if err != nil {
		log.Error(err, "Failed to reconcile Service.")
		return result, err
	}

	log.Info("All resource have been reconcile")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Application{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				setupLog.Info("The application has deleted.",
					"name", deleteEvent.Object.GetName())
				return false
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				if updateEvent.ObjectNew.GetResourceVersion() ==
					updateEvent.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(updateEvent.ObjectNew.(*appsv1.Application).Spec,
					updateEvent.ObjectOld.(*appsv1.Application).Spec) {
					return false
				}
				return true
			},
		})).
		Owns(&v1.Deployment{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				setupLog.Info("The Deployment has deleted.",
					"name", deleteEvent.Object.GetName())
				return true
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				if updateEvent.ObjectNew.GetResourceVersion() ==
					updateEvent.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(updateEvent.ObjectNew.(*v1.Deployment).Spec,
					updateEvent.ObjectOld.(*v1.Deployment).Spec) {
					return false
				}
				return true
			},
			GenericFunc: nil,
		})).
		Owns(&v12.Service{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				setupLog.Info("The Service has deleted.",
					"name", deleteEvent.Object.GetName())
				return true
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				if updateEvent.ObjectNew.GetResourceVersion() ==
					updateEvent.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(updateEvent.ObjectNew.(*v12.Service).Spec,
					updateEvent.ObjectOld.(*v12.Service).Spec) {
					return false
				}
				return true
			},
		})).
		Complete(r)
}

func (r *ApplicationReconciler) reconcileDeployment(ctx context.Context, app *appsv1.Application) (result ctrl.Result, err error) {
	log := log.FromContext(ctx)

	dp := &v1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, dp)
	if err == nil {
		log.Info("The Deployment has already exist.")
		if reflect.DeepEqual(dp.Status, app.Status.Workflow) {
			return ctrl.Result{}, nil
		}
		app.Status.Workflow = dp.Status
		if err := r.Status().Update(ctx, app); err != nil {
			log.Error(err, "Failed to update Application status")
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		log.Info("The Application status has been updated.")
		return ctrl.Result{}, nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get Deployment,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	newDp := &v1.Deployment{}
	newDp.SetName(app.Name)
	newDp.SetNamespace(app.Namespace)
	newDp.SetLabels(app.Labels)
	newDp.Spec = app.Spec.Deployment.DeploymentSpec
	newDp.Spec.Template.SetLabels(app.Labels)

	if err := ctrl.SetControllerReference(app, newDp, r.Scheme); err != nil {
		log.Error(err, "Failed to SetControllerReference,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	if err := r.Create(ctx, newDp); err != nil {
		log.Error(err, "Failed to create Deployment,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	log.Info("The Deployment has been created.")
	return result, nil
}
