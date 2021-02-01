package kafka

import (
	"context"
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	obj "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/object"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	p "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	"github.com/segmentio/kafka-go"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/utils"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type envVar struct {
	Name  string
	Value string
}

type localKafka struct {
	providers.Provider
	Config config.KafkaConfig
}

func createLocalKafkaTopic(ctx context.Context, host string, topicName string, partition int) error {
	conn, err := kafka.DialLeader(ctx, "tcp", host, topicName, partition)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func (k *localKafka) Provide(app *crd.ClowdApp, c *config.AppConfig) error {
	if app.Spec.Cyndi.Enabled {
		// for now we're assuming the kafka connect cluster is already present in the namespace
		err := createCyndiPipeline(
			k.Ctx,
			k.Client,
			app,
			getConnectNamespace(k.Env, k.Env.GetClowdNamespace()),
			getConnectClusterName(k.Env, "kafka-connect-cluster"),
		)
		if err != nil {
			return err
		}
	}

	host := fmt.Sprintf("%s:29092", k.Config.Brokers[0].Hostname)

	err := createLocalKafkaTopic(k.Ctx, host, KafkaLoggingTopicName, 0)
	if err != nil {
		return err
	}

	if len(app.Spec.KafkaTopics) == 0 {
		return nil
	}

	for _, topic := range app.Spec.KafkaTopics {
		topicName := fmt.Sprintf(
			"%s-%s-%s", topic.TopicName, k.Env.Name, k.Env.GetClowdNamespace(),
		)

		k.Config.Topics = append(
			k.Config.Topics,
			config.TopicConfig{
				Name:          topicName,
				RequestedName: topic.TopicName,
			},
		)

		err := createLocalKafkaTopic(k.Ctx, host, topicName, 0)
		if err != nil {
			return err
		}
	}

	c.Kafka = &k.Config
	return nil
}

// NewLocalKafka returns a LocalKafka provider
func NewLocalKafka(p *p.Provider) (providers.ClowderProvider, error) {
	config := config.KafkaConfig{
		Topics: []config.TopicConfig{},
		Brokers: []config.BrokerConfig{{
			Hostname: fmt.Sprintf("%v-kafka.%v.svc", p.Env.Name, p.Env.GetClowdNamespace()),
			Port:     utils.IntPtr(29092),
		}},
	}

	kafkaProvider := localKafka{
		Provider: *p,
		Config:   config,
	}

	if err := providers.MakeComponent(p.Ctx, p.Client, p.Env, "zookeeper", makeLocalZookeeper, p.Env.Spec.Providers.Kafka.PVC); err != nil {
		return &kafkaProvider, err
	}

	if err := providers.MakeComponent(p.Ctx, p.Client, p.Env, "kafka", makeLocalKafka, p.Env.Spec.Providers.Kafka.PVC); err != nil {
		return &kafkaProvider, err
	}

	return &kafkaProvider, nil
}

func makeEnvVars(list *[]envVar) []core.EnvVar {

	envVars := []core.EnvVar{}

	for _, ev := range *list {
		envVars = append(envVars, core.EnvVar{Name: ev.Name, Value: ev.Value})
	}

	return envVars
}

func makeLocalKafka(o obj.ClowdObject, dd *apps.Deployment, svc *core.Service, pvc *core.PersistentVolumeClaim, usePVC bool) {
	nn := providers.GetNamespacedName(o, "kafka")

	labels := o.GetLabels()
	labels["env-app"] = nn.Name
	labeler := utils.MakeLabeler(nn, labels, o)

	labeler(dd)

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

	dd.Spec.Replicas = utils.Int32(1)
	dd.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	dd.Spec.Template.Spec.Volumes = []core.Volume{
		{
			Name:         nn.Name,
			VolumeSource: volSource,
		},
		{
			Name: "mq-kafka-1",
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "mq-kafka-2",
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
	}
	dd.Spec.Template.ObjectMeta.Labels = labels

	envVars := makeEnvVars(&[]envVar{
		{"KAFKA_ADVERTISED_LISTENERS", "PLAINTEXT://" + nn.Name + "." + nn.Namespace + ".svc:29092, LOCAL://localhost:9092"},
		{"KAFKA_BROKER_ID", "1"},
		{"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR", "1"},
		{"KAFKA_ZOOKEEPER_CONNECT", o.GetClowdName() + "-zookeeper:32181"},
		{"LOG_DIR", "/var/lib/kafka"},
		{"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", "PLAINTEXT:PLAINTEXT, LOCAL:PLAINTEXT"},
		{"KAFKA_INTER_BROKER_LISTENER_NAME", "LOCAL"},
	})

	ports := []core.ContainerPort{
		{
			Name:          "kafka",
			ContainerPort: 9092,
		},
	}

	probeHandler := core.Handler{
		TCPSocket: &core.TCPSocketAction{
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 9092,
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
		Name:  nn.Name,
		Image: "quay.io/cloudservices/cp-kafka:5.3.2",
		Env:   envVars,
		Ports: ports,
		VolumeMounts: []core.VolumeMount{
			{
				Name:      nn.Name,
				MountPath: "/var/lib/kafka",
			},
			{
				Name:      "mq-kafka-1",
				MountPath: "/etc/kafka/secrets",
			},
			{
				Name:      "mq-kafka-2",
				MountPath: "/var/lib/kafka/data",
			},
		},
		ReadinessProbe: &readinessProbe,
		LivenessProbe:  &livenessProbe,
	}

	dd.Spec.Template.Spec.Containers = []core.Container{c}
	dd.Spec.Template.SetLabels(labels)

	servicePorts := []core.ServicePort{{Name: "kafka", Port: 29092, Protocol: "TCP"}}

	utils.MakeService(svc, nn, labels, servicePorts, o)
	if usePVC {
		utils.MakePVC(pvc, nn, labels, "1Gi", o)
	}
}

func makeLocalZookeeper(o obj.ClowdObject, dd *apps.Deployment, svc *core.Service, pvc *core.PersistentVolumeClaim, usePVC bool) {

	nn := providers.GetNamespacedName(o, "zookeeper")

	labels := o.GetLabels()
	labels["env-app"] = nn.Name
	labeler := utils.MakeLabeler(nn, labels, o)

	labeler(dd)

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

	dd.Spec.Replicas = utils.Int32(1)
	dd.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	dd.Spec.Template.Spec.Volumes = []core.Volume{
		{
			Name:         nn.Name,
			VolumeSource: volSource,
		},
		{
			Name: "mq-zookeeper-1",
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "mq-zookeeper-2",
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "mq-zookeeper-3",
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		},
	}
	dd.Spec.Template.ObjectMeta.Labels = labels

	envVars := makeEnvVars(&[]envVar{
		{"ZOOKEEPER_INIT_LIMIT", "10"},
		{"ZOOKEEPER_CLIENT_PORT", "32181"},
		{"ZOOKEEPER_SERVER_ID", "1"},
		{"ZOOKEEPER_SERVERS", nn.Name + ":32181"},
		{"ZOOKEEPER_TICK_TIME", "2000"},
		{"ZOOKEEPER_SYNC_LIMIT", "10"},
	})

	ports := []core.ContainerPort{
		{
			Name:          "zookeeper",
			ContainerPort: 2181,
		},
		{
			Name:          "zookeeper-1",
			ContainerPort: 2888,
		},
		{
			Name:          "zookeeper-2",
			ContainerPort: 3888,
		},
	}

	probeHandler := core.Handler{
		Exec: &core.ExecAction{
			Command: []string{
				"echo",
				"ruok",
				"|",
				"nc",
				"127.0.0.1",
				"32181",
				"|",
				"grep",
				"imok",
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
		Name:  nn.Name,
		Image: "quay.io/cloudservices/cp-zookeeper:5.3.2",
		Env:   envVars,
		Ports: ports,
		VolumeMounts: []core.VolumeMount{
			{
				Name:      nn.Name,
				MountPath: "/var/lib/zookeeper",
			},
			{
				Name:      "mq-zookeeper-1",
				MountPath: "/etc/zookeeper/secrets",
			},
			{
				Name:      "mq-zookeeper-2",
				MountPath: "/var/lib/zookeeper/data",
			},
			{
				Name:      "mq-zookeeper-3",
				MountPath: "/var/lib/zookeeper/log",
			},
		},
		LivenessProbe:  &livenessProbe,
		ReadinessProbe: &readinessProbe,
	}

	dd.Spec.Template.Spec.Containers = []core.Container{c}
	dd.Spec.Template.SetLabels(labels)

	servicePorts := []core.ServicePort{
		{
			Name: "zookeeper1", Port: 32181, Protocol: "TCP",
		},
		{
			Name: "zookeeper2", Port: 2888, Protocol: "TCP",
		},
		{
			Name: "zookeeper3", Port: 3888, Protocol: "TCP",
		},
	}

	utils.MakeService(svc, nn, labels, servicePorts, o)
	if usePVC {
		utils.MakePVC(pvc, nn, labels, "1Gi", o)
	}
}
