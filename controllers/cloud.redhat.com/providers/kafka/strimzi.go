package kafka

import (
	"fmt"
	"strconv"
	"strings"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1/common"
	strimzi "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta1"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/clowder_config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// KafkaTopic is the resource ident for a KafkaTopic object.
var KafkaTopic = providers.NewSingleResourceIdent(ProvName, "kafka_topic", &strimzi.KafkaTopic{})

// KafkaInstance is the resource ident for a Kafka object.
var KafkaInstance = providers.NewSingleResourceIdent(ProvName, "kafka_instance", &strimzi.Kafka{})

// KafkaConnect is the resource ident for a KafkaConnect object.
var KafkaConnect = providers.NewSingleResourceIdent(ProvName, "kafka_connect", &strimzi.KafkaConnect{})

// KafkaUser is the resource ident for a KafkaUser object.
var KafkaUser = providers.NewSingleResourceIdent(ProvName, "kafka_user", &strimzi.KafkaUser{})

// KafkaUser is the resource ident for a KafkaUser object.
var KafkaConnectUser = providers.NewSingleResourceIdent(ProvName, "kafka_connect_user", &strimzi.KafkaUser{})

// KafkaMetricsConfigMap is the resource ident for a KafkaMetricsConfigMap object.
var KafkaMetricsConfigMap = providers.NewSingleResourceIdent(ProvName, "kafka_metrics_config_map", &core.ConfigMap{})

// KafkaNetworkPolicy is the resource ident for the KafkaNetworkPolicy
var KafkaNetworkPolicy = providers.NewSingleResourceIdent(ProvName, "kafka_network_policy", &networking.NetworkPolicy{})

var conversionMap = map[string]func([]string) (string, error){
	"retention.ms":          utils.IntMax,
	"retention.bytes":       utils.IntMax,
	"min.compaction.lag.ms": utils.IntMax,
	"cleanup.policy":        utils.ListMerge,
}

type strimziProvider struct {
	providers.Provider
	Config config.KafkaConfig
}

func (s *strimziProvider) configureKafkaCluster() error {
	clusterNN := types.NamespacedName{
		Namespace: getKafkaNamespace(s.Env),
		Name:      getKafkaName(s.Env),
	}
	k := &strimzi.Kafka{}
	if err := s.Cache.Create(KafkaInstance, clusterNN, k); err != nil {
		return err
	}

	cmnn, err := s.createKafkaMetricsConfigMap()
	if err != nil {
		return err
	}

	// ensure that kafka cluster of this same name but labelled for different env does not exist
	if envLabel, ok := k.GetLabels()["env"]; ok {
		if envLabel != s.Env.Name {
			return fmt.Errorf(
				"kafka cluster named '%s' found in ns '%s' but tied to env '%s'",
				clusterNN.Name, clusterNN.Namespace, envLabel,
			)
		}
	}

	// populate options from the kafka provider's KafkaClusterOptions
	replicas := s.Env.Spec.Providers.Kafka.Cluster.Replicas
	if replicas < int32(1) {
		replicas = int32(1)
	}

	storageSize := s.Env.Spec.Providers.Kafka.Cluster.StorageSize
	if storageSize == "" {
		storageSize = "1Gi"
	}

	version := s.Env.Spec.Providers.Kafka.Cluster.Version
	if version == "" {
		version = "2.7.0"
	}

	deleteClaim := s.Env.Spec.Providers.Kafka.Cluster.DeleteClaim

	k.Spec = &strimzi.KafkaSpec{
		Kafka: strimzi.KafkaSpecKafka{
			Config: map[string]string{
				"offsets.topic.replication.factor": strconv.Itoa(int(replicas)),
			},
			Version:  &version,
			Replicas: replicas,
		},
		Zookeeper: strimzi.KafkaSpecZookeeper{
			Replicas: replicas,
		},
		EntityOperator: &strimzi.KafkaSpecEntityOperator{
			TopicOperator: &strimzi.KafkaSpecEntityOperatorTopicOperator{},
			UserOperator:  &strimzi.KafkaSpecEntityOperatorUserOperator{},
			TlsSidecar:    &strimzi.KafkaSpecEntityOperatorTlsSidecar{},
		},
	}

	if s.Env.Spec.Providers.Kafka.Cluster.Config != nil {
		k.Spec.Kafka.Config = s.Env.Spec.Providers.Kafka.Cluster.Config
	}

	k.Spec.Kafka.JvmOptions = &s.Env.Spec.Providers.Kafka.Cluster.JVMOptions

	var metrics apiextensions.JSON

	metrics.UnmarshalJSON(metricsData)

	metricsConfig := strimzi.KafkaSpecKafkaMetricsConfig{
		Type: "jmxPrometheusExporter",
		ValueFrom: strimzi.KafkaSpecKafkaMetricsConfigValueFrom{
			ConfigMapKeyRef: &strimzi.KafkaSpecKafkaMetricsConfigValueFromConfigMapKeyRef{
				Key:      common.StringPtr("metrics"),
				Name:     common.StringPtr(cmnn.Name),
				Optional: common.FalsePtr(),
			},
		},
	}

	k.Spec.Kafka.MetricsConfig = &metricsConfig
	k.Spec.Kafka.Resources = &s.Env.Spec.Providers.Kafka.Cluster.Resources

	listener := strimzi.KafkaSpecKafkaListenersElem{
		Type: "internal",
	}

	if s.Env.Spec.Providers.Kafka.EnableLegacyStrimzi {
		listener.Port = 9092
		listener.Tls = false
		listener.Name = "tcp"
	} else {
		listener.Port = 9093
		listener.Tls = true
		listener.Name = "tls"
		listener.Authentication = &strimzi.KafkaSpecKafkaListenersElemAuthentication{
			Type: strimzi.KafkaSpecKafkaListenersElemAuthenticationTypeScramSha512,
		}
		k.Spec.Kafka.Authorization = &strimzi.KafkaSpecKafkaAuthorization{
			Type: strimzi.KafkaSpecKafkaAuthorizationTypeSimple,
		}
	}

	k.Spec.Kafka.Listeners = []strimzi.KafkaSpecKafkaListenersElem{listener}

	if s.Env.Spec.Providers.Kafka.PVC {
		k.Spec.Kafka.Storage = strimzi.KafkaSpecKafkaStorage{
			Type:        strimzi.KafkaSpecKafkaStorageTypePersistentClaim,
			Size:        &storageSize,
			DeleteClaim: &deleteClaim,
		}

		zkStorageSize := "50Gi"

		kQuantity, err := resource.ParseQuantity(storageSize)

		if err == nil {
			zkQuantity, err := resource.ParseQuantity("50Gi")

			if err == nil && kQuantity.Cmp(zkQuantity) < 0 {
				// Kafka storage size is less than zkStorageSize
				zkStorageSize = storageSize
			}
		}

		k.Spec.Zookeeper.Storage = strimzi.KafkaSpecZookeeperStorage{
			Type:        strimzi.KafkaSpecZookeeperStorageTypePersistentClaim,
			Size:        &zkStorageSize,
			DeleteClaim: &deleteClaim,
		}
	} else {
		k.Spec.Kafka.Storage = strimzi.KafkaSpecKafkaStorage{
			Type: strimzi.KafkaSpecKafkaStorageTypeEphemeral,
		}
		k.Spec.Zookeeper.Storage = strimzi.KafkaSpecZookeeperStorage{
			Type: strimzi.KafkaSpecZookeeperStorageTypeEphemeral,
		}
	}

	k.SetName(getKafkaName(s.Env))
	k.SetNamespace(getKafkaNamespace(s.Env))
	k.SetLabels(providers.Labels{"env": s.Env.Name})
	k.SetOwnerReferences([]metav1.OwnerReference{s.Env.MakeOwnerReference()})

	if err := s.Cache.Update(KafkaInstance, k); err != nil {
		return err
	}

	return nil
}

func (s *strimziProvider) createKafkaMetricsConfigMap() (types.NamespacedName, error) {
	cm := &core.ConfigMap{}
	nn := types.NamespacedName{
		Namespace: getKafkaNamespace(s.Env),
		Name:      fmt.Sprintf("%s-metrics", getKafkaName(s.Env)),
	}

	if err := s.Cache.Create(KafkaMetricsConfigMap, nn, cm); err != nil {
		return types.NamespacedName{}, err
	}

	cm.Data = map[string]string{"metrics": string(metricsData)}

	cm.SetName(nn.Name)
	cm.SetNamespace(nn.Namespace)
	cm.SetLabels(providers.Labels{"env": s.Env.Name})
	cm.SetOwnerReferences([]metav1.OwnerReference{s.Env.MakeOwnerReference()})

	if err := s.Cache.Update(KafkaMetricsConfigMap, cm); err != nil {
		return types.NamespacedName{}, err
	}

	return nn, nil
}

func (s *strimziProvider) getBootstrapServersString() string {
	strArray := []string{}
	for _, bc := range s.Config.Brokers {
		if bc.Port != nil {
			strArray = append(strArray, fmt.Sprintf("%s:%d", bc.Hostname, *bc.Port))
		} else {
			strArray = append(strArray, bc.Hostname)
		}
	}
	return strings.Join(strArray, ",")
}

func (s *strimziProvider) createKafkaConnectUser() error {

	clusterNN := types.NamespacedName{
		Namespace: getConnectNamespace(s.Env),
		Name:      getConnectClusterUserName(s.Env),
	}

	ku := &strimzi.KafkaUser{}
	if err := s.Cache.Create(KafkaConnectUser, clusterNN, ku); err != nil {
		return err
	}

	labeler := utils.GetCustomLabeler(
		map[string]string{"strimzi.io/cluster": getKafkaName(s.Env)}, clusterNN, s.Env,
	)

	labeler(ku)

	ku.Spec = &strimzi.KafkaUserSpec{
		Authentication: &strimzi.KafkaUserSpecAuthentication{
			Type: strimzi.KafkaUserSpecAuthenticationTypeScramSha512,
		},
		Authorization: &strimzi.KafkaUserSpecAuthorization{
			Acls: []strimzi.KafkaUserSpecAuthorizationAclsElem{},
			Type: strimzi.KafkaUserSpecAuthorizationTypeSimple,
		},
	}

	address := "*"
	topic := "*"
	patternType := strimzi.KafkaUserSpecAuthorizationAclsElemResourcePatternTypeLiteral

	ku.Spec.Authorization.Acls = append(ku.Spec.Authorization.Acls, strimzi.KafkaUserSpecAuthorizationAclsElem{
		Host:      &address,
		Operation: strimzi.KafkaUserSpecAuthorizationAclsElemOperationAll,
		Resource: strimzi.KafkaUserSpecAuthorizationAclsElemResource{
			Name:        &topic,
			PatternType: &patternType,
			Type:        strimzi.KafkaUserSpecAuthorizationAclsElemResourceTypeTopic,
		},
	})

	group := "*"
	ku.Spec.Authorization.Acls = append(ku.Spec.Authorization.Acls, strimzi.KafkaUserSpecAuthorizationAclsElem{
		Host:      &address,
		Operation: strimzi.KafkaUserSpecAuthorizationAclsElemOperationAll,
		Resource: strimzi.KafkaUserSpecAuthorizationAclsElemResource{
			Name:        &group,
			PatternType: &patternType,
			Type:        strimzi.KafkaUserSpecAuthorizationAclsElemResourceTypeGroup,
		},
	})

	if err := s.Cache.Update(KafkaConnectUser, ku); err != nil {
		return err
	}

	return nil
}

func (s *strimziProvider) configureKafkaConnectCluster() error {
	clusterNN := types.NamespacedName{
		Namespace: getConnectNamespace(s.Env),
		Name:      getConnectClusterName(s.Env),
	}

	if err := s.createKafkaConnectUser(); err != nil {
		return err
	}

	k := &strimzi.KafkaConnect{}
	if err := s.Cache.Create(KafkaConnect, clusterNN, k); err != nil {
		return err
	}

	// ensure that connect cluster of this same name but labelled for different env does not exist
	if envLabel, ok := k.GetLabels()["env"]; ok {
		if envLabel != s.Env.Name {
			return fmt.Errorf(
				"kafka connect cluster named '%s' found in ns '%s' but tied to env '%s'",
				clusterNN.Name, clusterNN.Namespace, envLabel,
			)
		}
	}

	// populate options from the kafka provider's KafkaConnectClusterOptions
	replicas := s.Env.Spec.Providers.Kafka.Connect.Replicas
	if replicas < int32(1) {
		replicas = int32(1)
	}

	version := s.Env.Spec.Providers.Kafka.Connect.Version
	if version == "" {
		version = "2.7.0"
	}

	image := s.Env.Spec.Providers.Kafka.Connect.Image
	if image == "" {
		image = "quay.io/cloudservices/xjoin-kafka-connect-strimzi:latest"
	}

	username := getConnectClusterUserName(s.Env)

	k.Spec = &strimzi.KafkaConnectSpec{
		Replicas:         &replicas,
		BootstrapServers: s.getBootstrapServersString(),
		Version:          &version,
		Config: map[string]string{
			"group.id":                                "connect-cluster",
			"offset.storage.topic":                    "connect-cluster-offsets",
			"config.storage.topic":                    "connect-cluster-configs",
			"status.storage.topic":                    "connect-cluster-status",
			"offset.storage.replication.factor":       "1",
			"config.storage.replication.factor":       "1",
			"status.storage.replication.factor":       "1",
			"connector.client.config.override.policy": "All",
		},
		Image: &image,
	}
	if !s.Env.Spec.Providers.Kafka.EnableLegacyStrimzi {
		k.Spec.Tls = &strimzi.KafkaConnectSpecTls{
			TrustedCertificates: []strimzi.KafkaConnectSpecTlsTrustedCertificatesElem{{
				Certificate: "ca.crt",
				SecretName:  fmt.Sprintf("%s-cluster-ca-cert", getKafkaName(s.Env)),
			}},
		}
		k.Spec.Authentication = &strimzi.KafkaConnectSpecAuthentication{
			PasswordSecret: &strimzi.KafkaConnectSpecAuthenticationPasswordSecret{
				Password:   "password",
				SecretName: username,
			},
			Type:     "scram-sha-512",
			Username: &username,
		}
	}

	// configures this KafkaConnect to use KafkaConnector resources to avoid needing to call the
	// Connect REST API directly
	annotations := k.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["strimzi.io/use-connector-resources"] = "true"
	k.SetAnnotations(annotations)
	k.SetOwnerReferences([]metav1.OwnerReference{s.Env.MakeOwnerReference()})
	k.SetName(getConnectClusterName(s.Env))
	k.SetNamespace(getConnectNamespace(s.Env))
	k.SetLabels(providers.Labels{"env": s.Env.Name})

	if err := s.Cache.Update(KafkaConnect, k); err != nil {
		return err
	}

	return nil
}

func (s *strimziProvider) configureListeners() error {
	clusterNN := types.NamespacedName{
		Namespace: getKafkaNamespace(s.Env),
		Name:      getKafkaName(s.Env),
	}
	kafkaResource := strimzi.Kafka{}
	if _, err := utils.UpdateOrErr(s.Client.Get(s.Ctx, clusterNN, &kafkaResource)); err != nil {
		return err
	}

	// Return an err if we can't obtain listener status to trigger a requeue in the env controller
	if kafkaResource.Status == nil || kafkaResource.Status.Listeners == nil {
		return fmt.Errorf(
			"kafka cluster '%s' in ns '%s' has no listener status", clusterNN.Name, clusterNN.Namespace,
		)
	}

	kafkaCASecName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-cluster-ca-cert", getKafkaName(s.Env)),
		Namespace: getKafkaNamespace(s.Env),
	}
	kafkaCASecret := core.Secret{}
	if _, err := utils.UpdateOrErr(s.Client.Get(s.Ctx, kafkaCASecName, &kafkaCASecret)); err != nil {
		return err
	}

	kafkaCACert := string(kafkaCASecret.Data["ca.crt"])

	s.Config.Brokers = []config.BrokerConfig{}
	for _, listener := range kafkaResource.Status.Listeners {
		if listener.Type != nil && *listener.Type == "tls" {
			s.Config.Brokers = append(s.Config.Brokers, buildTlsBrokerConfig(listener, kafkaCACert))
		} else if listener.Type != nil && (*listener.Type == "plain" || *listener.Type == "tcp") {
			s.Config.Brokers = append(s.Config.Brokers, buildTcpBrokerConfig(listener))
		}
	}

	if len(s.Config.Brokers) < 1 {
		return fmt.Errorf(
			"kafka cluster '%s' in ns '%s' has no listeners", clusterNN.Name, clusterNN.Namespace,
		)
	}

	return nil
}

