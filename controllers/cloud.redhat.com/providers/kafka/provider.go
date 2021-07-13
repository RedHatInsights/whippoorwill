package kafka

import (
	"fmt"
	"strings"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	cyndi "cloud.redhat.com/clowder/v2/apis/cyndi-operator/v1alpha1"
	core "k8s.io/api/core/v1"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
)

// ProvName is the name/ident of the provider
var ProvName = "kafka"

// CyndiPipeline identifies the main cyndi pipeline object.
var CyndiPipeline = providers.NewSingleResourceIdent(ProvName, "cyndi_pipeline", &cyndi.CyndiPipeline{})

// CyndiAppSecret identifies the cyndi app secret object.
var CyndiAppSecret = providers.NewSingleResourceIdent(ProvName, "cyndi_app_secret", &core.Secret{})

// CyndiHostInventoryAppSecret identifies the cyndi host-inventory app secret object.
var CyndiHostInventoryAppSecret = providers.NewSingleResourceIdent(ProvName, "cyndi_host_inventory_secret", &core.Secret{})

// GetKafka returns the correct kafka provider based on the environment.
func GetKafka(c *providers.Provider) (providers.ClowderProvider, error) {
	c.Env.ConvertDeprecatedKafkaSpec()
	kafkaMode := c.Env.Spec.Providers.Kafka.Mode
	switch kafkaMode {
	case "operator":
		return NewStrimzi(c)
	case "local":
		return NewLocalKafka(c)
	case "app-interface":
		return NewAppInterface(c)
	case "managed":
		return NewManagedKafka(c)
	case "none", "":
		return NewNoneKafka(c)
	default:
		errStr := fmt.Sprintf("No matching kafka mode for %s", kafkaMode)
		return nil, errors.New(errStr)
	}
}

func getKafkaUsername(env *crd.ClowdEnvironment, app *crd.ClowdApp) string {
	return fmt.Sprintf("%s-%s", env.Name, app.Name)
}

func getKafkaName(e *crd.ClowdEnvironment) string {
	if e.Spec.Providers.Kafka.Cluster.Name == "" {
		// generate a unique name based on the ClowdEnvironment's UID

		// convert e.UID (which is a apimachinery types.UID) to string
		// types.UID is a string alias so this should not fail...
		uidString := string(e.UID)

		// append the initial portion of the UUID onto the kafka cluster's name
		return fmt.Sprintf("%s-%s", e.Name, strings.Split(uidString, "-")[0])
	}
	return e.Spec.Providers.Kafka.Cluster.Name
}

func getKafkaNamespace(e *crd.ClowdEnvironment) string {
	if e.Spec.Providers.Kafka.Cluster.Namespace == "" {
		return e.Status.TargetNamespace
	}
	return e.Spec.Providers.Kafka.Cluster.Namespace
}

func getConnectNamespace(env *crd.ClowdEnvironment) string {
	if env.Spec.Providers.Kafka.Connect.Namespace == "" {
		return getKafkaNamespace(env)
	}
	return env.Spec.Providers.Kafka.Connect.Namespace
}

func getConnectClusterName(env *crd.ClowdEnvironment) string {
	if env.Spec.Providers.Kafka.Connect.Name == "" {
		return getKafkaName(env)
	}
	return env.Spec.Providers.Kafka.Connect.Name
}

func getConnectClusterUserName(env *crd.ClowdEnvironment) string {
	return fmt.Sprintf("%s-connect", env.Name)
}

func init() {
	providers.ProvidersRegistration.Register(GetKafka, 6, ProvName)
}
