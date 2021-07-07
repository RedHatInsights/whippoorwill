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
	"errors"
	"fmt"

	"cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1/common"
	strimzi "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta1"
	batch "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InitContainer is a struct defining a k8s init container. This will be
// deployed along with the parent pod and is used to carry out one time
// initialization procedures.
type InitContainer struct {
	// A list of commands to run inside the parent Pod.
	Command []string `json:"command,omitempty"`

	// A list of args to be passed to the init container.
	Args []string `json:"args,omitempty"`

	// If true, inheirts the environment variables from the parent pod.
	// specification
	InheritEnv bool `json:"inheritEnv,omitempty"`

	// A list of environment variables used only by the initContainer.
	Env []v1.EnvVar `json:"env,omitempty"`
}

// DatabaseSpec is a struct defining a database to be exposed to a ClowdApp.
type DatabaseSpec struct {
	// Defines the Version of the PostGreSQL database, defaults to 12.
	// +kubebuilder:validation:Enum:=10;12;13
	Version *int32 `json:"version,omitempty"`

	// Defines the Name of the database to be created. This will be used as the
	// name of the logical database inside the database server in (*_local_*) mode
	// and the name of the secret to be used for Database configuration in (*_app-interface_*) mode.
	Name string `json:"name,omitempty"`

	// Defines the Name of the app to share a database from
	SharedDBAppName string `json:"sharedDbAppName,omitempty"`
}

