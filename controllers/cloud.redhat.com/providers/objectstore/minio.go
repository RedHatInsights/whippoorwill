package objectstore

import (
	"context"
	"fmt"
	"strconv"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	obj "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/object"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MinioDeployment is the resource ident for the Minio deployment object.
var MinioDeployment = providers.NewSingleResourceIdent(ProvName, "minio_db_deployment", &apps.Deployment{})

// MinioService is the resource ident for the Minio service object.
var MinioService = providers.NewSingleResourceIdent(ProvName, "minio_db_service", &core.Service{})

// MinioPVC is the resource ident for the Minio PVC object.
var MinioPVC = providers.NewSingleResourceIdent(ProvName, "minio_db_pvc", &core.PersistentVolumeClaim{})

// MinioSecret is the resource ident for the Minio secret object.
var MinioSecret = providers.NewSingleResourceIdent(ProvName, "minio_db_secret", &core.Secret{})

// MinioNetworkPolicy is the resource ident for the KafkaNetworkPolicy
var MinioNetworkPolicy = providers.NewSingleResourceIdent(ProvName, "minio_network_policy", &networking.NetworkPolicy{})

const bucketCheckErrorMsg = "failed to check if bucket exists"
const bucketCreateErrorMsg = "failed to create bucket"

func newBucketError(msg string, bucketName string, rootCause error) error {
	newErr := errors.Wrap(fmt.Sprintf("bucket %q -- %s", bucketName, msg), rootCause)
	newErr.Requeue = true
	return newErr
}

// Create a bucketHandler interface to allow for mocking of minio client actions in tests
type bucketHandler interface {
	Exists(ctx context.Context, bucketName string) (bool, error)
	Make(ctx context.Context, bucketName string) error
	CreateClient(hostname string, port int, accessKey *string, secretKey *string) error
}

// minioHandler will implement the above interface using minio-go
type minioHandler struct {
	Client *minio.Client
}

func (h *minioHandler) Exists(ctx context.Context, bucketName string) (bool, error) {
	return h.Client.BucketExists(ctx, bucketName)
}

func (h *minioHandler) Make(ctx context.Context, bucketName string) error {
	return h.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

func (h *minioHandler) CreateClient(
	hostname string, port int, accessKey *string, secretKey *string,
) error {
	endpoint := fmt.Sprintf("%v:%v", hostname, port)

	cl, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(*accessKey, *secretKey, ""),
		Secure: false,
	})

	if err != nil {
		return errors.Wrap("Failed to create minio client", err)
	}

	h.Client = cl

	return nil
}

// minio is an object store provider that deploys and configures MinIO
type minioProvider struct {
	providers.Provider
	Config        config.ObjectStoreConfig
	BucketHandler bucketHandler
}

// Provide creates new buckets
func (m *minioProvider) Provide(app *crd.ClowdApp, c *config.AppConfig) error {
	if len(app.Spec.ObjectStore) == 0 {
		return nil
	}

	for _, bucket := range app.Spec.ObjectStore {
		found, err := m.BucketHandler.Exists(m.Ctx, bucket)

		if err != nil {
			return newBucketError(bucketCheckErrorMsg, bucket, err)
		}

		if !found {
			err = m.BucketHandler.Make(m.Ctx, bucket)

			if err != nil {
				return newBucketError(bucketCreateErrorMsg, bucket, err)
			}
		}

		m.Config.Buckets = append(m.Config.Buckets, config.ObjectStoreBucket{
			Name:          bucket,
			RequestedName: bucket,
			AccessKey:     m.Config.AccessKey,
			SecretKey:     m.Config.SecretKey,
		})
	}
	c.ObjectStore = &config.ObjectStoreConfig{
		Hostname:  m.Config.Hostname,
		Port:      m.Config.Port,
		AccessKey: m.Config.AccessKey,
		SecretKey: m.Config.SecretKey,
		Buckets:   m.Config.Buckets,
		Tls:       false,
	}
	return nil
}

func createMinioProvider(
	p *providers.Provider, secMap map[string]string, handler bucketHandler,
) (*minioProvider, error) {
	mp := &minioProvider{Provider: *p, Config: config.ObjectStoreConfig{}}

	port, _ := strconv.Atoi(secMap["port"])
	mp.Ctx = p.Ctx
	mp.Config.AccessKey = providers.StrPtr(secMap["accessKey"])
	mp.Config.SecretKey = providers.StrPtr(secMap["secretKey"])
	mp.Config.Hostname = secMap["hostname"]
	mp.Config.Port = port

	mp.BucketHandler = handler
	err := mp.BucketHandler.CreateClient(
		mp.Config.Hostname,
		mp.Config.Port,
		mp.Config.AccessKey,
		mp.Config.SecretKey,
	)

	if err != nil {
		return nil, errors.Wrap("error creating minio client", err)
	}
	return mp, nil
}

