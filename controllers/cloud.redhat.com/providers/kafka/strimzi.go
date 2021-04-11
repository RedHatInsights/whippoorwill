package kafka

import (
	"fmt"
	"strconv"
	"strings"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	strimzi "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta1"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	p "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// KafkaTopic is the resource ident for a KafkaTopic object.
var KafkaTopic = p.NewSingleResourceIdent(ProvName, "kafka_topic", &strimzi.KafkaTopic{})

// KafkaInstance is the resource ident for a Kafka object.
var KafkaInstance = p.NewSingleResourceIdent(ProvName, "kafka_instance", &strimzi.Kafka{})

// KafkaConnect is the resource ident for a KafkaConnect object.
var KafkaConnect = p.NewSingleResourceIdent(ProvName, "kafka_connect", &strimzi.KafkaConnect{})

var conversionMap = map[string]func([]string) (string, error){
	"retention.ms":          utils.IntMax,
	"retention.bytes":       utils.IntMax,
	"min.compaction.lag.ms": utils.IntMax,
	"cleanup.policy":        utils.ListMerge,
}

type strimziProvider struct {
	p.Provider
	Config config.KafkaConfig
}

func (s *strimziProvider) configureKafkaCluster() error {
	clusterNN := types.NamespacedName{
		Namespace: getKafkaNamespace(s.Env),
		Name:      s.Env.Spec.Providers.Kafka.Cluster.Name,
	}
	k := &strimzi.Kafka{}
	if err := s.Cache.Create(KafkaInstance, clusterNN, k); err != nil {
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
		version = "2.5.0"
	}

	deleteClaim := s.Env.Spec.Providers.Kafka.Cluster.DeleteClaim

	k.Spec = &strimzi.KafkaSpec{
		Kafka: strimzi.KafkaSpecKafka{
			Version:  &version,
			Replicas: replicas,
			Listeners: []strimzi.KafkaSpecKafkaListenersElem{
				{Name: "tcp", Type: "internal", Tls: false, Port: 9092},
				{Name: "tls", Type: "internal", Tls: true, Port: 9093},
			},
		},
		Zookeeper: strimzi.KafkaSpecZookeeper{
			Replicas: replicas,
		},
		EntityOperator: &strimzi.KafkaSpecEntityOperator{
			TopicOperator: &strimzi.KafkaSpecEntityOperatorTopicOperator{},
		},
	}

	if s.Env.Spec.Providers.Kafka.PVC {
		k.Spec.Kafka.Storage = strimzi.KafkaSpecKafkaStorage{
			Type:        strimzi.KafkaSpecKafkaStorageTypePersistentClaim,
			Size:        &storageSize,
			DeleteClaim: &deleteClaim,
		}
		k.Spec.Zookeeper.Storage = strimzi.KafkaSpecZookeeperStorage{
			Type:        strimzi.KafkaSpecZookeeperStorageTypePersistentClaim,
			Size:        &storageSize,
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

	k.SetName(s.Env.Spec.Providers.Kafka.Cluster.Name)
	k.SetNamespace(getKafkaNamespace(s.Env))
	k.SetLabels(p.Labels{"env": s.Env.Name})
	k.SetOwnerReferences([]metav1.OwnerReference{s.Env.MakeOwnerReference()})

	if err := s.Cache.Update(KafkaInstance, k); err != nil {
		return err
	}

	return nil
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

func (s *strimziProvider) configureKafkaConnectCluster() error {
	clusterNN := types.NamespacedName{
		Namespace: getConnectNamespace(s.Env),
		Name:      getConnectClusterName(s.Env),
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
		version = "2.5.0"
	}

	image := s.Env.Spec.Providers.Kafka.Connect.Image
	if image == "" {
		image = "quay.io/cloudservices/xjoin-kafka-connect-strimzi:latest"
	}

	k.Spec = &strimzi.KafkaConnectSpec{
		Replicas:         &replicas,
		BootstrapServers: s.getBootstrapServersString(),
		Version:          &version,
		Config: map[string]string{
			"group.id":                          "connect-cluster",
			"offset.storage.topic":              "connect-cluster-offsets",
			"config.storage.topic":              "connect-cluster-configs",
			"status.storage.topic":              "connect-cluster-status",
			"offset.storage.replication.factor": "1",
			"config.storage.replication.factor": "1",
			"status.storage.replication.factor": "1",
		},
		Image: &image,
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
	k.SetLabels(p.Labels{"env": s.Env.Name})

	if err := s.Cache.Update(KafkaConnect, k); err != nil {
		return err
	}

	return nil
}

func (s *strimziProvider) configureListeners() error {
	// clusterNN := types.NamespacedName{
	// 	Namespace: getKafkaNamespace(s.Env),
	// 	Name:      s.Env.Spec.Providers.Kafka.Cluster.Name,
	// }
	// kafkaResource := strimzi.Kafka{}
	// if _, err := utils.UpdateOrErr(s.Client.Get(s.Ctx, clusterNN, &kafkaResource)); err != nil {
	// 	return err
	// }

	// // Return an err if we can't obtain listener status to trigger a requeue in the env controller
	// if kafkaResource.Status == nil || kafkaResource.Status.Listeners == nil {
	// 	return fmt.Errorf(
	// 		"Kafka cluster '%s' in ns '%s' has no listener status", clusterNN.Name, clusterNN.Namespace,
	// 	)
	// }

	s.Config.Brokers = []config.BrokerConfig{}
	s.Config.Brokers = append(s.Config.Brokers, config.BrokerConfig{
		Hostname: fmt.Sprintf("%s-kafka-bootstrap.%s.svc", s.Env.Spec.Providers.Kafka.Cluster.Name, getKafkaNamespace(s.Env)),
		Port:     utils.IntPtr(9092),
	})
	// for _, listener := range kafkaResource.Status.Listeners {
	// 	if listener.Type != nil && (*listener.Type == "plain" || *listener.Type == "tcp") {
	// 		bc := config.BrokerConfig{
	// 			Hostname: *listener.Addresses[0].Host,
	// 		}
	// 		port := listener.Addresses[0].Port
	// 		if port != nil {
	// 			p := int(*port)
	// 			bc.Port = &p
	// 		}
	// 		s.Config.Brokers = append(s.Config.Brokers, bc)
	// 	}
	// }

	// if len(s.Config.Brokers) < 1 {
	// 	return fmt.Errorf(
	// 		"Kafka cluster '%s' in ns '%s' has no listeners", clusterNN.Name, clusterNN.Namespace,
	// 	)
	// }

	return nil
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
func NewStrimzi(p *p.Provider) (providers.ClowderProvider, error) {
	kafkaProvider := &strimziProvider{
		Provider: *p,
		Config: config.KafkaConfig{
			Brokers: []config.BrokerConfig{},
		},
	}

	return kafkaProvider, kafkaProvider.configureBrokers()
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

	// update s.Config.Topics
	if err := s.processTopics(app); err != nil {
		return err
	}

	// set our provider's config on the AppConfig
	c.Kafka = &s.Config

	return nil
}

func (s *strimziProvider) processTopics(app *crd.ClowdApp) error {
	topicConfig := []config.TopicConfig{}

	appList := crd.ClowdAppList{}
	if err := s.Client.List(s.Ctx, &appList); err != nil {
		return errors.Wrap("Topic creation failed: Error listing apps", err)
	}

	nn := types.NamespacedName{
		Name:      app.Name,
		Namespace: app.Namespace,
	}

	for _, topic := range app.Spec.KafkaTopics {
		k := &strimzi.KafkaTopic{}

		topicName := fmt.Sprintf("%s-%s-%s", topic.TopicName, s.Env.Name, nn.Namespace)
		knn := types.NamespacedName{
			Namespace: getKafkaNamespace(s.Env),
			Name:      topicName,
		}

		if err := s.Cache.Create(KafkaTopic, knn, k); err != nil {
			return err
		}

		labels := p.Labels{
			"strimzi.io/cluster": s.Env.Spec.Providers.Kafka.Cluster.Name,
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

		err := processTopicValues(k, app, appList, topic, replicaValList, partitionValList)

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

func processTopicValues(
	k *strimzi.KafkaTopic,
	app *crd.ClowdApp,
	appList crd.ClowdAppList,
	topic crd.KafkaTopicSpec,
	replicaValList []string,
	partitionValList []string,
) error {

	for _, iapp := range appList.Items {

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
		maxReplicasInt, err := utils.Atoi32(maxReplicas)
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
		maxPartitionsInt, err := utils.Atoi32(maxPartitions)
		if err != nil {
			return errors.New(fmt.Sprintf("could not convert to string to int32 for %v", maxPartitions))
		}
		k.Spec.Partitions = maxPartitionsInt
		if k.Spec.Partitions < int32(1) {
			// if unset, default to 3
			k.Spec.Partitions = int32(3)
		}
	}

	return nil
}
