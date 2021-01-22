package kafka

import (
	"fmt"
	"strconv"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	strimzi "cloud.redhat.com/clowder/v2/apis/kafka.strimzi.io/v1beta1"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	p "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"
	"k8s.io/apimachinery/pkg/types"
)

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

func (s *strimziProvider) configureBrokers() error {
	clusterName := types.NamespacedName{
		Namespace: getKafkaNamespace(s.Env),
		Name:      s.Env.Spec.Providers.Kafka.ClusterName,
	}

	kafkaResource := strimzi.Kafka{}

	if _, err := utils.UpdateOrErr(s.Client.Get(s.Ctx, clusterName, &kafkaResource)); err != nil {
		return err
	}

	for _, listener := range kafkaResource.Status.Listeners {
		if listener.Type == "plain" {
			bc := config.BrokerConfig{
				Hostname: listener.Addresses[0].Host,
			}
			port := listener.Addresses[0].Port
			if port != nil {
				p := int(*port)
				bc.Port = &p
			}
			s.Config.Brokers = append(s.Config.Brokers, bc)
		}
	}

	return nil
}

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
			s.Ctx,
			s.Client,
			app,
			getConnectNamespace(s.Env, getKafkaNamespace(s.Env)),
			getConnectClusterName(s.Env, "kafka-connect-cluster"),
		)
		if err != nil {
			return err
		}
	}

	if len(app.Spec.KafkaTopics) == 0 {
		return nil
	}

	s.Config.Topics = []config.TopicConfig{}

	nn := types.NamespacedName{
		Name:      app.Name,
		Namespace: app.Namespace,
	}

	appList := crd.ClowdAppList{}
	err := s.Client.List(s.Ctx, &appList)

	for _, topic := range app.Spec.KafkaTopics {

		if err != nil {
			return errors.Wrap("Topic creation failed: Error listing apps", err)
		}

		k := strimzi.KafkaTopic{}

		topicName := fmt.Sprintf("%s-%s-%s", topic.TopicName, s.Env.Name, nn.Namespace)

		update, err := utils.UpdateOrErr(s.Client.Get(s.Ctx, types.NamespacedName{
			Namespace: getKafkaNamespace(s.Env),
			Name:      topicName,
		}, &k))

		if err != nil {
			return err
		}

		labels := p.Labels{
			"strimzi.io/cluster": s.Env.Spec.Providers.Kafka.ClusterName,
			"env":                app.Spec.EnvName,
			// If we label it with the app name, since app names should be
			// unique? can we use for delete selector?
		}

		k.SetName(topicName)
		k.SetNamespace(getKafkaNamespace(s.Env))
		k.SetLabels(labels)

		newConfig := make(map[string]string)

		// This can be improved from an efficiency PoV
		// Loop through all key/value pairs in the config
		replicaValList := []string{}
		partitionValList := []string{}

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
					if itopic.Replicas != nil {
						replicaValList = append(replicaValList, strconv.Itoa(int(*itopic.Replicas)))
					}
					if itopic.Partitions != nil {
						partitionValList = append(partitionValList, strconv.Itoa(int(*itopic.Partitions)))
					}
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
						if itopic.Replicas != nil {
							replicaValList = append(replicaValList, strconv.Itoa(int(*itopic.Replicas)))
						}
						if itopic.Partitions != nil {
							partitionValList = append(partitionValList, strconv.Itoa(int(*itopic.Partitions)))
						}
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
				newConfig[key] = out
			} else {
				return errors.New(fmt.Sprintf("no conversion type for %s", key))
			}
		}
		if len(replicaValList) > 0 {
			maxReplicas, err := utils.IntMax(replicaValList)
			if err != nil {
				return errors.New(fmt.Sprintf("could not compute max for %v", replicaValList))
			}
			maxReplicaString, err := strconv.Atoi(maxReplicas)
			if err != nil {
				return errors.New(fmt.Sprintf("could not convert to string %v", maxReplicas))
			}
			k.Spec.Replicas = utils.Int32(maxReplicaString)
		}

		if len(partitionValList) > 0 {
			maxPartitions, err := utils.IntMax(partitionValList)
			if err != nil {
				return errors.New(fmt.Sprintf("could not compute max for %v", partitionValList))
			}
			maxPartitionString, err := strconv.Atoi(maxPartitions)
			if err != nil {
				return errors.New(fmt.Sprintf("could not convert to string %v", maxPartitions))
			}
			k.Spec.Partitions = utils.Int32(maxPartitionString)
		}

		k.Spec.Config = newConfig

		if err = update.Apply(s.Ctx, s.Client, &k); err != nil {
			return err
		}

		s.Config.Topics = append(
			s.Config.Topics,
			config.TopicConfig{Name: topicName, RequestedName: topic.TopicName},
		)
	}

	c.Kafka = &s.Config

	return nil
}
