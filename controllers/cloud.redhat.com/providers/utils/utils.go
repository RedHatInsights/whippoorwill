package providers

import (
	"cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1/common"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	obj "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/object"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// MakeLocalDB populates the given deployment object with the local DB struct.
func MakeLocalDB(dd *apps.Deployment, nn types.NamespacedName, baseResource obj.ClowdObject, cfg *config.DatabaseConfig, image string, usePVC bool, dbName string) {
	labels := baseResource.GetLabels()
	labels["service"] = "db"
	labler := utils.MakeLabeler(nn, labels, baseResource)
	labler(dd)

	var volSource core.VolumeSource
	if usePVC {
		volSource = core.VolumeSource{
			PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
				ClaimName: nn.Name,
			},
		}
	} else {
		volSource = core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		}
	}

	dd.Spec.Replicas = common.Int32Ptr(1)
	dd.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	dd.Spec.Template.Spec.Volumes = []core.Volume{
		{
			Name:         nn.Name,
			VolumeSource: volSource,
		},
	}
	dd.Spec.Template.ObjectMeta.Labels = labels

	envVars := []core.EnvVar{
		{Name: "POSTGRESQL_USER", Value: cfg.Username},
		{Name: "POSTGRESQL_PASSWORD", Value: cfg.Password},
		{Name: "PGPASSWORD", Value: cfg.AdminPassword}, // Legacy for old db images can likely be removed soon
		{Name: "POSTGRESQL_MASTER_USER", Value: cfg.AdminUsername},
		{Name: "POSTGRESQL_MASTER_PASSWORD", Value: cfg.AdminPassword},
		// TODO: Do we need to set the DB name?
		{Name: "POSTGRESQL_DATABASE", Value: dbName},
	}
	ports := []core.ContainerPort{{
		Name:          "database",
		ContainerPort: 5432,
	}}

	probeHandler := core.Handler{
		Exec: &core.ExecAction{
			Command: []string{
				"psql",
				"-U",
				"$(POSTGRESQL_USER)",
				"-d",
				"$(POSTGRESQL_DATABASE)",
				"-c",
				"SELECT 1",
			},
		},
	}

	livenessProbe := core.Probe{
		Handler:             probeHandler,
		InitialDelaySeconds: 15,
		TimeoutSeconds:      2,
	}
	readinessProbe := core.Probe{
		Handler:             probeHandler,
		InitialDelaySeconds: 45,
		TimeoutSeconds:      2,
	}

	c := core.Container{
		Name:           nn.Name,
		Image:          image,
		Env:            envVars,
		LivenessProbe:  &livenessProbe,
		ReadinessProbe: &readinessProbe,
		Ports:          ports,
		VolumeMounts: []core.VolumeMount{{
			Name:      nn.Name,
			MountPath: "/var/lib/pgsql/data",
		}},
	}

	dd.Spec.Template.Spec.Containers = []core.Container{c}
}

// MakeLocalDBService populates the given service object with the local DB struct.
func MakeLocalDBService(s *core.Service, nn types.NamespacedName, baseResource obj.ClowdObject) {
	servicePorts := []core.ServicePort{{
		Name:     "database",
		Port:     5432,
		Protocol: "TCP",
	}}
	utils.MakeService(s, nn, providers.Labels{"service": "db", "app": baseResource.GetClowdName()}, servicePorts, baseResource, false)
}

// MakeLocalDBPVC populates the given PVC object with the local DB struct.
func MakeLocalDBPVC(pvc *core.PersistentVolumeClaim, nn types.NamespacedName, baseResource obj.ClowdObject) {
	utils.MakePVC(pvc, nn, providers.Labels{"service": "db", "app": baseResource.GetClowdName()}, "1Gi", baseResource)
}
