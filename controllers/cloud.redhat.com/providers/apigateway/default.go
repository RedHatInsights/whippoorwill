package apigateway

import (
	"crypto/sha256"
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	p "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

type apigatewayProvider struct {
	p.Provider
}

func NewApiGatewayProvider(p *p.Provider) (p.ClowderProvider, error) {
	if p.Env.Spec.Providers.ApiGateway.Mode != "enabled" {
		return &apigatewayProvider{Provider: *p}, nil
	}

	appList := crd.ClowdAppList{}
	p.Client.List(p.Ctx, &appList)

	configData := `:8080

    log
	
	`

	nn := types.NamespacedName{
		Namespace: p.Env.Status.TargetNamespace,
		Name:      "caddy",
	}

	for _, app := range appList.Items {
		if app.Spec.EnvName == p.Env.Name {
			for _, deployment := range app.Spec.Deployments {
				if deployment.WebServices.Public.Enabled {

					var endpoint string

					if deployment.WebServices.Public.ApiEndpoint == "" {
						endpoint = fmt.Sprintf("%s-%s", app.Name, deployment.Name)
					} else {
						endpoint = deployment.WebServices.Public.ApiEndpoint
					}
					configData += fmt.Sprintf(
						"reverse_proxy /api/%s/* %s-%s.%s.svc:%d\n",
						endpoint,
						app.Name,
						deployment.Name,
						app.Namespace,
						int(p.Env.Spec.Providers.Web.Port),
					)
				}
			}
		}
	}

	h := sha256.New()
	h.Write([]byte(configData))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	labels := p.Env.GetLabels()
	labels["env-app"] = nn.Name
	labeler := utils.MakeLabeler(nn, labels, p.Env)

	d := &apps.Deployment{}

	p.Cache.Create(ApiGatewayDeployment, nn, d)

	annotations := d.Spec.Template.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["configHash"] = hash

	d.Spec.Template.SetAnnotations(annotations)

	d.Spec.Template.ObjectMeta.Labels = labels
	d.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}

	d.Spec.Template.Spec.Containers = []core.Container{{
		Name:  "caddy",
		Image: "caddy",
		Ports: []core.ContainerPort{{
			Name:          "gateway",
			ContainerPort: 8080,
			Protocol:      "TCP",
		}},
		VolumeMounts: []core.VolumeMount{{
			Name:      "config",
			MountPath: "/etc/caddy/Caddyfile",
			SubPath:   "Caddyfile",
		}},
		ImagePullPolicy: "IfNotPresent",
	}}

	d.Spec.Template.Spec.Volumes = []core.Volume{{
		Name: "config",
		VolumeSource: core.VolumeSource{
			ConfigMap: &core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: "caddy",
				},
			},
		},
	}}

	labeler(d)

	p.Cache.Update(ApiGatewayDeployment, d)

	configMap := &core.ConfigMap{}

	p.Cache.Create(ApiGatewayConfig, nn, configMap)

	configMap.Data = map[string]string{"Caddyfile": configData}
	labeler(configMap)

	p.Cache.Update(ApiGatewayConfig, configMap)

	svc := &core.Service{}

	p.Cache.Create(ApiGatewayService, nn, svc)

	servicePorts := []core.ServicePort{{
		Name:     "gateway",
		Port:     8080,
		Protocol: "TCP",
	}}

	utils.MakeService(svc, nn, labels, servicePorts, p.Env, p.Env.IsNodePort())

	p.Cache.Update(ApiGatewayService, svc)

	return &apigatewayProvider{Provider: *p}, nil
}

func (j *apigatewayProvider) Provide(app *crd.ClowdApp, c *config.AppConfig) error {
	return nil
}