func buildTcpBrokerConfig(listener strimzi.KafkaStatusListenersElem) config.BrokerConfig {
	bc := config.BrokerConfig{
		Hostname: *listener.Addresses[0].Host,
	}
	port := listener.Addresses[0].Port
	if port != nil {
		p := int(*port)
		bc.Port = &p
	}
	return bc
}

func buildTlsBrokerConfig(listener strimzi.KafkaStatusListenersElem, caCert string) config.BrokerConfig {
	authType := config.BrokerConfigAuthtypeSasl
	bc := config.BrokerConfig{
		Sasl:     &config.KafkaSASLConfig{},
		Cacert:   &caCert,
		Hostname: *listener.Addresses[0].Host,
		Authtype: &authType,
	}
	port := listener.Addresses[0].Port
	if port != nil {
		p := int(*port)
		bc.Port = &p
	}
	return bc
}

func (s *strimziProvider) configureBrokers() error {
	if err := s.configureKafkaCluster(); err != nil {
		return errors.Wrap("failed to provision kafka cluster", err)
	}

	// Look up Kafka cluster's listeners and configure s.Config.Brokers
	// (we need to know the bootstrap server addresses before provisioning KafkaConnect)
	if err := s.configureListeners(); err != nil {
		clowdErr := errors.Wrap("unable to determine kafka broker addresses", err)
		clowdErr.Requeue = true
		return clowdErr
	}

	if err := s.configureKafkaConnectCluster(); err != nil {
		return errors.Wrap("failed to provision kafka connect cluster", err)
	}

	return nil
}