// Job defines a CronJob as Schedule is required. In the future omitting the
// Schedule field will allow support for a standard Job resource.
type Job struct {
	// Name defines identifier of the Job. This name will be used to name the
	// CronJob resource, the container will be name identically.
	Name string `json:"name"`

	// Defines the schedule for the job to run
	Schedule string `json:"schedule,omitempty"`

	// PodSpec defines a container running inside the CronJob.
	PodSpec PodSpec `json:"podSpec"`

	// Defines the restart policy for the CronJob, defaults to never
	RestartPolicy v1.RestartPolicy `json:"restartPolicy,omitempty"`

	// Defines the concurrency policy for the CronJob, defaults to Allow
	ConcurrencyPolicy batch.ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// Defines the StartingDeadlineSeconds for the CronJob
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`
}

// WebDeprecated defines a boolean flag to help distinguish from the newer WebServices
type WebDeprecated bool

// PublicWebService is the definition of the public web service. There can be only
// one public service managed by Clowder.
type PublicWebService struct {

	// Enabled describes if Clowder should enable the public service and provide the
	// configuration in the cdappconfig.
	Enabled bool `json:"enabled,omitempty"`
}

// PrivateWebService is the definition of the private web service. There can be only
// one private service managed by Clowder.
type PrivateWebService struct {
	// Enabled describes if Clowder should enable the private service and provide the
	// configuration in the cdappconfig.
	Enabled bool `json:"enabled,omitempty"`
}

// MetricsWebService is the definition of the metrics web service. This is automatically
// enabled and the configuration here at the moment is included for completeness, as there
// are no configurable options.
type MetricsWebService struct {
}

// WebServices defines the structs for the three exposed web services: public,
// private and metrics.
type WebServices struct {
	Public  PublicWebService  `json:"public,omitempty"`
	Private PrivateWebService `json:"private,omitempty"`
	Metrics MetricsWebService `json:"metrics,omitempty"`
}

// K8sAccessLevel defines the access level for the deployment, one of 'default', 'view' or 'edit'
// +kubebuilder:validation:Enum={"default", "view", "", "edit"}
type K8sAccessLevel string

// Deployment defines a service running inside a ClowdApp and will output a deployment resource.
// Only one container per pod is allowed and this is defined in the PodSpec attribute.
type Deployment struct {
	// Name defines the identifier of a Pod inside the ClowdApp. This name will
	// be used along side the name of the ClowdApp itself to form a <app>-<pod>
	// pattern which will be used for all other created resources and also for
	// some labels. It must be unique within a ClowdApp.
	Name string `json:"name"`

	// Defines the minimum replica count for the pod.
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// If set to true, creates a service on the webPort defined in
	// the ClowdEnvironment resource, along with the relevant liveness and
	// readiness probes.
	Web WebDeprecated `json:"web,omitempty"`

	WebServices WebServices `json:"webServices,omitempty"`

	// PodSpec defines a container running inside a ClowdApp.
	PodSpec PodSpec `json:"podSpec"`

	// K8sAccessLevel defines the level of access for this deployment
	K8sAccessLevel K8sAccessLevel `json:"k8sAccessLevel,omitempty"`
}

type Sidecar struct {
	// The name of the sidecar, only supported names allowed, (token-refresher)
	Name string `json:"name"`

	// Defines if the sidecar is enabled, defaults to False
	Enabled bool `json:"enabled"`
}

// PodSpec defines a container running inside a ClowdApp.
type PodSpec struct {

	// Image refers to the container image used to create the pod.
	Image string `json:"image,omitempty"`

	// A list of init containers used to perform at-startup operations.
	InitContainers []InitContainer `json:"initContainers,omitempty"`

	// The command that will be invoked inside the pod at startup.
	Command []string `json:"command,omitempty"`

	// A list of args to be passed to the pod container.
	Args []string `json:"args,omitempty"`

	// A list of environment variables in k8s defined format.
	Env []v1.EnvVar `json:"env,omitempty"`

	// A pass-through of a resource requirements in k8s ResourceRequirements
	// format. If omitted, the default resource requirements from the
	// ClowdEnvironment will be used.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// A pass-through of a Liveness Probe specification in standard k8s format.
	// If omitted, a standard probe will be setup point to the webPort defined
	// in the ClowdEnvironment and a path of /healthz. Ignored if Web is set to
	// false.
	LivenessProbe *v1.Probe `json:"livenessProbe,omitempty"`

	// A pass-through of a Readiness Probe specification in standard k8s format.
	// If omitted, a standard probe will be setup point to the webPort defined
	// in the ClowdEnvironment and a path of /healthz. Ignored if Web is set to
	// false.
	ReadinessProbe *v1.Probe `json:"readinessProbe,omitempty"`

	// A pass-through of a list of Volumes in standa k8s format.
	Volumes []v1.Volume `json:"volumes,omitempty"`

	// A pass-through of a list of VolumesMounts in standa k8s format.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// Lists the expected side cars, will be validated in the validating webhook
	Sidecars []Sidecar `json:"sidecars,omitempty"`
}

// PodSpecDeprecated is a deprecated in favour of using the real k8s PodSpec object.
type PodSpecDeprecated struct {
	Name           string                  `json:"name"`
	Web            WebDeprecated           `json:"web,omitempty"`
	MinReplicas    *int32                  `json:"minReplicas,omitempty"`
	Image          string                  `json:"image,omitempty"`
	InitContainers []InitContainer         `json:"initContainers,omitempty"`
	Command        []string                `json:"command,omitempty"`
	Args           []string                `json:"args,omitempty"`
	Env            []v1.EnvVar             `json:"env,omitempty"`
	Resources      v1.ResourceRequirements `json:"resources,omitempty"`
	LivenessProbe  *v1.Probe               `json:"livenessProbe,omitempty"`
	ReadinessProbe *v1.Probe               `json:"readinessProbe,omitempty"`
	Volumes        []v1.Volume             `json:"volumes,omitempty"`
	VolumeMounts   []v1.VolumeMount        `json:"volumeMounts,omitempty"`
}

// CyndiSpec is used to indicate whether a ClowdApp needs database syndication configured by the
// cyndi operator and exposes a limited set of cyndi configuration options
type CyndiSpec struct {
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:validation:MinLength:=1
	// +kubebuilder:validation:MaxLength:=64
	// +kubebuilder:validation:Pattern:="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	AppName string `json:"appName,omitempty"`

	InsightsOnly bool `json:"insightsOnly,omitempty"`
}

// KafkaTopicSpec defines the desired state of KafkaTopic
type KafkaTopicSpec struct {
	// we re-define this spec rather than use strimzi.KafkaTopicSpec so that a ClowdApp's topic
	// spec has:
	//   * partitions optional
	//   * replicas optional
	//   * topicName required

	// A key/value pair describing the configuration of a particular topic.
	// +optional
	Config strimzi.KafkaTopicSpecConfig `json:"config,omitempty"`

	// The requested number of partitions for this topic. If unset, default is '3'
	// +optional
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=200000
	Partitions int32 `json:"partitions,omitempty"`

	// The requested number of replicas for this topic. If unset, default is '3'
	// +optional
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=32767
	Replicas int32 `json:"replicas,omitempty"`

	// The requested name for this topic.
	// +kubebuilder:validation:MinLength:=1
	// +kubebuilder:validation:MaxLength:=249
	// +kubebuilder:validation:Pattern:="[a-zA-Z0-9\\._\\-]"
	TopicName string `json:"topicName"`
}

type TestingSpec struct {
	IqePlugin string `json:"iqePlugin"`
}

// ClowdAppSpec is the main specification for a single Clowder Application
// it defines n pods along with dependencies that are shared between them.
type ClowdAppSpec struct {
	// A list of deployments
	Deployments []Deployment `json:"deployments,omitempty"`

	// A list of jobs
	Jobs []Job `json:"jobs,omitempty"`

	// Deprecated
	Pods []PodSpecDeprecated `json:"pods,omitempty"`

	// The name of the ClowdEnvironment resource that this ClowdApp will use as
	// its base. This does not mean that the ClowdApp needs to be placed in the
	// same directory as the targetNamespace of the ClowdEnvironment.
	EnvName string `json:"envName"`

	// A list of Kafka topics that will be created and made available to all
	// the pods listed in the ClowdApp.
	KafkaTopics []KafkaTopicSpec `json:"kafkaTopics,omitempty"`

	// The database specification defines a single database, the configuration
	// of which will be made available to all the pods in the ClowdApp.
	Database DatabaseSpec `json:"database,omitempty"`

	// A list of string names defining storage buckets. In certain modes,
	// defined by the ClowdEnvironment, Clowder will create those buckets.
	ObjectStore []string `json:"objectStore,omitempty"`

	// If inMemoryDb is set to true, Clowder will pass configuration
	// of an In Memory Database to the pods in the ClowdApp. This single
	// instance will be shared between all apps.
	InMemoryDB bool `json:"inMemoryDb,omitempty"`

	// If featureFlags is set to true, Clowder will pass configuration of a
	// FeatureFlags instance to the pods in the ClowdApp. This single
	// instance will be shared between all apps.
	FeatureFlags bool `json:"featureFlags,omitempty"`

	// A list of dependencies in the form of the name of the ClowdApps that are
	// required to be present for this ClowdApp to function.
	Dependencies []string `json:"dependencies,omitempty"`

	// A list of optional dependencies in the form of the name of the ClowdApps that are
	// will be added to the configuration when present.
	OptionalDependencies []string `json:"optionalDependencies,omitempty"`

	// Iqe plugin and other specifics
	Testing TestingSpec `json:"testing,omitempty"`

	// Configures 'cyndi' database syndication for this app. When the app's ClowdEnvironment has
	// the kafka provider set to (*_operator_*) mode, Clowder will configure a CyndiPipeline
	// for this app in the environment's kafka-connect namespace. When the kafka provider is in
	// (*_app-interface_*) mode, Clowder will check to ensure that a CyndiPipeline resource exists
	// for the application in the environment's kafka-connect namespace. For all other kafka
	// provider modes, this configuration option has no effect.
	Cyndi CyndiSpec `json:"cyndi,omitempty"`
}

type ClowdConditionType string

const (
	// Ready means all the deployments are ready
	DeploymentsReady ClowdConditionType = "DeploymentsReady"
	// ReconciliationSuccessful represents status of successful reconciliation
	ReconciliationSuccessful ClowdConditionType = "ReconciliationSuccessful"
	// ReconciliationPartiallySuccessful means the reconciliation is in a partial success state
	ReconciliationPartiallySuccessful ClowdConditionType = "ReconciliationPartiallySuccessful"
	// ReconciliationFailed means the reconciliation failed
	ReconciliationFailed ClowdConditionType = "ReconciliationFailed"
)

type ClowdCondition struct {
	// Type is the type of the condition.
	Type ClowdConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// ClowdAppStatus defines the observed state of ClowdApp
type ClowdAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// ClowdEnvironmentStatus defines the observed state of ClowdEnvironment
	Deployments common.DeploymentStatus `json:"deployments,omitempty"`
	Ready       bool                    `json:"ready"`
	Conditions  []ClowdCondition        `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=app
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.deployments.readyDeployments"
// +kubebuilder:printcolumn:name="Managed",type="integer",JSONPath=".status.deployments.managedDeployments"
// +kubebuilder:printcolumn:name="EnvName",type="string",JSONPath=".spec.envName"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ClowdApp is the Schema for the clowdapps API
type ClowdApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// A ClowdApp specification.
	Spec   ClowdAppSpec   `json:"spec,omitempty"`
	Status ClowdAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClowdAppList contains a list of ClowdApp
type ClowdAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// A list of ClowdApp Resources.
	Items []ClowdApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClowdApp{}, &ClowdAppList{})
}

