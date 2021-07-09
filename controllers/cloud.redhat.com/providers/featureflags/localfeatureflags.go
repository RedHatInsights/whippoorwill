package featureflags

import (
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	obj "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/object"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	provutils "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers/utils"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// LocalFFDeployment is the ident refering to the local Feature Flags deployment object.
var LocalFFDeployment = providers.NewSingleResourceIdent(ProvName, "ff_deployment", &apps.Deployment{})

// LocalFFService is the ident refering to the local Feature Flags service object.
var LocalFFService = providers.NewSingleResourceIdent(ProvName, "ff_service", &core.Service{})

// LocalFFDBDeployment is the ident refering to the local Feature Flags DB deployment object.
var LocalFFDBDeployment = providers.NewSingleResourceIdent(ProvName, "ff_db_deployment", &apps.Deployment{})

// LocalFFDBService is the ident refering to the local Feature Flags DB service object.
var LocalFFDBService = providers.NewSingleResourceIdent(ProvName, "ff_db_service", &core.Service{})

// LocalFFDBPVC is the ident refering to the local Feature Flags DB PVC object.
var LocalFFDBPVC = providers.NewSingleResourceIdent(ProvName, "ff_db_pvc", &core.PersistentVolumeClaim{})

// LocalFFDBSecret is the ident refering to the local Feature Flags DB secret object.
var LocalFFDBSecret = providers.NewSingleResourceIdent(ProvName, "ff_db_secret", &core.Secret{})

type localFeatureFlagsProvider struct {
	providers.Provider
	Config config.FeatureFlagsConfig
}

// NewLocalFeatureFlagsProvider returns a new local featureflags provider object.
func NewLocalFeatureFlagsProvider(p *providers.Provider) (providers.ClowderProvider, error) {

	ffp := &localFeatureFlagsProvider{Provider: *p, Config: config.FeatureFlagsConfig{}}

	objList := []providers.ResourceIdent{
		LocalFFDeployment,
		LocalFFService,
	}

	if err := providers.CachedMakeComponent(p.Cache, objList, p.Env, "featureflags", makeLocalFeatureFlags, false, p.Env.IsNodePort()); err != nil {
		return nil, err
	}

	nn := types.NamespacedName{
		Name:      "featureflags-db",
		Namespace: p.Env.Status.TargetNamespace,
	}

	dd := &apps.Deployment{}
	if err := p.Cache.Create(LocalFFDBDeployment, nn, dd); err != nil {
		return nil, err
	}

	dbCfg := config.DatabaseConfig{}
	dataInit := func() map[string]string {
		username := utils.RandString(16)
		password := utils.RandString(16)
		pgPass := utils.RandString(16)
		hostname := fmt.Sprintf("%v.%v.svc", nn.Name, nn.Namespace)
		connectionURL := fmt.Sprintf("postgres://%s:%s@%s/%s", username, password, hostname, "unleash")

		return map[string]string{
			"hostname":      hostname,
			"port":          "5432",
			"username":      username,
			"password":      password,
			"pgPass":        pgPass,
			"name":          "unleash",
			"connectionURL": connectionURL,
		}
	}

	secMap, err := providers.MakeOrGetSecret(p.Ctx, p.Env, p.Cache, LocalFFDBSecret, nn, dataInit)
	if err != nil {
		return nil, errors.Wrap("Couldn't set/get secret", err)
	}

	dbCfg.Populate(secMap)
	dbCfg.AdminUsername = "postgres"

	ffp.Config = config.FeatureFlagsConfig{
		Hostname: fmt.Sprintf("%s-featureflags.%s.svc", p.Env.Name, p.Env.Status.TargetNamespace),
		Port:     4242,
	}

	provutils.MakeLocalDB(dd, nn, p.Env, &dbCfg, "quay.io/cloudservices/postgresql-rds:12-1", p.Env.Spec.Providers.FeatureFlags.PVC, "unleash")

	if err = p.Cache.Update(LocalFFDBDeployment, dd); err != nil {
		return nil, err
	}

	s := &core.Service{}
	if err := p.Cache.Create(LocalFFDBService, nn, s); err != nil {
		return nil, err
	}

	provutils.MakeLocalDBService(s, nn, p.Env)

	if err = p.Cache.Update(LocalFFDBService, s); err != nil {
		return nil, err
	}

	if p.Env.Spec.Providers.FeatureFlags.PVC {
		pvc := &core.PersistentVolumeClaim{}
		if err = p.Cache.Create(LocalFFDBPVC, nn, pvc); err != nil {
			return nil, err
		}

		provutils.MakeLocalDBPVC(pvc, nn, p.Env)

		if err = p.Cache.Update(LocalFFDBPVC, pvc); err != nil {
			return nil, err
		}
	}

	return ffp, nil
}

// CreateDatabase ensures a database is created for the given app.  The
// namespaced name passed in must be the actual name of the db resources
func (ff *localFeatureFlagsProvider) Provide(app *crd.ClowdApp, c *config.AppConfig) error {
	c.FeatureFlags = &ff.Config
	return nil
}

func makeLocalFeatureFlags(o obj.ClowdObject, objMap providers.ObjectMap, usePVC bool, nodePort bool) {
	nn := providers.GetNamespacedName(o, "featureflags")

	dd := objMap[LocalFFDeployment].(*apps.Deployment)
	svc := objMap[LocalFFService].(*core.Service)

	labels := o.GetLabels()
	labels["env-app"] = nn.Name

	labeler := utils.MakeLabeler(nn, labels, o)

	labeler(dd)

	replicas := int32(1)

	dd.Spec.Replicas = &replicas
	dd.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}

	dd.Spec.Template.ObjectMeta.Labels = labels

	port := int32(4242)

	envVars := []core.EnvVar{{
		Name: "DATABASE_URL",
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: "featureflags-db",
				},
				Key: "connectionURL",
			},
		},
	},
	}

	ports := []core.ContainerPort{{
		Name:          "service",
		ContainerPort: port,
	}}

	probeHandler := core.Handler{
		TCPSocket: &core.TCPSocketAction{
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 4242,
			},
		},
	}

	livenessProbe := core.Probe{
		Handler:             probeHandler,
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
	}
	readinessProbe := core.Probe{
		Handler:             probeHandler,
		InitialDelaySeconds: 20,
		TimeoutSeconds:      2,
	}

	c := core.Container{
		Name:           nn.Name,
		Image:          "quay.io/cloudservices/unleash-docker:3.9",
		Env:            envVars,
		Ports:          ports,
		LivenessProbe:  &livenessProbe,
		ReadinessProbe: &readinessProbe,
	}

	dd.Spec.Template.Spec.Containers = []core.Container{c}
	dd.Spec.Template.SetLabels(labels)

	servicePorts := []core.ServicePort{{
		Name:     "featureflags",
		Port:     port,
		Protocol: "TCP",
	}}

	utils.MakeService(svc, nn, labels, servicePorts, o, nodePort)
}
