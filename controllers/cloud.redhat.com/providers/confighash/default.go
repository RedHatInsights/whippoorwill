package confighash

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	p "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	cronjobProvider "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/cronjob"
	deployProvider "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/deployment"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type confighashProvider struct {
	p.Provider
}

// CoreConfigSecret is the config that is presented as the cdappconfig.json file.
var CoreConfigSecret = p.NewSingleResourceIdent(ProvName, "core_config_secret", &core.Secret{})

// NewConfigHashProvider returns a new End provider run at the end of the provider set.
func NewConfigHashProvider(p *p.Provider) (p.ClowderProvider, error) {
	return &confighashProvider{Provider: *p}, nil
}

func (ch *confighashProvider) Provide(app *crd.ClowdApp, c *config.AppConfig) error {

	hash, err := ch.persistConfig(app, c)

	if err != nil {
		return err
	}

	dList := apps.DeploymentList{}
	if err := ch.Cache.List(deployProvider.CoreDeployment, &dList); err != nil {
		return err
	}

	for _, deployment := range dList.Items {
		annotations := deployment.Spec.Template.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["configHash"] = hash

		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.ConfigMap != nil {
				cfgmap := &core.ConfigMap{}

				nn := types.NamespacedName{
					Name:      volume.ConfigMap.Name,
					Namespace: app.Namespace,
				}

				if err := ch.Client.Get(ch.Ctx, nn, cfgmap); err != nil {
					return errors.Wrap(fmt.Sprintf("%v - %v", nn, volume), err)
				}

				jsonData, err := json.Marshal(cfgmap.Data)
				if err != nil {
					return errors.Wrap("failed to marshal configmap JSON", err)
				}

				h := sha256.New()
				h.Write([]byte(jsonData))
				hash := fmt.Sprintf("%x", h.Sum(nil))

				annotations["clowderconfigmapdep_"+volume.ConfigMap.Name] = hash
			}
		}

		deployment.Spec.Template.SetAnnotations(annotations)

		ch.Cache.Update(deployProvider.CoreDeployment, &deployment)

	}

	jList := batch.CronJobList{}
	if err := ch.Cache.List(cronjobProvider.CoreCronJob, &jList); err != nil {
		return err
	}

	for _, job := range jList.Items {

		annotations := job.Spec.JobTemplate.Spec.Template.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["configHash"] = hash
		job.Spec.JobTemplate.Spec.Template.SetAnnotations(annotations)

		ch.Cache.Update(cronjobProvider.CoreCronJob, &job)
	}

	return nil
}