// GetLabels returns a base set of labels relating to the ClowdApp.
func (i *ClowdApp) GetLabels() map[string]string {
	if i.Labels == nil {
		i.Labels = map[string]string{}
	}

	if _, ok := i.Labels["app"]; !ok {
		i.Labels["app"] = i.ObjectMeta.Name
	}

	newMap := make(map[string]string, len(i.Labels))

	for k, v := range i.Labels {
		newMap[k] = v
	}

	return newMap
}

// GetNamespacedName contructs a new namespaced name for an object from the pattern.
func (i *ClowdApp) GetNamespacedName(pattern string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: i.Namespace,
		Name:      fmt.Sprintf(pattern, i.Name),
	}
}

// GetIdent returns an ident <env>.<app> that should be unique across the cluster.
func (i *ClowdApp) GetIdent() string {
	return fmt.Sprintf("%v.%v", i.Spec.EnvName, i.Name)
}

// MakeOwnerReference defines the owner reference pointing to the ClowdApp resource.
func (i *ClowdApp) MakeOwnerReference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: i.APIVersion,
		Kind:       i.Kind,
		Name:       i.ObjectMeta.Name,
		UID:        i.ObjectMeta.UID,
		Controller: common.TruePtr(),
	}
}

// GetPrimaryLabel returns the primary label name use for identification.
func (i *ClowdApp) GetPrimaryLabel() string {
	return "app"
}

