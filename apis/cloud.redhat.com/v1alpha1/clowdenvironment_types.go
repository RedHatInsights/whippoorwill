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

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1/common"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"
	strimzi "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta1"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WebMode details the mode of operation of the Clowder Web Provider
// +kubebuilder:validation:Enum=none;operator
type WebMode string

// WebConfig configures the Clowder provider controlling the creation of web
// services and their probes.
type WebConfig struct {
	// The port that web services inside ClowdApp pods should be served on.
	Port int32 `json:"port"`

	// The private port that web services inside a ClowdApp should be served on.
	PrivatePort int32 `json:"privatePort,omitempty"`

	// An api prefix path that pods will be instructed to use when setting up
	// their web server.
	ApiPrefix string `json:"apiPrefix,omitempty"`

	// The mode of operation of the Web provider. The allowed modes are
	// (*_none_*), which disables web service generation, or (*_operator_*)
	// where services and probes are generated.
	Mode WebMode `json:"mode"`
}

// MetricsMode details the mode of operation of the Clowder Metrics Provider
// +kubebuilder:validation:Enum=none;operator;app-interface
type MetricsMode string

type PrometheusConfig struct {
	// Determines whether to deploy prometheus in operator mode
	Deploy bool `json:"deploy,omitempty"`
}

// MetricsConfig configures the Clowder provider controlling the creation of
// metrics services and their probes.
type MetricsConfig struct {
	// The port that metrics services inside ClowdApp pods should be served on.
	Port int32 `json:"port"`

	// A prefix path that pods will be instructed to use when setting up their
	// metrics server.
	Path string `json:"path,omitempty"`

	// The mode of operation of the Metrics provider. The allowed modes are
	//  (*_none_*), which disables metrics service generation, or
	// (*_operator_*) where services and probes are generated.
	// (*_app-interface_*) where services and probes are generated for app-interface.
	Mode MetricsMode `json:"mode"`

	// Prometheus specific configuration
	Prometheus PrometheusConfig `json:"prometheus,omitempty"`
}

// TODO: Other potential mode: saas

// KafkaMode details the mode of operation of the Clowder Kafka Provider
// +kubebuilder:validation:Enum=managed;operator;app-interface;local;none
type KafkaMode string

// KafkaClusterConfig defines options related to the Kafka cluster managed/monitored by Clowder
type KafkaClusterConfig struct {
	// Defines the kafka cluster name (default: name of ClowdEnvironment)
	Name string `json:"name,omitempty"`

	// The namespace the kafka cluster is expected to reside in (default: the environment's targetNamespace)
	Namespace string `json:"namespace,omitempty"`

	// The requested number of replicas for kafka/zookeeper. If unset, default is '1'
	// +kubebuilder:validation:Minimum:=1
	Replicas int32 `json:"replicas,omitempty"`

	// Persistent volume storage size. If unset, default is '1Gi'
	// Only applies when KafkaConfig.PVC is set to 'true'
	StorageSize string `json:"storageSize,omitempty"`

	// Delete persistent volume claim if the Kafka cluster is deleted
	// Only applies when KafkaConfig.PVC is set to 'true'
	DeleteClaim bool `json:"deleteClaim,omitempty"`

	// Version. If unset, default is '2.5.0'
	Version string `json:"version,omitempty"`

	// Config full options
	Config strimzi.KafkaSpecKafkaConfig `json:"config,omitempty"`

	// JVM Options
	JVMOptions strimzi.KafkaSpecKafkaJvmOptions `json:"jvmOptions,omitempty"`

	// Resource Limits
	Resources strimzi.KafkaSpecKafkaResources `json:"resources,omitempty"`
}

// KafkaConnectClusterConfig defines options related to the Kafka Connect cluster managed/monitored by Clowder
type KafkaConnectClusterConfig struct {
	// Defines the kafka connect cluster name (default: '<kafka cluster's name>-connect')
	Name string `json:"name,omitempty"`

	// The namespace the kafka connect cluster is expected to reside in (default: the kafka cluster's namespace)
	Namespace string `json:"namespace,omitempty"`

	// The requested number of replicas for kafka connect. If unset, default is '1'
	// +kubebuilder:validation:Minimum:=1
	Replicas int32 `json:"replicas,omitempty"`

	// Version. If unset, default is '2.5.0'
	Version string `json:"version,omitempty"`

	// Image. If unset, default is 'quay.io/cloudservices/xjoin-kafka-connect-strimzi:latest'
	Image string `json:"image,omitempty"`
}