// NewStrimzi returns a new strimzi provider object.
func NewStrimzi(p *providers.Provider) (providers.ClowderProvider, error) {
	kafkaProvider := &strimziProvider{
		Provider: *p,
		Config: config.KafkaConfig{
			Brokers: []config.BrokerConfig{},
		},
	}

	if err := createNetworkPolicies(p); err != nil {
		return nil, err
	}

	return kafkaProvider, kafkaProvider.configureBrokers()
}

func createNetworkPolicies(p *providers.Provider) error {
	appList, err := p.Env.GetAppsInEnv(p.Ctx, p.Client)
	if err != nil {
		return err
	}

	namespaceSet := map[string]bool{}

	np := &networking.NetworkPolicy{}
	nn := types.NamespacedName{
		Name:      getKafkaName(p.Env),
		Namespace: getKafkaNamespace(p.Env),
	}

	if err := p.Cache.Create(KafkaNetworkPolicy, nn, np); err != nil {
		return err
	}

	npFrom := []networking.NetworkPolicyPeer{}

	for _, app := range appList.Items {
		if _, ok := namespaceSet[app.Namespace]; ok {
			continue
		}

		npFrom = append(npFrom, networking.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": app.Namespace,
				},
			},
		})
		namespaceSet[app.Namespace] = true
	}

	np.Spec.Ingress = []networking.NetworkPolicyIngressRule{{
		From: npFrom,
	}}

	np.Spec.PolicyTypes = []networking.PolicyType{"Ingress"}

	labeler := utils.GetCustomLabeler(nil, nn, p.Env)
	labeler(np)

	if err := p.Cache.Update(KafkaNetworkPolicy, np); err != nil {
		return err
	}

	return nil
}

