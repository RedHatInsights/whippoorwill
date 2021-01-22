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

package makers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"github.com/go-logr/logr"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"

	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//DependencyMaker makes the DependencyConfig object
type DependencyMaker struct {
	*Maker
	config []config.DependencyEndpoint
}

//Maker struct for passing variables into SubMakers
type Maker struct {
	App     *crd.ClowdApp
	Env     *crd.ClowdEnvironment
	Client  client.Client
	Ctx     context.Context
	Request *ctrl.Request
	Log     logr.Logger
}

func New(maker *Maker) (*Maker, error) {
	return maker, nil
}

//Make generates objects and dependencies for operator
func (m *Maker) Make() error {
	c, err := m.runProviders()

	if err != nil {
		return err
	}

	if err := m.makeDependencies(c); err != nil {
		return err
	}

	hash, err := m.persistConfig(c)

	if err != nil {
		return err
	}

	for _, deployment := range m.App.Spec.Deployments {

		if err := m.makeDeployment(deployment, m.App, hash); err != nil {
			return err
		}

		if err := m.makeService(deployment, m.App); err != nil {
			return err
		}
	}

	for _, job := range m.App.Spec.Jobs {

		if err := m.makeJob(job, m.App, hash); err != nil {
			return err
		}
	}

	return nil
}

func (m *Maker) makeService(deployment crd.Deployment, app *crd.ClowdApp) error {

	s := core.Service{}
	nn := types.NamespacedName{
		Name:      fmt.Sprintf("%v-%v", app.Name, deployment.Name),
		Namespace: m.Request.Namespace,
	}
	err := m.Client.Get(m.Ctx, nn, &s)

	update, err := utils.UpdateOrErr(err)
	if err != nil {
		return err
	}

	ports := []core.ServicePort{
		{Name: "metrics", Port: m.Env.Spec.Providers.Metrics.Port, Protocol: "TCP"},
	}

	if deployment.Web == true {
		webPort := core.ServicePort{Name: "web", Port: m.Env.Spec.Providers.Web.Port, Protocol: "TCP"}
		ports = append(ports, webPort)
	}

	utils.MakeService(&s, nn, map[string]string{"pod": nn.Name}, ports, m.App)

	return update.Apply(m.Ctx, m.Client, &s)
}

func (m *Maker) persistConfig(c *config.AppConfig) (string, error) {

	// In any case, we want to overwrite the secret, so this just
	// tests to see if the secret exists
	err := m.Client.Get(m.Ctx, m.Request.NamespacedName, &core.Secret{})

	update, err := utils.UpdateOrErr(err)
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(c)
	if err != nil {
		return "", errors.Wrap("Failed to marshal config JSON", err)
	}

	h := sha256.New()
	h.Write([]byte(jsonData))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	secret := core.Secret{
		StringData: map[string]string{
			"cdappconfig.json": string(jsonData),
		},
	}

	m.App.SetObjectMeta(&secret)

	err = update.Apply(m.Ctx, m.Client, &secret)
	return hash, err
}

func (m *Maker) getConfig() (*config.AppConfig, error) {
	secret := core.Secret{}
	appConfig := config.AppConfig{}
	err := m.Client.Get(m.Ctx, m.Request.NamespacedName, &secret)

	if err != nil {
		return nil, errors.Wrap("Failed to fetch config seret", err)
	}

	err = json.Unmarshal([]byte(secret.Data["cdappconfig.json"]), &appConfig)
	if err != nil {
		return nil, errors.Wrap("Failed to unmarshal JSON", err)
	}
	return &appConfig, nil
}