// NamespacedName type to represent a real Namespaced Name
type NamespacedName struct {
	// Name defines the Name of a resource.
	Name string `json:"name"`

	// Namespace defines the Namespace of a resource.
	Namespace string `json:"namespace"`
}

// KafkaConfig configures the Clowder provider controlling the creation of
// Kafka instances.
type KafkaConfig struct {
	// The mode of operation of the Clowder Kafka Provider. Valid options are:
	// (*_operator_*) which provisions Strimzi resources and will configure
	// KafkaTopic CRs and place them in the Kafka cluster's namespace described in the configuration,
	// (*_app-interface_*) which simply passes the topic names through to the App's
	// cdappconfig.json and expects app-interface to have created the relevant
	// topics, and (*_local_*) where a small instance of Kafka is created in the desired cluster namespace
	// and configured to auto-create topics.
	Mode KafkaMode `json:"mode"`

	// EnableLegacyStrimzi disables TLS + user auth
	EnableLegacyStrimzi bool `json:"enableLegacyStrimzi,omitempty"`

	// If using the (*_local_*) or (*_operator_*) mode and PVC is set to true, this sets the provisioned
	// Kafka instance to use a PVC instead of emptyDir for its volumes.
	PVC bool `json:"pvc,omitempty"`

	// Defines options related to the Kafka cluster for this environment. Ignored for (*_local_*) mode.
	Cluster KafkaClusterConfig `json:"cluster,omitempty"`

	// Defines options related to the Kafka Connect cluster for this environment. Ignored for (*_local_*) mode.
	Connect KafkaConnectClusterConfig `json:"connect,omitempty"`

	// Defines the secret reference for the Managed Kafka mode. Only used in (*_managed_*) mode.
	ManagedSecretRef NamespacedName `json:"managedSecretRef,omitempty"`

	// (Deprecated) Defines the cluster name to be used by the Kafka Provider this will
	// be used in some modes to locate the Kafka instance.
	ClusterName string `json:"clusterName,omitempty"`

	// (Deprecated) The Namespace the cluster is expected to reside in. This is only used
	// in (*_app-interface_*) and (*_operator_*) modes.
	Namespace string `json:"namespace,omitempty"`

	// (Deprecated) The namespace that the Kafka Connect cluster is expected to reside in. This is only used
	// in (*_app-interface_*) and (*_operator_*) modes.
	ConnectNamespace string `json:"connectNamespace,omitempty"`

	// (Deprecated) Defines the kafka connect cluster name that is used in this environment.
	ConnectClusterName string `json:"connectClusterName,omitempty"`

	// (Deprecated) (Unused)
	Suffix string `json:"suffix,omitempty"`
}

// TODO: Other potential modes: RDS and Operator (e.g. CrunchyDB)

// DatabaseMode details the mode of operation of the Clowder Database Provider
// +kubebuilder:validation:Enum=app-interface;local;none
type DatabaseMode string

// DatabaseConfig configures the Clowder provider controlling the creation of
// Database instances.
type DatabaseConfig struct {
	// The mode of operation of the Clowder Database Provider. Valid options are:
	// (*_app-interface_*) where the provider will pass through database credentials
	// found in the secret defined by the database name in the ClowdApp, and (*_local_*)
	// where the provider will spin up a local instance of the database.
	Mode DatabaseMode `json:"mode"`

	// If using the (*_local_*) mode and PVC is set to true, this instructs the local
	// Database instance to use a PVC instead of emptyDir for its volumes.
	PVC bool `json:"pvc,omitempty"`
}

// TODO: Other potential modes: splunk, kafka

// LoggingMode details the mode of operation of the Clowder Logging Provider
// +kubebuilder:validation:Enum=app-interface;null;none
type LoggingMode string

// LoggingConfig configures the Clowder provider controlling the creation of
// Logging instances.
type LoggingConfig struct {
	// The mode of operation of the Clowder Logging Provider. Valid options are:
	// (*_app-interface_*) where the provider will pass through cloudwatch credentials
	// to the app configuration, and (*_none_*) where no logging will be configured.
	Mode LoggingMode `json:"mode"`
}

// ServiceMeshMode just determines if we enable or disable the service mesh
// +kubebuilder:validation:Enum=enabled;disabled
type ServiceMeshMode string

