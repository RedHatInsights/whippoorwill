package deployment

import (
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetDeploymentName(app *crd.ClowdApp, deployment *crd.Deployment) string {
	return fmt.Sprintf("%s-%s", app.Name, deployment.Name)
}

func (dp *deploymentProvider) makeDeployment(deployment crd.Deployment, app *crd.ClowdApp) error {

	d := &apps.Deployment{}
	nn := types.NamespacedName{
		Name:      GetDeploymentName(app, &deployment),
		Namespace: app.Namespace,
	}

	if err := dp.Cache.Create(CoreDeployment, nn, d); err != nil {
		return err
	}

	initDeployment(app, dp.Env, d, nn, deployment)

	if err := dp.Cache.Update(CoreDeployment, d); err != nil {
		return err
	}

	return nil
}

func initDeployment(app *crd.ClowdApp, env *crd.ClowdEnvironment, d *apps.Deployment, nn types.NamespacedName, deployment crd.Deployment) {
	labels := app.GetLabels()
	labels["pod"] = nn.Name
	app.SetObjectMeta(d, crd.Name(nn.Name), crd.Labels(labels))

	pod := deployment.PodSpec
	d.Spec.Template.SetAnnotations(make(map[string]string))
	d.Spec.Replicas = deployment.MinReplicas
	d.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	d.Spec.Template.ObjectMeta.Labels = labels
	d.Spec.Strategy = apps.DeploymentStrategy{
		Type: apps.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &apps.RollingUpdateDeployment{
			MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: string("25%")},
			MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: string("25%")},
		},
	}
	d.Spec.ProgressDeadlineSeconds = utils.Int32Ptr(600)

	d.Spec.Template.Spec.ImagePullSecrets = []core.LocalObjectReference{
		{Name: "quay-cloudservices-pull"},
	}

	envvar := pod.Env
	envvar = append(envvar, core.EnvVar{Name: "ACG_CONFIG", Value: "/cdapp/cdappconfig.json"})

	var livenessProbe core.Probe
	var readinessProbe core.Probe

	baseProbe := core.Probe{
		Handler: core.Handler{
			HTTPGet: &core.HTTPGetAction{
				Path:   "/healthz",
				Scheme: "HTTP",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: env.Spec.Providers.Web.Port,
				},
			},
		},
		FailureThreshold:    3,
		InitialDelaySeconds: 10,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
	}
	if pod.LivenessProbe != nil {
		livenessProbe = *pod.LivenessProbe
	} else if deployment.WebServices.Public.Enabled {
		livenessProbe = baseProbe
	}
	if pod.ReadinessProbe != nil {
		readinessProbe = *pod.ReadinessProbe
	} else if deployment.WebServices.Public.Enabled {
		readinessProbe = baseProbe
		readinessProbe.InitialDelaySeconds = 45
	}

	c := core.Container{
		Name:            nn.Name,
		Image:           pod.Image,
		Command:         pod.Command,
		Args:            pod.Args,
		Env:             envvar,
		Resources:       ProcessResources(&pod, env),
		VolumeMounts:    pod.VolumeMounts,
		ImagePullPolicy: core.PullIfNotPresent,
	}

	if (core.Probe{}) != livenessProbe {
		c.LivenessProbe = &livenessProbe
	}
	if (core.Probe{}) != readinessProbe {
		c.ReadinessProbe = &readinessProbe
	}

	// TODO: THIS NEEDS TO GO IN SERVICE
	if deployment.WebServices.Public.Enabled {
		c.Ports = append(c.Ports, core.ContainerPort{
			Name:          "web",
			ContainerPort: env.Spec.Providers.Web.Port,
		})
	}

	c.VolumeMounts = append(c.VolumeMounts, core.VolumeMount{
		Name:      "config-secret",
		MountPath: "/cdapp/",
	})

	d.Spec.Template.Spec.Containers = []core.Container{c}

	d.Spec.Template.Spec.InitContainers = ProcessInitContainers(nn, &c, pod.InitContainers)

	d.Spec.Template.Spec.Volumes = pod.Volumes
	d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, core.Volume{
		Name: "config-secret",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: app.ObjectMeta.Name,
			},
		},
	})

	ApplyPodAntiAffinity(&d.Spec.Template)
}