func createDefaultMinioSecMap(name string, namespace string) map[string]string {
	return map[string]string{
		"accessKey": utils.RandString(12),
		"secretKey": utils.RandString(12),
		"hostname":  fmt.Sprintf("%v.%v.svc", name, namespace),
		"port":      strconv.Itoa(int(9000)),
	}
}

// NewMinIO constructs a new minio for the given config
func NewMinIO(p *providers.Provider) (providers.ClowderProvider, error) {
	nn := providers.GetNamespacedName(p.Env, "minio")

	dataInit := func() map[string]string {
		return createDefaultMinioSecMap(nn.Name, nn.Namespace)
	}
	// MakeOrGetSecret will set data if it already exists
	secMap, err := providers.MakeOrGetSecret(p.Ctx, p.Env, p.Cache, MinioSecret, nn, dataInit)
	if err != nil {
		raisedErr := errors.Wrap("Couldn't set/get secret", err)
		raisedErr.Requeue = true
		return nil, raisedErr
	}

	mp, err := createMinioProvider(p, *secMap, &minioHandler{})

	if err != nil {
		return nil, err
	}

	minioCacheMap := []providers.ResourceIdent{
		MinioDeployment,
		MinioService,
	}

	if p.Env.Spec.Providers.ObjectStore.PVC {
		minioCacheMap = append(minioCacheMap, MinioPVC)
	}

	err = providers.CachedMakeComponent(p.Cache, minioCacheMap, p.Env, "minio", makeLocalMinIO, p.Env.Spec.Providers.ObjectStore.PVC, p.Env.IsNodePort())

	if err != nil {
		raisedErr := errors.Wrap("Couldn't make component", err)
		raisedErr.Requeue = true
		return nil, raisedErr
	}

	return mp, createNetworkPolicy(p)
}

func createNetworkPolicy(p *providers.Provider) error {
	clowderNs, err := utils.GetClowderNamespace()

	if err != nil {
		return nil
	}

	np := &networking.NetworkPolicy{}
	nn := types.NamespacedName{
		Name:      fmt.Sprintf("allow-from-%s-namespace", clowderNs),
		Namespace: p.Env.Status.TargetNamespace,
	}

	if err := p.Cache.Create(MinioNetworkPolicy, nn, np); err != nil {
		return err
	}

	npFrom := []networking.NetworkPolicyPeer{{
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"kubernetes.io/metadata.name": clowderNs,
			},
		},
	}}

	np.Spec.Ingress = []networking.NetworkPolicyIngressRule{{
		From: npFrom,
	}}

	np.Spec.PolicyTypes = []networking.PolicyType{"Ingress"}

	labeler := utils.GetCustomLabeler(nil, nn, p.Env)
	labeler(np)

	if err := p.Cache.Update(MinioNetworkPolicy, np); err != nil {
		return err
	}

	return nil
}

func makeLocalMinIO(o obj.ClowdObject, objMap providers.ObjectMap, usePVC bool, nodePort bool) {
	nn := providers.GetNamespacedName(o, "minio")

	dd := objMap[MinioDeployment].(*apps.Deployment)
	svc := objMap[MinioService].(*core.Service)

	labels := o.GetLabels()
	labels["env-app"] = nn.Name

	labeler := utils.MakeLabeler(nn, labels, o)

	labeler(dd)

	replicas := int32(1)

	dd.Spec.Replicas = &replicas
	dd.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}

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

	dd.Spec.Template.Spec.Volumes = []core.Volume{
		{
			Name:         nn.Name,
			VolumeSource: volSource,
		},
	}
	dd.Spec.Template.ObjectMeta.Labels = labels

	port := int32(9000)

	envVars := []core.EnvVar{{
		Name: "MINIO_ACCESS_KEY",
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: nn.Name,
				},
				Key: "accessKey",
			},
		},
	}, {
		Name: "MINIO_SECRET_KEY",
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: nn.Name,
				},
				Key: "secretKey",
			},
		},
	}}

	ports := []core.ContainerPort{{
		Name:          "minio",
		ContainerPort: port,
	}}

	probeHandler := core.Handler{
		TCPSocket: &core.TCPSocketAction{
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 9000,
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
		Name:  nn.Name,
		Image: "quay.io/cloudservices/minio:RELEASE.2020-11-19T23-48-16Z-amd64",
		Env:   envVars,
		Ports: ports,
		VolumeMounts: []core.VolumeMount{{
			Name:      nn.Name,
			MountPath: "/storage",
		}},
		Args: []string{
			"server",
			"/storage",
		},
		LivenessProbe:  &livenessProbe,
		ReadinessProbe: &readinessProbe,
	}

	dd.Spec.Template.Spec.Containers = []core.Container{c}
	dd.Spec.Template.SetLabels(labels)

	servicePorts := []core.ServicePort{{
		Name:     "minio",
		Port:     port,
		Protocol: "TCP",
	}}

	utils.MakeService(svc, nn, labels, servicePorts, o, nodePort)
	if usePVC {
		pvc := objMap[MinioPVC].(*core.PersistentVolumeClaim)
		utils.MakePVC(pvc, nn, labels, "1Gi", o)
	}
}