func applyPodAntiAffinity(t *core.PodTemplateSpec) {
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

func (m *Maker) makeJob(job crd.Job, app *crd.ClowdApp, hash string) error {

	nn := types.NamespacedName{
		Name:      fmt.Sprintf("%v-%v", app.Name, job.Name),
		Namespace: m.Request.Namespace,
	}

	c := batch.CronJob{}
	err := m.Client.Get(m.Ctx, nn, &c)

	update, err := utils.UpdateOrErr(err)
	if err != nil {
		return err
	}

	initJob(m.App, m.Env, &c, nn, job, hash)

	if err := update.Apply(m.Ctx, m.Client, &c); err != nil {
		return err
	}

	return nil
}

func initJob(app *crd.ClowdApp, env *crd.ClowdEnvironment, cj *batch.CronJob, nn types.NamespacedName, job crd.Job, hash string) {
	labels := app.GetLabels()
	labels["pod"] = nn.Name
	app.SetObjectMeta(cj, crd.Name(nn.Name), crd.Labels(labels))

	pod := job.PodSpec

	cj.Spec.Schedule = job.Schedule

	cj.Spec.JobTemplate.ObjectMeta.Labels = labels
	cj.Spec.JobTemplate.Spec.Template.ObjectMeta.Labels = labels

	if job.RestartPolicy == "" {
		cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = core.RestartPolicyNever
	} else {
		cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = job.RestartPolicy
	}

	cj.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets = []core.LocalObjectReference{
		{Name: "quay-cloudservices-pull"},
	}

	envvar := pod.Env
	envvar = append(envvar, core.EnvVar{Name: "ACG_CONFIG", Value: "/cdapp/cdappconfig.json"})

	var livenessProbe core.Probe
	var readinessProbe core.Probe

	if pod.LivenessProbe != nil {
		livenessProbe = *pod.LivenessProbe
	} else {
		livenessProbe = core.Probe{}
	}
	if pod.ReadinessProbe != nil {
		readinessProbe = *pod.ReadinessProbe
	} else {
		readinessProbe = core.Probe{}
	}

	c := core.Container{
		Name:         nn.Name,
		Image:        pod.Image,
		Command:      pod.Command,
		Args:         pod.Args,
		Env:          envvar,
		Resources:    processResources(&pod, env),
		VolumeMounts: pod.VolumeMounts,
		Ports: []core.ContainerPort{{
			Name:          "metrics",
			ContainerPort: env.Spec.Providers.Metrics.Port,
		}},
		ImagePullPolicy: core.PullIfNotPresent,
	}

	if (core.Probe{}) != livenessProbe {
		c.LivenessProbe = &livenessProbe
	}
	if (core.Probe{}) != readinessProbe {
		c.ReadinessProbe = &readinessProbe
	}

	c.VolumeMounts = append(c.VolumeMounts, core.VolumeMount{
		Name:      "config-secret",
		MountPath: "/cdapp/",
	})

	cj.Spec.JobTemplate.Spec.Template.Spec.Containers = []core.Container{c}

	cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers = processInitContainers(nn, &c, pod.InitContainers)

	cj.Spec.JobTemplate.Spec.Template.Spec.Volumes = pod.Volumes
	cj.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cj.Spec.JobTemplate.Spec.Template.Spec.Volumes, core.Volume{
		Name: "config-secret",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: app.ObjectMeta.Name,
			},
		},
	})

	applyPodAntiAffinity(&cj.Spec.JobTemplate.Spec.Template)

	annotations := make(map[string]string)
	annotations["configHash"] = hash
	cj.Spec.JobTemplate.Spec.Template.SetAnnotations(annotations)
}

func processResources(pod *crd.PodSpec, env *crd.ClowdEnvironment) core.ResourceRequirements {
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

func processInitContainers(nn types.NamespacedName, c *core.Container, ics []crd.InitContainer) []core.Container {
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

// This should probably take arguments for addtional volumes, so that we can
// add those and then do one Apply
func (m *Maker) makeDeployment(deployment crd.Deployment, app *crd.ClowdApp, hash string) error {

	nn := types.NamespacedName{
		Name:      fmt.Sprintf("%v-%v", app.Name, deployment.Name),
		Namespace: m.Request.Namespace,
	}

	d := apps.Deployment{}
	err := m.Client.Get(m.Ctx, nn, &d)

	update, err := utils.UpdateOrErr(err)
	if err != nil {
		return err
	}

	initDeployment(m.App, m.Env, &d, nn, deployment, hash)

	if err := update.Apply(m.Ctx, m.Client, &d); err != nil {
		return err
	}

	return nil
}

func initDeployment(app *crd.ClowdApp, env *crd.ClowdEnvironment, d *apps.Deployment, nn types.NamespacedName, deployment crd.Deployment, hash string) {
	labels := app.GetLabels()
	labels["pod"] = nn.Name
	app.SetObjectMeta(d, crd.Name(nn.Name), crd.Labels(labels))

	pod := deployment.PodSpec

	d.Spec.Replicas = deployment.MinReplicas
	d.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	d.Spec.Template.ObjectMeta.Labels = labels
	d.Spec.Strategy = apps.DeploymentStrategy{
		Type: apps.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &apps.RollingUpdateDeployment{
			MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: int32(25)},
			MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: int32(25)},
		},
	}
	d.Spec.ProgressDeadlineSeconds = utils.Int32(600)

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
	} else if deployment.Web {
		livenessProbe = baseProbe
	}
	if pod.ReadinessProbe != nil {
		readinessProbe = *pod.ReadinessProbe
	} else {
		readinessProbe = baseProbe
		readinessProbe.InitialDelaySeconds = 45

	}

	c := core.Container{
		Name:         nn.Name,
		Image:        pod.Image,
		Command:      pod.Command,
		Args:         pod.Args,
		Env:          envvar,
		Resources:    processResources(&pod, env),
		VolumeMounts: pod.VolumeMounts,
		Ports: []core.ContainerPort{{
			Name:          "metrics",
			ContainerPort: env.Spec.Providers.Metrics.Port,
		}},
		ImagePullPolicy: core.PullIfNotPresent,
	}

	if (core.Probe{}) != livenessProbe {
		c.LivenessProbe = &livenessProbe
	}
	if (core.Probe{}) != readinessProbe {
		c.ReadinessProbe = &readinessProbe
	}

	if deployment.Web {
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

	d.Spec.Template.Spec.InitContainers = processInitContainers(nn, &c, pod.InitContainers)

	d.Spec.Template.Spec.Volumes = pod.Volumes
	d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, core.Volume{
		Name: "config-secret",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: app.ObjectMeta.Name,
			},
		},
	})

	applyPodAntiAffinity(&d.Spec.Template)

	annotations := make(map[string]string)
	annotations["configHash"] = hash
	d.Spec.Template.SetAnnotations(annotations)
}