func (s *strimziProvider) Provide(app *crd.ClowdApp, c *config.AppConfig) error {
	if app.Spec.Cyndi.Enabled {
		err := createCyndiPipeline(
			s.Ctx, s.Client, s.Cache, app, s.Env, getConnectNamespace(s.Env), getConnectClusterName(s.Env),
		)
		if err != nil {
			return err
		}
	}

	if len(app.Spec.KafkaTopics) == 0 {
		return nil
	}

	if err := s.processTopics(app); err != nil {
		return err
	}

	if err := s.createKafkaUser(app); err != nil {
		return err
	}

	if err := s.setBrokerCredentials(app); err != nil {
		return err
	}

	// set our provider's config on the AppConfig
	c.Kafka = &s.Config

	return nil
}

func (s *strimziProvider) setBrokerCredentials(app *crd.ClowdApp) error {
	for _, broker := range s.Config.Brokers {
		if broker.Authtype == nil {
			continue
		}
		if *broker.Authtype == config.BrokerConfigAuthtypeSasl {
			ku := &strimzi.KafkaUser{}
			nn := types.NamespacedName{
				Name:      getKafkaUsername(s.Env, app),
				Namespace: getKafkaNamespace(s.Env),
			}

			err := s.Client.Get(s.Ctx, nn, ku)
			if err != nil {
				return err
			}

			if ku.Status == nil || ku.Status.Username == nil {
				return errors.New("no username in kafkauser status")
			}
			broker.Sasl.Username = ku.Status.Username

			if ku.Status.Secret == nil {
				return errors.New("no secret in kafkauser status")
			}

			secnn := types.NamespacedName{
				Name:      *ku.Status.Secret,
				Namespace: getKafkaNamespace(s.Env),
			}

			kafkaSecret := &core.Secret{}

			err = s.Client.Get(s.Ctx, secnn, kafkaSecret)
			if err != nil {
				return err
			}

			if kafkaSecret.Data["password"] == nil {
				return errors.New("no password in kafkauser secret")
			}
			password := string(kafkaSecret.Data["password"])
			broker.Sasl.Password = &password
		}
	}
	return nil
}

