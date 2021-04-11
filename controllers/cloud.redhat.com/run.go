package controllers

import (
	"fmt"
	"os"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/clowder_config"

	cloudredhatcomv1alpha1 "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	cyndi "cloud.redhat.com/clowder/v2/apis/cyndi-operator/v1alpha1"
	strimzi "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var secretCompare schema.GroupVersionKind

func getKindFromObj(scheme *runtime.Scheme, object runtime.Object) (schema.GroupVersionKind, error) {
	gvks, nok, err := scheme.ObjectKinds(object)

	if err != nil {
		return schema.EmptyObjectKind.GroupVersionKind(), err
	}

	if nok {
		return schema.EmptyObjectKind.GroupVersionKind(), fmt.Errorf("object type is unknown")
	}

	return gvks[0], nil
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cloudredhatcomv1alpha1.AddToScheme(scheme))
	utilruntime.Must(strimzi.AddToScheme(scheme))
	utilruntime.Must(cyndi.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme

	secretCompare, _ = getKindFromObj(scheme, &core.Secret{})
}

// Run inits the manager and controllers and then starts the manager
func Run(metricsAddr string, enableLeaderElection bool, config *rest.Config, signalHandler <-chan struct{}, enableWebHooks bool) {
	setupLog.Info("Loaded config", "config", clowder_config.LoadedConfig)

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "068b0003.cloud.redhat.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&ClowdAppReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ClowdApp"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClowdApp")
		os.Exit(1)
	}
	if err = (&ClowdEnvironmentReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ClowdEnvironment"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClowdEnvironment")
		os.Exit(1)
	}
	if err = (&ClowdJobInvocationReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ClowdJobInvocation"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClowdJobInvocation")
		os.Exit(1)
	}

	if enableWebHooks {
		mgr.GetWebhookServer().Register(
			"/mutate-app",
			&webhook.Admission{
				Handler: &mutantAnnotatorApp{
					Client:   mgr.GetClient(),
					Recorder: mgr.GetEventRecorderFor("app"),
				},
			},
		)

		mgr.GetWebhookServer().Register(
			"/mutate-env",
			&webhook.Admission{
				Handler: &mutantAnnotatorEnv{
					Client:   mgr.GetClient(),
					Recorder: mgr.GetEventRecorderFor("env"),
				},
			},
		)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(signalHandler); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
	setupLog.Info("Exiting manager")
}