func (m *Maker) runProviders() (*config.AppConfig, error) {
	provider := providers.Provider{
		Client: m.Client,
		Ctx:    m.Ctx,
		Env:    m.Env,
	}

	c := config.AppConfig{}

	c.WebPort = int(m.Env.Spec.Providers.Web.Port)
	c.MetricsPort = int(m.Env.Spec.Providers.Metrics.Port)
	c.MetricsPath = m.Env.Spec.Providers.Metrics.Path

	for _, provAcc := range providers.ProvidersRegistration.Registry {
		prov, err := provAcc.SetupProvider(&provider)
		if err != nil {
			return &c, errors.Wrap(fmt.Sprintf("getprov: %s", provAcc.Name), err)
		}
		err = prov.Provide(m.App, &c)
		if err != nil {
			reterr := errors.Wrap(fmt.Sprintf("runapp: %s", provAcc.Name), err)
			reterr.Requeue = true
			return &c, reterr
		}
	}

	return &c, nil
}

func (m *Maker) makeDependencies(c *config.AppConfig) error {

	// Return if no deps

	deps := m.App.Spec.Dependencies
	odeps := m.App.Spec.OptionalDependencies
	if (deps == nil || len(deps) == 0) && (odeps == nil || len(odeps) == 0) {
		return nil
	}

	// Get all ClowdApps

	apps := crd.ClowdAppList{}
	err := m.Client.List(m.Ctx, &apps)

	if err != nil {
		return errors.Wrap("Failed to list apps", err)
	}

	// Iterate over all deps

	depConfig, missingDeps := makeDepConfig(m.Env.Spec.Providers.Web.Port, m.App, &apps)

	if len(missingDeps) > 0 {
		depVal := map[string][]string{"services": missingDeps}
		return &errors.MissingDependencies{MissingDeps: depVal}
	}

	c.Endpoints = depConfig
	return nil
}

func makeDepConfig(webPort int32, app *crd.ClowdApp, apps *crd.ClowdAppList) (depConfig []config.DependencyEndpoint, missingDeps []string) {
	appMap := map[string]crd.ClowdApp{}

	for _, iapp := range apps.Items {

		if iapp.Spec.EnvName == app.Spec.EnvName {
			appMap[iapp.Name] = iapp
		}
	}

	depConfig = []config.DependencyEndpoint{}

	missingDeps = processAppEndpoints(appMap, app.Spec.Dependencies, &depConfig, webPort)
	_ = processAppEndpoints(appMap, app.Spec.OptionalDependencies, &depConfig, webPort)

	return depConfig, missingDeps
}

func processAppEndpoints(appMap map[string]crd.ClowdApp, depList []string, depConfig *[]config.DependencyEndpoint, webPort int32) (missingDeps []string) {

	missingDeps = []string{}

	for _, dep := range depList {
		depApp, exists := appMap[dep]
		if !exists {
			missingDeps = append(missingDeps, dep)
			continue
		}

		// If app has public endpoint, add it to app config

		for _, deployment := range depApp.Spec.Deployments {
			if deployment.Web {
				name := fmt.Sprintf("%s-%s", depApp.Name, deployment.Name)
				*depConfig = append(*depConfig, config.DependencyEndpoint{
					Hostname: fmt.Sprintf("%s.%s.svc", name, depApp.Namespace),
					Port:     int(webPort),
					Name:     deployment.Name,
					App:      depApp.Name,
				})
			}
		}
	}

	return missingDeps
}