// ServiceMeshConfig determines if this env should be part of a service mesh
// and, if enabled, configures the service mesh
type ServiceMeshConfig struct {
	Mode ServiceMeshMode `json:"mode,omitempty"`
}

// TODO: Other potential mode: ceph, S3

// ObjectStoreMode details the mode of operation of the Clowder ObjectStore
// Provider
// +kubebuilder:validation:Enum=minio;app-interface;none
type ObjectStoreMode string

// ObjectStoreConfig configures the Clowder provider controlling the creation of
// ObjectStore instances.
type ObjectStoreConfig struct {
	// The mode of operation of the Clowder ObjectStore Provider. Valid options are:
	// (*_app-interface_*) where the provider will pass through Amazon S3 credentials
	// to the app configuration, and (*_minio_*) where a local Minio instance will
	// be created.
	Mode ObjectStoreMode `json:"mode"`

	// Currently unused.
	Suffix string `json:"suffix,omitempty"`

	// If using the (*_local_*) mode and PVC is set to true, this instructs the local
	// Database instance to use a PVC instead of emptyDir for its volumes.
	PVC bool `json:"pvc,omitempty"`
}

// FeatureFlagsMode details the mode of operation of the Clowder FeatureFlags
// Provider
// +kubebuilder:validation:Enum=local;app-interface;none
// +kubebuilder:validation:Optional
type FeatureFlagsMode string

// FeatureFlagsConfig configures the Clowder provider controlling the creation of
// FeatureFlag instances.
type FeatureFlagsConfig struct {
	// The mode of operation of the Clowder FeatureFlag Provider. Valid options are:
	// (*_app-interface_*) where the provider will pass through credentials
	// to the app configuration, and (*_local_*) where a local Unleash instance will
	// be created.
	Mode FeatureFlagsMode `json:"mode,omitempty"`

	// If using the (*_local_*) mode and PVC is set to true, this instructs the local
	// Database instance to use a PVC instead of emptyDir for its volumes.
	PVC bool `json:"pvc,omitempty"`

	// Defines the secret containing the client access token, only used for (*_app-interface_*)
	// mode.
	CredentialRef NamespacedName `json:"credentialRef,omitempty"`

	// Defines the hostname for (*_app-interface_*) mode
	Hostname string `json:"hostname,omitempty"`

	// Defineds the port for (*_app-interface_*) mode
	Port int32 `json:"port,omitempty"`
}

// InMemoryMode details the mode of operation of the Clowder InMemoryDB
// Provider
// +kubebuilder:validation:Enum=redis;app-interface;elasticache;none
type InMemoryMode string

// InMemoryDBConfig configures the Clowder provider controlling the creation of
// InMemoryDB instances.
type InMemoryDBConfig struct {
	// The mode of operation of the Clowder InMemory Provider. Valid options are:
	// (*_redis_*) where a local Minio instance will be created, and (*_elasticache_*)
	// which will search the namespace of the ClowdApp for a secret called 'elasticache'
	Mode InMemoryMode `json:"mode"`

	// If using the (*_local_*) mode and PVC is set to true, this instructs the local
	// Database instance to use a PVC instead of emptyDir for its volumes.
	PVC bool `json:"pvc,omitempty"`
}

// Describes what amount of app config is mounted to the pod
// +kubebuilder:validation:Enum={"none", "app", "", "environment"}
type ConfigAccessMode string

type TestingConfig struct {
	// Defines the environment for iqe/smoke testing
	Iqe IqeConfig `json:"iqe,omitempty"`

	// The mode of operation of the testing Pod. Valid options are:
	// 'default', 'view' or 'edit'
	K8SAccessLevel K8sAccessLevel `json:"k8sAccessLevel"`

	// The mode of operation for access to outside app configs. Valid
	// options are:
	// (*_none_*) -- no app config is mounted to the pod
	// (*_app_*) -- only the ClowdApp's config is mounted to the pod
	// (*_environment_*) -- the config for all apps in the env are mounted
	ConfigAccess ConfigAccessMode `json:"configAccess"`
}

type IqeConfig struct {
	ImageBase string `json:"imageBase"`

	// A pass-through of a resource requirements in k8s ResourceRequirements
	// format. If omitted, the default resource requirements from the
	// ClowdEnvironment will be used.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Defines the secret reference for loading vault credentials into the IQE job
	VaultSecretRef NamespacedName `json:"vaultSecretRef,omitempty"`
}