// GetClowdNamespace returns the namespace of the ClowdApp object.
func (i *ClowdApp) GetClowdNamespace() string {
	return i.Namespace
}

// GetClowdName returns the name of the ClowdApp object.
func (i *ClowdApp) GetClowdName() string {
	return i.Name
}

// GetUID returns ObjectMeta.UID
func (i *ClowdApp) GetUID() types.UID {
	return i.ObjectMeta.UID
}

// GetDeploymentStatus returns the Status.Deployments member
func (i *ClowdApp) GetDeploymentStatus() *common.DeploymentStatus {
	return &i.Status.Deployments
}

// GetDeploymentStatus returns the Status.Deployments member
func (i *ClowdApp) GetDeploymentNamespacedName(d *Deployment) types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-%s", i.Name, d.Name),
		Namespace: i.Namespace,
	}
}

// GetDeploymentStatus returns the Status.Deployments member
func (i *ClowdApp) GetCronJobNamespacedName(d *Job) types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-%s", i.Name, d.Name),
		Namespace: i.Namespace,
	}
}

// IsReady returns true when all the ManagedDeployments are Ready
func (i *ClowdApp) IsReady() bool {
	return (i.Status.Deployments.ManagedDeployments == i.Status.Deployments.ReadyDeployments)
}

// GetClowdSAName returns the ServiceAccount Name for the App
func (i *ClowdApp) GetClowdSAName() string {
	return fmt.Sprintf("%s-app", i.GetClowdName())
}

