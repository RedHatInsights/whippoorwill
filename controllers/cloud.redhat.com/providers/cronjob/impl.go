package cronjob

import (
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	deployProvider "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/deployment"

	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetCronJobName(app *crd.ClowdApp, cronjob *crd.Job) string {
	return fmt.Sprintf("%s-%s", app.Name, cronjob.Name)
}

func (j *cronjobProvider) makeCronJob(cronjob *crd.Job, app *crd.ClowdApp) error {

	nn := types.NamespacedName{
		Name:      GetCronJobName(app, cronjob),
		Namespace: app.Namespace,
	}

	pt := core.PodTemplateSpec{}
	buildPodTemplate(app, j.Env, &pt, nn, cronjob)

	c := &batch.CronJob{}
	if err := j.Cache.Create(CoreCronJob, nn, c); err != nil {
		return err
	}

	applyCronCronJob(app, j.Env, c, &pt, nn, cronjob)

	if err := j.Cache.Update(CoreCronJob, c); err != nil {
		return err
	}
	return nil
}

func buildPodTemplate(app *crd.ClowdApp, env *crd.ClowdEnvironment, pt *core.PodTemplateSpec, nn types.NamespacedName, cronjob *crd.Job) {
	labels := app.GetLabels()
	labels["pod"] = nn.Name

	pod := cronjob.PodSpec

	pt.ObjectMeta.Labels = labels

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
		Resources:    deployProvider.ProcessResources(&pod, env),
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

	pt.Spec.Containers = []core.Container{c}

	pt.Spec.InitContainers = deployProvider.ProcessInitContainers(nn, &c, pod.InitContainers)

	pt.Spec.Volumes = pod.Volumes
	pt.Spec.Volumes = append(pt.Spec.Volumes, core.Volume{
		Name: "config-secret",
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: app.ObjectMeta.Name,
			},
		},
	})

	deployProvider.ApplyPodAntiAffinity(pt)
}

func applyCronCronJob(app *crd.ClowdApp, env *crd.ClowdEnvironment, cj *batch.CronJob, pt *core.PodTemplateSpec, nn types.NamespacedName, cronjob *crd.Job) {
	labels := app.GetLabels()
	labels["pod"] = nn.Name
	app.SetObjectMeta(cj, crd.Name(nn.Name), crd.Labels(labels))

	cj.Spec.Schedule = cronjob.Schedule

	cj.Spec.JobTemplate.ObjectMeta.Labels = labels
	cj.Spec.JobTemplate.Spec.Template = *pt

	if cronjob.ConcurrencyPolicy == "" {
		cj.Spec.ConcurrencyPolicy = batch.AllowConcurrent
	} else {
		cj.Spec.ConcurrencyPolicy = cronjob.ConcurrencyPolicy
	}

	if cronjob.StartingDeadlineSeconds != nil {
		cj.Spec.StartingDeadlineSeconds = cronjob.StartingDeadlineSeconds
	}

	cj.Spec.JobTemplate.Spec.Template.Annotations = make(map[string]string)

	if cronjob.RestartPolicy == "" {
		cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = core.RestartPolicyNever
	} else {
		cj.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = cronjob.RestartPolicy
	}

	deployProvider.ApplyPodAntiAffinity(&cj.Spec.JobTemplate.Spec.Template)

}