// ProcessInitContainers returns a container object which has been processed from the container spec.
func ProcessInitContainers(nn types.NamespacedName, c *core.Container, ics []crd.InitContainer) []core.Container {
	if len(ics) == 0 {
		return []core.Container{}
	}
	containerList := make([]core.Container, len(ics))

	for i, ic := range ics {
		icStruct := core.Container{
			Name:            nn.Name + "-init",
			Image:           c.Image,
			Command:         ic.Command,
			Args:            ic.Args,
			Resources:       c.Resources,
			VolumeMounts:    c.VolumeMounts,
			ImagePullPolicy: c.ImagePullPolicy,
		}

		if ic.InheritEnv {
			// The idea here is that you can override the inherited values by
			// setting them on the initContainer env
			icStruct.Env = []core.EnvVar{}
			usedVars := make(map[string]bool)

			for _, iEnvVar := range ic.Env {
				icStruct.Env = append(icStruct.Env, core.EnvVar{
					Name:      iEnvVar.Name,
					Value:     iEnvVar.Value,
					ValueFrom: iEnvVar.ValueFrom,
				})
				usedVars[iEnvVar.Name] = true
			}
			for _, e := range c.Env {
				if _, ok := usedVars[e.Name]; !ok {
					icStruct.Env = append(icStruct.Env, core.EnvVar{
						Name:      e.Name,
						Value:     e.Value,
						ValueFrom: e.ValueFrom,
					})
				}
			}
		} else {
			for _, envvar := range ic.Env {
				icStruct.Env = append(icStruct.Env, core.EnvVar{
					Name:      envvar.Name,
					Value:     envvar.Value,
					ValueFrom: envvar.ValueFrom,
				})
			}
		}

		icStruct.Env = append(
			icStruct.Env, core.EnvVar{Name: "ACG_CONFIG", Value: "/cdapp/cdappconfig.json"},
		)

		containerList[i] = icStruct
	}
	return containerList
}

// ApplyPodAntiAffinity applies pod anti affinity rules to a pod template
func ApplyPodAntiAffinity(t *core.PodTemplateSpec) {
	labelSelector := &metav1.LabelSelector{MatchLabels: t.Labels}
	t.Spec.Affinity = &core.Affinity{PodAntiAffinity: &core.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{{
			Weight: 100,
			PodAffinityTerm: core.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   "failure-domain.beta.kubernetes.io/zone",
			},
		}, {
			Weight: 99,
			PodAffinityTerm: core.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   "kubernetes.io/hostname",
			},
		}},
	}}
}

// ProcessResources takes a pod spec and a clowd environment and returns the resource requirements
// object.
func ProcessResources(pod *crd.PodSpec, env *crd.ClowdEnvironment) core.ResourceRequirements {
	var lcpu, lmemory, rcpu, rmemory resource.Quantity
	nullCPU := resource.Quantity{Format: resource.DecimalSI}
	nullMemory := resource.Quantity{Format: resource.BinarySI}

	if *pod.Resources.Limits.Cpu() != nullCPU {
		lcpu = pod.Resources.Limits["cpu"]
	} else {
		lcpu = env.Spec.ResourceDefaults.Limits["cpu"]
	}

	if *pod.Resources.Limits.Memory() != nullMemory {
		lmemory = pod.Resources.Limits["memory"]
	} else {
		lmemory = env.Spec.ResourceDefaults.Limits["memory"]
	}

	if *pod.Resources.Requests.Cpu() != nullCPU {
		rcpu = pod.Resources.Requests["cpu"]
	} else {
		rcpu = env.Spec.ResourceDefaults.Requests["cpu"]
	}

	if *pod.Resources.Requests.Memory() != nullMemory {
		rmemory = pod.Resources.Requests["memory"]
	} else {
		rmemory = env.Spec.ResourceDefaults.Requests["memory"]
	}

	return core.ResourceRequirements{
		Limits: core.ResourceList{
			"cpu":    lcpu,
			"memory": lmemory,
		},
		Requests: core.ResourceList{
			"cpu":    rcpu,
			"memory": rmemory,
		},
	}
}