// ConvertToNewShim converts an old "pod" based spec into the new "deployment" style.
func (i *ClowdApp) ConvertToNewShim() {
	deps := []Deployment{}
	for _, pod := range i.Spec.Pods {
		dep := Deployment{
			Name:        pod.Name,
			Web:         pod.Web,
			MinReplicas: pod.MinReplicas,
			PodSpec: PodSpec{
				Image:          pod.Image,
				InitContainers: pod.InitContainers,
				Command:        pod.Command,
				Args:           pod.Args,
				Env:            pod.Env,
				Resources:      pod.Resources,
				LivenessProbe:  pod.LivenessProbe,
				ReadinessProbe: pod.ReadinessProbe,
				Volumes:        pod.Volumes,
				VolumeMounts:   pod.VolumeMounts,
			},
		}
		deps = append(deps, dep)
	}
	i.Spec.Deployments = deps
}

// Omfunc is a utility function that performs an operation on a metav1.Object.
type omfunc func(o metav1.Object)

// SetObjectMeta sets the metadata on a ClowdApp object.
func (i *ClowdApp) SetObjectMeta(o metav1.Object, opts ...omfunc) {
	o.SetName(i.Name)
	o.SetNamespace(i.Namespace)
	o.SetLabels(i.GetLabels())
	o.SetOwnerReferences([]metav1.OwnerReference{i.MakeOwnerReference()})

	for _, opt := range opts {
		opt(o)
	}
}

// Name returns a function that sets the name of an object to that of the
// passed in string.
func Name(name string) omfunc {
	return func(o metav1.Object) {
		o.SetName(name)
	}
}

// Namespace returns a function that sets the namespace of an object to that of the
// passed in string.
func Namespace(namespace string) omfunc {
	return func(o metav1.Object) {
		o.SetNamespace(namespace)
	}
}

// Labels returns a function that sets the labels of an object to that of the passed in labels.
func Labels(labels map[string]string) omfunc {
	return func(o metav1.Object) {
		o.SetLabels(labels)
	}
}

// GetAppInSameEnv populates the appList with a list of all apps in the same ClowdEnvironment. The
// environment is inferred from the given app.
func GetAppInSameEnv(ctx context.Context, pClient client.Client, app *ClowdApp, appList *ClowdAppList) error {
	err := pClient.List(ctx, appList, client.MatchingFields{"spec.envName": app.Spec.EnvName})

	if err != nil {
		return err
		// return errors.New("Could not get app list")
	}

	return nil
}

// GetAppForDBInSameEnv returns a point to a ClowdApp that has the sharedDB referenced by the given
// ClowdApp.
func GetAppForDBInSameEnv(ctx context.Context, pClient client.Client, app *ClowdApp) (*ClowdApp, error) {
	appList := &ClowdAppList{}
	var refApp ClowdApp

	err := GetAppInSameEnv(ctx, pClient, app, appList)

	if err != nil {
		return nil, err
	}

	for _, iapp := range appList.Items {
		if iapp.Name == app.Spec.Database.SharedDBAppName {
			refApp = iapp
			return &refApp, nil
		}
	}
	return nil, errors.New("could not get app for db in env")
}