// ServiceConfig provides options for k8s Service resources
type ServiceConfig struct {
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;""
	Type string `json:"type"`
}

// ClowdEnvironmentSpec defines the desired state of ClowdEnvironment.
type ClowdEnvironmentSpec struct {
	// TargetNamespace describes the namespace where any generated environmental
	// resources should end up, this is particularly important in (*_local_*) mode.
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// A ProvidersConfig object, detailing the setup and configuration of all the
	// providers used in this ClowdEnvironment.
	Providers ProvidersConfig `json:"providers"`

	// Defines the default resource requirements in standard k8s format in the
	// event that they omitted from a PodSpec inside a ClowdApp.
	ResourceDefaults v1.ResourceRequirements `json:"resourceDefaults"`

	ServiceConfig ServiceConfig `json:"serviceConfig,omitempty"`
}

// ProvidersConfig defines a group of providers configuration for a ClowdEnvironment.
type ProvidersConfig struct {
	// Defines the Configuration for the Clowder Database Provider.
	Database DatabaseConfig `json:"db,omitempty"`

	// Defines the Configuration for the Clowder InMemoryDB Provider.
	InMemoryDB InMemoryDBConfig `json:"inMemoryDb"`

	// Defines the Configuration for the Clowder Kafka Provider.
	Kafka KafkaConfig `json:"kafka"`

	// Defines the Configuration for the Clowder Logging Provider.
	Logging LoggingConfig `json:"logging"`

	// Defines the Configuration for the Clowder Metrics Provider.
	Metrics MetricsConfig `json:"metrics,omitempty"`

	// Defines the Configuration for the Clowder ObjectStore Provider.
	ObjectStore ObjectStoreConfig `json:"objectStore"`

	// Defines the Configuration for the Clowder Web Provider.
	Web WebConfig `json:"web,omitempty"`

	// Defines the Configuration for the Clowder FeatureFlags Provider.
	FeatureFlags FeatureFlagsConfig `json:"featureFlags,omitempty"`

	// Defines the Configuration for the Clowder ServiceMesh Provider.
	ServiceMesh ServiceMeshConfig `json:"serviceMesh,omitempty"`

	// Defines the pull secret to use for the service accounts.
	PullSecrets []NamespacedName `json:"pullSecrets,omitempty"`

	// Defines the environment for iqe/smoke testing
	Testing TestingConfig `json:"testing,omitempty"`
}

// MinioStatus defines the status of a minio instance in local mode.
type MinioStatus struct {
	// A reference to standard k8s secret.
	Credentials core.SecretReference `json:"credentials"`

	// The hostname of a Minio instance.
	Hostname string `json:"hostname"`

	// The port number the Minio instance is to be served on.
	Port int32 `json:"port"`
}

// ClowdEnvironmentStatus defines the observed state of ClowdEnvironment
type ClowdEnvironmentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions      []ClowdCondition        `json:"conditions,omitempty"`
	TargetNamespace string                  `json:"targetNamespace,omitempty"`
	Ready           bool                    `json:"ready,omitempty"`
	Deployments     common.DeploymentStatus `json:"deployments,omitempty"`
	Apps            []AppInfo               `json:"apps,omitempty"`
	Generation      int64                   `json:"generation,omitempty"`
}

// AppInfo details information about a specific app.
type AppInfo struct {
	Name        string           `json:"name"`
	Deployments []DeploymentInfo `json:"deployments"`
}