func (s *strimziProvider) createKafkaUser(app *crd.ClowdApp) error {

	ku := &strimzi.KafkaUser{}
	nn := types.NamespacedName{
		Name:      getKafkaUsername(s.Env, app),
		Namespace: getKafkaNamespace(s.Env),
	}

	if err := s.Cache.Create(KafkaUser, nn, ku); err != nil {
		return err
	}
	labeler := utils.GetCustomLabeler(
		map[string]string{"strimzi.io/cluster": getKafkaName(s.Env)}, nn, s.Env,
	)

	labeler(ku)

	ku.Spec = &strimzi.KafkaUserSpec{
		Authentication: &strimzi.KafkaUserSpecAuthentication{
			Type: strimzi.KafkaUserSpecAuthenticationTypeScramSha512,
		},
		Authorization: &strimzi.KafkaUserSpecAuthorization{
			Acls: []strimzi.KafkaUserSpecAuthorizationAclsElem{},
			Type: strimzi.KafkaUserSpecAuthorizationTypeSimple,
		},
	}

	address := "*"
	patternType := strimzi.KafkaUserSpecAuthorizationAclsElemResourcePatternTypeLiteral

	for _, topic := range app.Spec.KafkaTopics {
		topicName := getTopicName(topic, *s.Env, app.Namespace)

		ku.Spec.Authorization.Acls = append(ku.Spec.Authorization.Acls, strimzi.KafkaUserSpecAuthorizationAclsElem{
			Host:      &address,
			Operation: strimzi.KafkaUserSpecAuthorizationAclsElemOperationAll,
			Resource: strimzi.KafkaUserSpecAuthorizationAclsElemResource{
				Name:        &topicName,
				PatternType: &patternType,
				Type:        strimzi.KafkaUserSpecAuthorizationAclsElemResourceTypeTopic,
			},
		})
	}

	group := "*"
	ku.Spec.Authorization.Acls = append(ku.Spec.Authorization.Acls, strimzi.KafkaUserSpecAuthorizationAclsElem{
		Host:      &address,
		Operation: strimzi.KafkaUserSpecAuthorizationAclsElemOperationAll,
		Resource: strimzi.KafkaUserSpecAuthorizationAclsElemResource{
			Name:        &group,
			PatternType: &patternType,
			Type:        strimzi.KafkaUserSpecAuthorizationAclsElemResourceTypeGroup,
		},
	})

	if err := s.Cache.Update(KafkaUser, ku); err != nil {
		return err
	}

	return nil
}

