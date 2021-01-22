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
	"fmt"
	"sort"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	// Import the providers to initialize them
	_ "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/database"
	_ "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/featureflags"
	_ "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/inmemorydb"
	_ "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/kafka"
	_ "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/logging"
	_ "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/objectstore"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const envFinalizer = "finalizer.env.cloud.redhat.com"

// ClowdEnvironmentReconciler reconciles a ClowdEnvironment object
type ClowdEnvironmentReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cloud.redhat.com,resources=clowdenvironments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.redhat.com,resources=clowdenvironments/status,verbs=get;update;patch

//Reconcile fn
func (r *ClowdEnvironmentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("env", req.Name)
	ctx := context.WithValue(context.Background(), errors.ClowdKey("log"), &log)
	ctx = context.WithValue(ctx, errors.ClowdKey("recorder"), &r.Recorder)
	proxyClient := ProxyClient{Client: r.Client, Ctx: ctx}
	env := crd.ClowdEnvironment{}
	err := r.Client.Get(ctx, req.NamespacedName, &env)

	if err != nil {
		if k8serr.IsNotFound(err) {
			// Must have been deleted
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if env.Status.TargetNamespace == "" {
		if env.Spec.TargetNamespace != "" {
			namespace := core.Namespace{}
			namespaceName := types.NamespacedName{
				Name: env.Spec.TargetNamespace,
			}
			err := r.Client.Get(ctx, namespaceName, &namespace)
			if err != nil {
				r.Recorder.Eventf(&env, "Warning", "NamespaceMissing", "Requested Target Namespace [%s] is missing", env.Spec.TargetNamespace)
				return ctrl.Result{Requeue: true}, err
			}
			env.Status.TargetNamespace = env.Spec.TargetNamespace
		} else {
			env.Status.TargetNamespace = env.GenerateTargetNamespace()
			namespace := &core.Namespace{}
			namespace.SetName(env.Status.TargetNamespace)
			err := r.Client.Create(ctx, namespace)
			if err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		}
		err := r.Client.Status().Update(ctx, &env)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	ctx = context.WithValue(ctx, errors.ClowdKey("obj"), &env)

	provider := providers.Provider{
		Ctx:    ctx,
		Client: &proxyClient,
		Env:    &env,
	}

	err = runProvidersForEnv(provider)

	if err == nil {
		r.Log.Info("Reconciliation successful", "env", env.Name)
		if _, ok := managedEnvironments[env.Name]; ok == false {
			managedEnvironments[env.Name] = true
		}
		managedEnvsMetric.Set(float64(len(managedEnvironments)))
	}

	requeue := errors.HandleError(ctx, err)
	if requeue {
		r.Log.Error(err, "Requeueing due to error")
	}

	isEnvMarkedForDeletion := env.GetDeletionTimestamp() != nil
	if isEnvMarkedForDeletion {
		if contains(env.GetFinalizers(), envFinalizer) {
			if err := r.finalizeEnvironment(log, &env); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&env, envFinalizer)
			err := r.Update(ctx, &env)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !contains(env.GetFinalizers(), envFinalizer) {
		if err := r.addFinalizer(log, &env); err != nil {
			return ctrl.Result{}, err
		}
	}

	if !requeue {
		// Delete all resources that are not used anymore
		uid := env.ObjectMeta.UID
		err := proxyClient.Reconcile(uid)
		if err != nil {
			return ctrl.Result{Requeue: requeue}, nil
		}
	}

	err = r.SetAppInfo(provider)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	SetDeploymentStatus(ctx, &proxyClient, &env)

	env.Status.Ready = false
	if env.Status.Deployments.ManagedDeployments == env.Status.Deployments.ReadyDeployments {
		env.Status.Ready = true
	}

	err = proxyClient.Status().Update(ctx, &env)

	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: requeue}, nil
}

func runProvidersForEnv(provider providers.Provider) error {
	for _, provAcc := range providers.ProvidersRegistration.Registry {
		if _, err := provAcc.SetupProvider(&provider); err != nil {
			return errors.Wrap(fmt.Sprintf("getprov: %s", provAcc.Name), err)
		}
	}
	return nil
}

// SetupWithManager sets up with manager
func (r *ClowdEnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("env")
	return ctrl.NewControllerManagedBy(mgr).
		Owns(&apps.Deployment{}).
		Owns(&core.Service{}).
		Watches(
			&source.Kind{Type: &crd.ClowdApp{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(r.envToEnqueueUponAppUpdate),
			},
		).
		For(&crd.ClowdEnvironment{}).
		WithEventFilter(ignoreStatusUpdatePredicate()).
		Complete(r)
}

func (r *ClowdEnvironmentReconciler) envToEnqueueUponAppUpdate(a handler.MapObject) []reconcile.Request {
	ctx := context.Background()
	obj := types.NamespacedName{
		Name:      a.Meta.GetName(),
		Namespace: a.Meta.GetNamespace(),
	}

	// Get the ClowdEnvironment resource

	app := crd.ClowdApp{}
	err := r.Client.Get(ctx, obj, &app)

	if err != nil {
		if k8serr.IsNotFound(err) {
			// Must have been deleted
			return []reconcile.Request{}
		}
		r.Log.Error(err, "Failed to fetch ClowdApp")
		return nil
	}

	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{
			Name: app.Spec.EnvName,
		},
	}}
}

func (r *ClowdEnvironmentReconciler) SetAppInfo(p providers.Provider) error {
	// Get all the ClowdApp resources
	appList := crd.ClowdAppList{}
	r.Client.List(p.Ctx, &appList)
	apps := []crd.AppInfo{}

	appMap := map[string]crd.ClowdApp{}
	names := []string{}

	for _, app := range appList.Items {

		if app.Spec.EnvName != p.Env.Name {
			continue
		}
		name := fmt.Sprintf("%s-%s", app.Name, app.Namespace)
		names = append(names, name)
		appMap[name] = app
	}

	sort.Strings(names)

	// Populate
	for _, name := range names {
		app := appMap[name]

		if app.Spec.EnvName != p.Env.Name {
			continue
		}

		if app.GetDeletionTimestamp() != nil {
			continue
		}

		appstatus := crd.AppInfo{
			Name:        app.Name,
			Deployments: []crd.DeploymentInfo{},
		}

		depMap := map[string]crd.Deployment{}
		depNames := []string{}

		for _, pod := range app.Spec.Deployments {
			depNames = append(depNames, pod.Name)
			depMap[pod.Name] = pod
		}

		sort.Strings(depNames)

		for _, podName := range depNames {
			pod := depMap[podName]

			deploymentStatus := crd.DeploymentInfo{
				Name: fmt.Sprintf("%s-%s", app.Name, pod.Name),
			}
			if pod.Web {
				deploymentStatus.Hostname = fmt.Sprintf("%s.%s.svc", deploymentStatus.Name, app.Namespace)
				deploymentStatus.Port = p.Env.Spec.Providers.Web.Port
			}
			appstatus.Deployments = append(appstatus.Deployments, deploymentStatus)
		}
		apps = append(apps, appstatus)
	}

	p.Env.Status.Apps = apps
	return nil
}

func (r *ClowdEnvironmentReconciler) finalizeEnvironment(reqLogger logr.Logger, e *crd.ClowdEnvironment) error {

	if _, ok := managedEnvironments[e.Name]; ok == true {
		delete(managedEnvironments, e.Name)
	}
	if e.Spec.TargetNamespace == "" {
		namespace := &core.Namespace{}
		namespace.SetName(e.Status.TargetNamespace)
		reqLogger.Info(fmt.Sprintf("Removing auto-generated namespace for %s", e.Name))
		r.Recorder.Eventf(e, "Warning", "NamespaceDeletion", "Clowder Environment [%s] had no targetNamespace, deleting generated namespace", e.Name)
		r.Delete(context.TODO(), namespace)
	}
	managedEnvsMetric.Set(float64(len(managedEnvironments)))
	reqLogger.Info("Successfully finalized ClowdEnvironment")
	return nil
}

func (r *ClowdEnvironmentReconciler) addFinalizer(reqLogger logr.Logger, e *crd.ClowdEnvironment) error {
	reqLogger.Info("Adding Finalizer for the ClowdEnvironment")
	controllerutil.AddFinalizer(e, envFinalizer)

	// Update CR
	err := r.Update(context.TODO(), e)
	if err != nil {
		reqLogger.Error(err, "Failed to update ClowdEnvironment with finalizer")
		return err
	}
	return nil
}