// DeploymentInfo defailts information about a specific deployment.
type DeploymentInfo struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname,omitempty"`
	Port     int32  `json:"port,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=env
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.deployments.readyDeployments"
// +kubebuilder:printcolumn:name="Managed",type="integer",JSONPath=".status.deployments.managedDeployments"
// +kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".status.targetNamespace"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ClowdEnvironment is the Schema for the clowdenvironments API
type ClowdEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// A ClowdEnvironmentSpec object.
	Spec   ClowdEnvironmentSpec   `json:"spec,omitempty"`
	Status ClowdEnvironmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// ClowdEnvironmentList contains a list of ClowdEnvironment
type ClowdEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// A list of ClowdEnvironment objects.
	Items []ClowdEnvironment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClowdEnvironment{}, &ClowdEnvironmentList{})
}

// GetLabels returns a base set of labels relating to the ClowdEnvironment.
func (i *ClowdEnvironment) GetLabels() map[string]string {
	return map[string]string{
		"app": i.ObjectMeta.Name,
	}
}

// MakeOwnerReference defines the owner reference pointing to the ClowdApp resource.
func (i *ClowdEnvironment) MakeOwnerReference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: i.APIVersion,
		Kind:       i.Kind,
		Name:       i.ObjectMeta.Name,
		UID:        i.ObjectMeta.UID,
		Controller: common.TruePtr(),
	}
}

// GetClowdNamespace returns the namespace of the ClowdApp object.
func (i *ClowdEnvironment) GetClowdNamespace() string {
	return i.Status.TargetNamespace
}

// GetClowdName returns the name of the ClowdApp object.
func (i *ClowdEnvironment) GetClowdName() string {
	return i.Name
}

// GetPrimaryLabel returns the primary label name use for igentification.
func (i *ClowdEnvironment) GetPrimaryLabel() string {
	return "env"
}

// GetClowdSAName returns the ServiceAccount Name for the App
func (i *ClowdEnvironment) GetClowdSAName() string {
	return fmt.Sprintf("%s-env", i.GetClowdName())
}

// GetUID returns ObjectMeta.UID
func (i *ClowdEnvironment) GetUID() types.UID {
	return i.ObjectMeta.UID
}

// GetDeploymentStatus returns the Status.Deployments member
func (i *ClowdEnvironment) GetDeploymentStatus() *common.DeploymentStatus {
	return &i.Status.Deployments
}

// GenerateTargetNamespace gets a generated target namespace if one is not provided
func (i *ClowdEnvironment) GenerateTargetNamespace() string {
	return fmt.Sprintf("clowdenv-%s-%s", i.Name, strings.ToLower(utils.RandString(6)))
}

// IsReady returns true when all the ManagedDeployments are Ready
func (i *ClowdEnvironment) IsReady() bool {
	return (i.Status.Deployments.ManagedDeployments == i.Status.Deployments.ReadyDeployments)
}

// ConvertDeprecatedKafkaSpec converts values from the old Kafka provider spec into the new format
func (i *ClowdEnvironment) ConvertDeprecatedKafkaSpec() {
	if i.Spec.Providers.Kafka.ClusterName != "" {
		i.Spec.Providers.Kafka.Cluster.Name = i.Spec.Providers.Kafka.ClusterName
	}

	if i.Spec.Providers.Kafka.Namespace != "" {
		i.Spec.Providers.Kafka.Cluster.Namespace = i.Spec.Providers.Kafka.Namespace
	}

	if i.Spec.Providers.Kafka.ConnectNamespace != "" {
		i.Spec.Providers.Kafka.Connect.Namespace = i.Spec.Providers.Kafka.ConnectNamespace
	}

	if i.Spec.Providers.Kafka.ConnectClusterName != "" {
		i.Spec.Providers.Kafka.Connect.Name = i.Spec.Providers.Kafka.ConnectClusterName
	}
}

// GetAppsInEnv populates the appList with a list of all apps in the ClowdEnvironment.
func (i *ClowdEnvironment) GetAppsInEnv(ctx context.Context, pClient client.Client) (*ClowdAppList, error) {

	appList := &ClowdAppList{}

	err := pClient.List(ctx, appList, client.MatchingFields{"spec.envName": i.Name})

	if err != nil {
		return appList, errors.Wrap("could not list apps", err)
	}

	return appList, nil
}

// GetAppsInEnv populates the appList with a list of all apps in the ClowdEnvironment.
func (i *ClowdEnvironment) GetNamespacesInEnv(ctx context.Context, pClient client.Client) ([]string, error) {

	var err error
	var appList *ClowdAppList

	if appList, err = i.GetAppsInEnv(ctx, pClient); err != nil {
		return nil, err
	}

	tmpNamespace := map[string]bool{}

	for _, app := range appList.Items {
		tmpNamespace[app.Namespace] = true
	}

	namespaceList := []string{}

	for namespace, _ := range tmpNamespace {
		namespaceList = append(namespaceList, namespace)
	}

	return namespaceList, nil
}

// IsNodePort indicates whether or not services are configured as NodePort or not
func (i *ClowdEnvironment) IsNodePort() bool {
	return i.Spec.ServiceConfig.Type == "NodePort"
}
