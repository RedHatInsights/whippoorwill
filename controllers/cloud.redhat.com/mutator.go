package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type mutantAnnotatorApp struct {
	Client   client.Client
	Recorder record.EventRecorder
	decoder  *admission.Decoder
}

func (a *mutantAnnotatorApp) Handle(ctx context.Context, req admission.Request) admission.Response {
	app := &crd.ClowdApp{}
	err := a.decoder.Decode(req, app)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if app.Spec.Pods != nil {
		// TODO events don't work here due to object not being "created" yet - could use status to deal with in reconcile
		// a.Recorder.Eventf(app, "Warning", "ClowdAppOldVersion", "ClowdApp spec [%s] is using deprecated Pods", app.Name)

		deps := []crd.Deployment{}
		for _, pod := range app.Spec.Pods {
			dep := crd.Deployment{
				Name:        pod.Name,
				Web:         pod.Web,
				MinReplicas: pod.MinReplicas,
				PodSpec: crd.PodSpec{
					Image:          pod.Image,
					InitContainers: pod.InitContainers,
					Command:        pod.Command,
					Args:           pod.Args,
					Env:            pod.Env,
					Resources:      pod.Resources,
					LivenessProbe:  pod.LivenessProbe,
					ReadinessProbe: pod.ReadinessProbe,
					Volumes:        pod.Volumes,
					VolumeMounts:   pod.VolumeMounts,
				},
			}
			deps = append(deps, dep)
		}
		if app.Spec.Deployments != nil {
			app.Spec.Deployments = append(app.Spec.Deployments, deps...)
		} else {
			app.Spec.Deployments = deps
		}
		app.Spec.Pods = nil
	}

	for i, deployment := range app.Spec.Deployments {
		if deployment.Web {
			// TODO events don't work here due to object not being "created" yet - could use status to deal with in reconcile
			// a.Recorder.Eventf(app, "Warning", "ClowdAppOldVersion", "ClowdApp spec [%s] is using deprecated Web", app.Name)

			app.Spec.Deployments[i].WebServices = crd.WebServices{
				Public: crd.PublicWebService{
					Enabled: true,
				},
				Private: crd.PrivateWebService{},
				Metrics: crd.MetricsWebService{},
			}
			app.Spec.Deployments[i].Web = false
		}
	}

	marshaledApp, err := json.Marshal(app)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledApp)
}

func (a *mutantAnnotatorApp) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

type mutantAnnotatorEnv struct {
	Client   client.Client
	Recorder record.EventRecorder
	decoder  *admission.Decoder
}

func (a *mutantAnnotatorEnv) Handle(ctx context.Context, req admission.Request) admission.Response {
	env := &crd.ClowdEnvironment{}
	err := a.decoder.Decode(req, env)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	kafkaConfig := env.Spec.Providers.Kafka

	if kafkaConfig.Namespace != "" || kafkaConfig.ClusterName != "" {
		env.Spec.Providers.Kafka.Cluster = crd.KafkaClusterConfig{
			Name:      kafkaConfig.ClusterName,
			Namespace: kafkaConfig.Namespace,
		}
		env.Spec.Providers.Kafka.ClusterName = ""
		env.Spec.Providers.Kafka.Namespace = ""
	}

	if kafkaConfig.ConnectClusterName != "" || kafkaConfig.ConnectNamespace != "" {
		env.Spec.Providers.Kafka.Connect = crd.KafkaConnectClusterConfig{
			Name:      kafkaConfig.ConnectClusterName,
			Namespace: kafkaConfig.ConnectNamespace,
		}
		env.Spec.Providers.Kafka.ConnectClusterName = ""
		env.Spec.Providers.Kafka.ConnectNamespace = ""
	}

	marshaledApp, err := json.Marshal(env)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledApp)
}

func (a *mutantAnnotatorEnv) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