func (s *strimziProvider) processTopics(app *crd.ClowdApp) error {
	topicConfig := []config.TopicConfig{}

	appList := crd.ClowdAppList{}
	if err := s.Client.List(s.Ctx, &appList); err != nil {
		return errors.Wrap("Topic creation failed: Error listing apps", err)
	}

	for _, topic := range app.Spec.KafkaTopics {
		k := &strimzi.KafkaTopic{}

		topicName := getTopicName(topic, *s.Env, app.Namespace)
		knn := types.NamespacedName{
			Namespace: getKafkaNamespace(s.Env),
			Name:      topicName,
		}

		if err := s.Cache.Create(KafkaTopic, knn, k); err != nil {
			return err
		}

		labels := providers.Labels{
			"strimzi.io/cluster": getKafkaName(s.Env),
			"env":                app.Spec.EnvName,
			// If we label it with the app name, since app names should be
			// unique? can we use for delete selector?
		}

		k.SetName(topicName)
		k.SetNamespace(getKafkaNamespace(s.Env))
		// the ClowdEnvironment is the owner of this topic
		k.SetOwnerReferences([]metav1.OwnerReference{s.Env.MakeOwnerReference()})
		k.SetLabels(labels)

		k.Spec = &strimzi.KafkaTopicSpec{
			Config: make(map[string]string),
		}

		// This can be improved from an efficiency PoV
		// Loop through all key/value pairs in the config
		replicaValList := []string{}
		partitionValList := []string{}

		err := processTopicValues(k, s.Env, app, appList, topic, replicaValList, partitionValList)

		if err != nil {
			return err
		}

		if err := s.Cache.Update(KafkaTopic, k); err != nil {
			return err
		}

		topicConfig = append(
			topicConfig,
			config.TopicConfig{Name: topicName, RequestedName: topic.TopicName},
		)
	}

	s.Config.Topics = topicConfig

	return nil
}

