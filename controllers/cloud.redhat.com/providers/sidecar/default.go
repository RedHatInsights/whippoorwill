package sidecar

import (
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1/common"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	cronjobProvider "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/cronjob"
	deployProvider "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/deployment"

	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type sidecarProvider struct {
	providers.Provider
}

func NewSidecarProvider(p *providers.Provider) (providers.ClowderProvider, error) {
	return &sidecarProvider{Provider: *p}, nil
}

func (sc *sidecarProvider) Provide(app *crd.ClowdApp, c *config.AppConfig) error {
	for _, deployment := range app.Spec.Deployments {

		d := &apps.Deployment{}

		sc.Cache.Get(deployProvider.CoreDeployment, d, app.GetDeploymentNamespacedName(&deployment))

		for _, sidecar := range deployment.PodSpec.Sidecars {
			switch sidecar.Name {
			case "splunk":
				if sidecar.Enabled && sc.Env.Spec.Providers.Sidecars.Splunk.Enabled {
					cont := getSplunk()
					if cont != nil {
						d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, *cont)
						d.Spec.Template.Spec.Volumes = appendSplunkVolumes(d.Spec.Template.Spec.Volumes, app.Name)
					}
				}
			case "token-refresher":
				if sidecar.Enabled && sc.Env.Spec.Providers.Sidecars.TokenRefresher.Enabled {
					cont := getTokenRefresher(app.Name)
					if cont != nil {
						d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, *cont)
					}
				}
			default:
				return fmt.Errorf("%s is not a valid sidecar name", sidecar.Name)
			}

			sc.Cache.Update(deployProvider.CoreDeployment, d)
		}
	}

	for _, cronJob := range app.Spec.Jobs {

		cj := &batch.CronJob{}

		sc.Cache.Get(cronjobProvider.CoreCronJob, cj, app.GetCronJobNamespacedName(&cronJob))

		for _, sidecar := range cronJob.PodSpec.Sidecars {
			switch sidecar.Name {
			case "splunk":
				if sidecar.Enabled && sc.Env.Spec.Providers.Sidecars.Splunk.Enabled {
					cont := getSplunk()
					if cont != nil {
						cj.Spec.JobTemplate.Spec.Template.Spec.Containers = append(cj.Spec.JobTemplate.Spec.Template.Spec.Containers, *cont)
						cj.Spec.JobTemplate.Spec.Template.Spec.Volumes = appendSplunkVolumes(cj.Spec.JobTemplate.Spec.Template.Spec.Volumes, app.Name)
					}
				}
			case "token-refresher":
				if sidecar.Enabled && sc.Env.Spec.Providers.Sidecars.TokenRefresher.Enabled {
					cont := getTokenRefresher(app.Name)
					if cont != nil {
						cj.Spec.JobTemplate.Spec.Template.Spec.Containers = append(cj.Spec.JobTemplate.Spec.Template.Spec.Containers, *cont)
					}
				}
			default:
				return fmt.Errorf("%s is not a valid sidecar name", sidecar.Name)
			}

			sc.Cache.Update(deployProvider.CoreDeployment, cj)
		}
	}

	return nil
}

func appendSplunkVolumes(vols []core.Volume, appName string) []core.Volume {
	vols = append(vols, core.Volume{
		Name: "splunkpem",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName:  fmt.Sprintf("%s-splunk", appName),
				DefaultMode: common.Int32Ptr(272),
			},
		},
	})
	vols = append(vols, core.Volume{
		Name: "inputsconf",
		VolumeSource: core.VolumeSource{
			ConfigMap: &core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: fmt.Sprintf("%s-splunk", appName),
				},
				DefaultMode: common.Int32Ptr(436),
			},
		},
	})
	return vols
}

func getTokenRefresher(appName string) *core.Container {
	cont := core.Container{}

	cont.Name = "token-refresher"
	cont.Image = "quay.io/observatorium/token-refresher:master-2021-02-05-5da9663"
	cont.Args = []string{
		"--oidc.audience=observatorium-telemeter",
		"--oidc.client-id=$(CLIENT_ID)",
		"--oidc.client-secret=$(CLIENT_SECRET)",
		"--oidc.issuer-url=$(ISSUER_URL)",
		"--url=$(URL)",
		"--web.listen=:8082",
	}
	cont.Resources = core.ResourceRequirements{
		Limits: core.ResourceList{
			"cpu":    resource.MustParse("100m"),
			"memory": resource.MustParse("150Mi"),
		},
		Requests: core.ResourceList{
			"cpu":    resource.MustParse("50m"),
			"memory": resource.MustParse("128Mi"),
		},
	}
	cont.Env = []core.EnvVar{
		{
			Name: "CLIENT_ID",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: fmt.Sprintf("%s-token-refresher", appName),
					},
					Key: "CLIENT_ID",
				},
			},
		},
		{
			Name: "CLIENT_SECRET",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: fmt.Sprintf("%s-token-refresher", appName),
					},
					Key: "CLIENT_SECRET",
				},
			},
		},
		{
			Name: "ISSUER_URL",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: fmt.Sprintf("%s-token-refresher", appName),
					},
					Key: "ISSUER_URL",
				},
			},
		},
		{
			Name: "URL",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: fmt.Sprintf("%s-token-refresher", appName),
					},
					Key: "URL",
				},
			},
		},
	}

	return &cont
}

func getSplunk() *core.Container {
	cont := core.Container{}

	cont.Name = "splunk"
	cont.Image = "quay.io/cloudservices/rhsm-splunk-forwarder:8f72cfb"
	cont.VolumeMounts = []core.VolumeMount{
		{
			Name:      "splunkpem",
			MountPath: "/tls/splunk.pem",
			SubPath:   "splunk.pem",
		},
		{
			Name:      "inputsconf",
			MountPath: "/opt/splunkforwarder/etc/system/local/inputs.conf",
			SubPath:   "inputs.conf",
		},
	}
	cont.Resources = core.ResourceRequirements{
		Limits: core.ResourceList{
			"cpu":    resource.MustParse("100m"),
			"memory": resource.MustParse("150Mi"),
		},
		Requests: core.ResourceList{
			"cpu":    resource.MustParse("50m"),
			"memory": resource.MustParse("128Mi"),
		},
	}
	cont.Env = []core.EnvVar{{
		Name: "SPLUNKMETA_namespace",
		ValueFrom: &core.EnvVarSource{
			FieldRef: &core.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	}}

	return &cont
}