func getTopicName(topic crd.KafkaTopicSpec, env crd.ClowdEnvironment, namespace string) string {
	if clowder_config.LoadedConfig.Features.UseComplexStrimziTopicNames {
		return fmt.Sprintf("%s-%s-%s", topic.TopicName, env.Name, namespace)
	} else {
		return topic.TopicName
	}
}

func processTopicValues(
	k *strimzi.KafkaTopic,
	env *crd.ClowdEnvironment,
	app *crd.ClowdApp,
	appList crd.ClowdAppList,
	topic crd.KafkaTopicSpec,
	replicaValList []string,
	partitionValList []string,
) error {

	for _, iapp := range appList.Items {

		if app.Spec.Pods != nil {
			app.ConvertToNewShim()
		}

		if iapp.Spec.EnvName != app.Spec.EnvName {
			// Only consider apps within this ClowdEnvironment
			continue
		}
		if iapp.Spec.KafkaTopics != nil {
			for _, itopic := range iapp.Spec.KafkaTopics {
				if itopic.TopicName != topic.TopicName {
					// Only consider a topic that matches the name
					continue
				}
				replicaValList = append(replicaValList, strconv.Itoa(int(itopic.Replicas)))
				partitionValList = append(partitionValList, strconv.Itoa(int(itopic.Partitions)))
			}
		}
	}

	for key := range topic.Config {
		valList := []string{}
		for _, iapp := range appList.Items {
			if iapp.Spec.EnvName != app.Spec.EnvName {
				// Only consider apps within this ClowdEnvironment
				continue
			}
			if iapp.Spec.KafkaTopics != nil {
				for _, itopic := range app.Spec.KafkaTopics {
					if itopic.TopicName != topic.TopicName {
						// Only consider a topic that matches the name
						continue
					}
					replicaValList = append(replicaValList, strconv.Itoa(int(itopic.Replicas)))
					partitionValList = append(partitionValList, strconv.Itoa(int(itopic.Partitions)))
					if itopic.Config != nil {
						if val, ok := itopic.Config[key]; ok {
							valList = append(valList, val)
						}
					}
				}
			}
		}
		f, ok := conversionMap[key]
		if ok {
			out, _ := f(valList)
			k.Spec.Config[key] = out
		} else {
			return errors.New(fmt.Sprintf("no conversion type for %s", key))
		}
	}

	if len(replicaValList) > 0 {
		maxReplicas, err := utils.IntMax(replicaValList)
		if err != nil {
			return errors.New(fmt.Sprintf("could not compute max for %v", replicaValList))
		}
		maxReplicasInt, err := common.Atoi32(maxReplicas)
		if err != nil {
			return errors.New(fmt.Sprintf("could not convert string to int32 for %v", maxReplicas))
		}
		k.Spec.Replicas = maxReplicasInt
		if k.Spec.Replicas < int32(1) {
			// if unset, default to 3
			k.Spec.Replicas = int32(3)
		}
	}

	if len(partitionValList) > 0 {
		maxPartitions, err := utils.IntMax(partitionValList)
		if err != nil {
			return errors.New(fmt.Sprintf("could not compute max for %v", partitionValList))
		}
		maxPartitionsInt, err := common.Atoi32(maxPartitions)
		if err != nil {
			return errors.New(fmt.Sprintf("could not convert to string to int32 for %v", maxPartitions))
		}
		k.Spec.Partitions = maxPartitionsInt
		if k.Spec.Partitions < int32(1) {
			// if unset, default to 3
			k.Spec.Partitions = int32(3)
		}
	}

	if env.Spec.Providers.Kafka.Cluster.Replicas < int32(1) {
		k.Spec.Replicas = 1
	} else if env.Spec.Providers.Kafka.Cluster.Replicas < k.Spec.Replicas {
		k.Spec.Replicas = env.Spec.Providers.Kafka.Cluster.Replicas
	}

	return nil
}
