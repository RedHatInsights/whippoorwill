# AppConfig Schema

```txt
https://cloud.redhat.com/schemas/clowder-appconfig
```




| Abstract               | Extensible | Status         | Identifiable            | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                  |
| :--------------------- | ---------- | -------------- | ----------------------- | :---------------- | --------------------- | ------------------- | ----------------------------------------------------------- |
| Cannot be instantiated | Yes        | Unknown status | Unknown identifiability | Forbidden         | Allowed               | none                | [schema.json](../../out/schema.json "open original schema") |

## AppConfig Type

unknown ([AppConfig](schema.md))

# AppConfig Definitions

## Definitions group AppConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig"}
```

| Property                              | Type      | Required | Nullable       | Defined by                                                                                                                                                                            |
| :------------------------------------ | --------- | -------- | -------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [privatePort](#privateport)           | `integer` | Optional | cannot be null | [AppConfig](schema-definitions-appconfig-properties-privateport.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/privatePort")           |
| [publicPort](#publicport)             | `integer` | Optional | cannot be null | [AppConfig](schema-definitions-appconfig-properties-publicport.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/publicPort")             |
| [webPort](#webport)                   | `integer` | Optional | cannot be null | [AppConfig](schema-definitions-appconfig-properties-webport.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/webPort")                   |
| [metricsPort](#metricsport)           | `integer` | Required | cannot be null | [AppConfig](schema-definitions-appconfig-properties-metricsport.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/metricsPort")           |
| [metricsPath](#metricspath)           | `string`  | Required | cannot be null | [AppConfig](schema-definitions-appconfig-properties-metricspath.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/metricsPath")           |
| [logging](#logging)                   | `object`  | Required | cannot be null | [AppConfig](schema-definitions-loggingconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/logging")                                  |
| [kafka](#kafka)                       | `object`  | Optional | cannot be null | [AppConfig](schema-definitions-kafkaconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/kafka")                                      |
| [database](#database)                 | `object`  | Optional | cannot be null | [AppConfig](schema-definitions-databaseconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/database")                                |
| [objectStore](#objectstore)           | `object`  | Optional | cannot be null | [AppConfig](schema-definitions-objectstoreconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/objectStore")                          |
| [inMemoryDb](#inmemorydb)             | `object`  | Optional | cannot be null | [AppConfig](schema-definitions-inmemorydbconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/inMemoryDb")                            |
| [featureFlags](#featureflags)         | `object`  | Optional | cannot be null | [AppConfig](schema-definitions-featureflagsconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/featureFlags")                        |
| [endpoints](#endpoints)               | `array`   | Optional | cannot be null | [AppConfig](schema-definitions-appconfig-properties-endpoints.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/endpoints")               |
| [privateEndpoints](#privateendpoints) | `array`   | Optional | cannot be null | [AppConfig](schema-definitions-appconfig-properties-privateendpoints.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/privateEndpoints") |

### privatePort

Defines the private port that the app should be configured to listen on for API traffic.


`privatePort`

-   is optional
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-appconfig-properties-privateport.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/privatePort")

#### privatePort Type

`integer`

### publicPort

Defines the public port that the app should be configured to listen on for API traffic.


`publicPort`

-   is optional
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-appconfig-properties-publicport.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/publicPort")

#### publicPort Type

`integer`

### webPort

Deprecated: Use 'publicPort' instead.


`webPort`

-   is optional
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-appconfig-properties-webport.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/webPort")

#### webPort Type

`integer`

### metricsPort

Defines the metrics port that the app should be configured to listen on for metric traffic.


`metricsPort`

-   is required
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-appconfig-properties-metricsport.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/metricsPort")

#### metricsPort Type

`integer`

### metricsPath

Defines the path to the metrics server that the app should be configured to listen on for metric traffic.


`metricsPath`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-appconfig-properties-metricspath.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/metricsPath")

#### metricsPath Type

`string`

### logging

Logging Configuration


`logging`

-   is required
-   Type: `object` ([LoggingConfig](schema-definitions-loggingconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-loggingconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/logging")

#### logging Type

`object` ([LoggingConfig](schema-definitions-loggingconfig.md))

### kafka

Kafka Configuration


`kafka`

-   is optional
-   Type: `object` ([Details](schema-definitions-kafkaconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-kafkaconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/kafka")

#### kafka Type

`object` ([Details](schema-definitions-kafkaconfig.md))

### database

Database Configuration


`database`

-   is optional
-   Type: `object` ([DatabaseConfig](schema-definitions-databaseconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/database")

#### database Type

`object` ([DatabaseConfig](schema-definitions-databaseconfig.md))

### objectStore

Object Storage Configuration


`objectStore`

-   is optional
-   Type: `object` ([Details](schema-definitions-objectstoreconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstoreconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/objectStore")

#### objectStore Type

`object` ([Details](schema-definitions-objectstoreconfig.md))

### inMemoryDb

In Memory DB Configuration


`inMemoryDb`

-   is optional
-   Type: `object` ([Details](schema-definitions-inmemorydbconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-inmemorydbconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/inMemoryDb")

#### inMemoryDb Type

`object` ([Details](schema-definitions-inmemorydbconfig.md))

### featureFlags

Feature Flags Configuration


`featureFlags`

-   is optional
-   Type: `object` ([Details](schema-definitions-featureflagsconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-featureflagsconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/featureFlags")

#### featureFlags Type

`object` ([Details](schema-definitions-featureflagsconfig.md))

### endpoints




`endpoints`

-   is optional
-   Type: `object[]` ([Details](schema-definitions-dependencyendpoint.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-appconfig-properties-endpoints.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/endpoints")

#### endpoints Type

`object[]` ([Details](schema-definitions-dependencyendpoint.md))

### privateEndpoints




`privateEndpoints`

-   is optional
-   Type: `object[]` ([Details](schema-definitions-privatedependencyendpoint.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-appconfig-properties-privateendpoints.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/AppConfig/properties/privateEndpoints")

#### privateEndpoints Type

`object[]` ([Details](schema-definitions-privatedependencyendpoint.md))

## Definitions group LoggingConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig"}
```

| Property                  | Type     | Required | Nullable       | Defined by                                                                                                                                                                |
| :------------------------ | -------- | -------- | -------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [type](#type)             | `string` | Required | cannot be null | [AppConfig](schema-definitions-loggingconfig-properties-type.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/type")     |
| [cloudwatch](#cloudwatch) | `object` | Optional | cannot be null | [AppConfig](schema-definitions-cloudwatchconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/cloudwatch")            |
| [kafka](#kafka-1)         | `object` | Optional | cannot be null | [AppConfig](schema-definitions-kafkalogconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/kafka")                   |
| [tags](#tags)             | `array`  | Optional | cannot be null | [AppConfig](schema-definitions-loggingconfig-properties-tags.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/tags")     |
| [labels](#labels)         | `array`  | Optional | cannot be null | [AppConfig](schema-definitions-loggingconfig-properties-labels.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/labels") |

### type

Defines the type of logging configuration


`type`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-loggingconfig-properties-type.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/type")

#### type Type

`string`

### cloudwatch

Cloud Watch configuration


`cloudwatch`

-   is optional
-   Type: `object` ([CloudWatchConfig](schema-definitions-cloudwatchconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-cloudwatchconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/cloudwatch")

#### cloudwatch Type

`object` ([CloudWatchConfig](schema-definitions-cloudwatchconfig.md))

### kafka

Kafka based logging config


`kafka`

-   is optional
-   Type: `object` ([KafkaLogConfig](schema-definitions-kafkalogconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-kafkalogconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/kafka")

#### kafka Type

`object` ([KafkaLogConfig](schema-definitions-kafkalogconfig.md))

### tags

List of tags


`tags`

-   is optional
-   Type: `string[]`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-loggingconfig-properties-tags.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/tags")

#### tags Type

`string[]`

### labels

List of Labels


`labels`

-   is optional
-   Type: `object[]` ([Details](schema-definitions-labelconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-loggingconfig-properties-labels.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/labels")

#### labels Type

`object[]` ([Details](schema-definitions-labelconfig.md))

## Definitions group KafkaLogConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaLogConfig"}
```

| Property                | Type     | Required | Nullable       | Defined by                                                                                                                                                                        |
| :---------------------- | -------- | -------- | -------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [topicName](#topicname) | `string` | Required | cannot be null | [AppConfig](schema-definitions-kafkalogconfig-properties-topicname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaLogConfig/properties/topicName") |

### topicName

Kafka Logging Topic name


`topicName`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-kafkalogconfig-properties-topicname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaLogConfig/properties/topicName")

#### topicName Type

`string`

## Definitions group CloudWatchConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig"}
```

| Property                            | Type     | Required | Nullable       | Defined by                                                                                                                                                                                        |
| :---------------------------------- | -------- | -------- | -------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [accessKeyId](#accesskeyid)         | `string` | Required | cannot be null | [AppConfig](schema-definitions-cloudwatchconfig-properties-accesskeyid.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig/properties/accessKeyId")         |
| [secretAccessKey](#secretaccesskey) | `string` | Required | cannot be null | [AppConfig](schema-definitions-cloudwatchconfig-properties-secretaccesskey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig/properties/secretAccessKey") |
| [region](#region)                   | `string` | Required | cannot be null | [AppConfig](schema-definitions-cloudwatchconfig-properties-region.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig/properties/region")                   |
| [logGroup](#loggroup)               | `string` | Required | cannot be null | [AppConfig](schema-definitions-cloudwatchconfig-properties-loggroup.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig/properties/logGroup")               |

### accessKeyId

Defines the access key that the app should use for configuring CloudWatch.


`accessKeyId`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-cloudwatchconfig-properties-accesskeyid.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig/properties/accessKeyId")

#### accessKeyId Type

`string`

### secretAccessKey

Defines the secret key that the app should use for configuring CloudWatch.


`secretAccessKey`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-cloudwatchconfig-properties-secretaccesskey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig/properties/secretAccessKey")

#### secretAccessKey Type

`string`

### region

Defines the region that the app should use for configuring CloudWatch.


`region`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-cloudwatchconfig-properties-region.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig/properties/region")

#### region Type

`string`

### logGroup

Defines the logGroup that the app should use for configuring CloudWatch.


`logGroup`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-cloudwatchconfig-properties-loggroup.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/CloudWatchConfig/properties/logGroup")

#### logGroup Type

`string`

## Definitions group KafkaConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaConfig"}
```

| Property            | Type    | Required | Nullable       | Defined by                                                                                                                                                              |
| :------------------ | ------- | -------- | -------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [brokers](#brokers) | `array` | Required | cannot be null | [AppConfig](schema-definitions-kafkaconfig-properties-brokers.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaConfig/properties/brokers") |
| [topics](#topics)   | `array` | Required | cannot be null | [AppConfig](schema-definitions-kafkaconfig-properties-topics.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaConfig/properties/topics")   |

### brokers

Defines the brokers the app should connect to for Kafka services.


`brokers`

-   is required
-   Type: `object[]` ([Details](schema-definitions-brokerconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-kafkaconfig-properties-brokers.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaConfig/properties/brokers")

#### brokers Type

`object[]` ([Details](schema-definitions-brokerconfig.md))

### topics

Defines a list of the topic configurations available to the application.


`topics`

-   is required
-   Type: `object[]` ([Details](schema-definitions-topicconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-kafkaconfig-properties-topics.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaConfig/properties/topics")

#### topics Type

`object[]` ([Details](schema-definitions-topicconfig.md))

## Definitions group BrokerConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/BrokerConfig"}
```

| Property              | Type      | Required | Nullable       | Defined by                                                                                                                                                                  |
| :-------------------- | --------- | -------- | -------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [hostname](#hostname) | `string`  | Required | cannot be null | [AppConfig](schema-definitions-brokerconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/BrokerConfig/properties/hostname") |
| [port](#port)         | `integer` | Optional | cannot be null | [AppConfig](schema-definitions-brokerconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/BrokerConfig/properties/port")         |

### hostname




`hostname`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-brokerconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/BrokerConfig/properties/hostname")

#### hostname Type

`string`

### port




`port`

-   is optional
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-brokerconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/BrokerConfig/properties/port")

#### port Type

`integer`

## Definitions group TopicConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/TopicConfig"}
```

| Property                        | Type     | Required | Nullable       | Defined by                                                                                                                                                                          |
| :------------------------------ | -------- | -------- | -------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [requestedName](#requestedname) | `string` | Required | cannot be null | [AppConfig](schema-definitions-topicconfig-properties-requestedname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/TopicConfig/properties/requestedName") |
| [name](#name)                   | `string` | Required | cannot be null | [AppConfig](schema-definitions-topicconfig-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/TopicConfig/properties/name")                   |
| [consumerGroup](#consumergroup) | `string` | Optional | cannot be null | [AppConfig](schema-definitions-topicconfig-properties-consumergroup.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/TopicConfig/properties/consumerGroup") |

### requestedName

The name that the app requested in the ClowdApp definition.


`requestedName`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-topicconfig-properties-requestedname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/TopicConfig/properties/requestedName")

#### requestedName Type

`string`

### name

The name of the actual topic on the Kafka server.


`name`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-topicconfig-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/TopicConfig/properties/name")

#### name Type

`string`

### consumerGroup

Defines the consumer group that should be used for the topic.


`consumerGroup`

-   is optional
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-topicconfig-properties-consumergroup.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/TopicConfig/properties/consumerGroup")

#### consumerGroup Type

`string`

## Definitions group DatabaseConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig"}
```

| Property                        | Type      | Required | Nullable       | Defined by                                                                                                                                                                                |
| :------------------------------ | --------- | -------- | -------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [name](#name-1)                 | `string`  | Required | cannot be null | [AppConfig](schema-definitions-databaseconfig-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/name")                   |
| [username](#username)           | `string`  | Required | cannot be null | [AppConfig](schema-definitions-databaseconfig-properties-username.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/username")           |
| [password](#password)           | `string`  | Required | cannot be null | [AppConfig](schema-definitions-databaseconfig-properties-password.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/password")           |
| [hostname](#hostname-1)         | `string`  | Required | cannot be null | [AppConfig](schema-definitions-databaseconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/hostname")           |
| [port](#port-1)                 | `integer` | Required | cannot be null | [AppConfig](schema-definitions-databaseconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/port")                   |
| [adminUsername](#adminusername) | `string`  | Required | cannot be null | [AppConfig](schema-definitions-databaseconfig-properties-adminusername.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/adminUsername") |
| [adminPassword](#adminpassword) | `string`  | Required | cannot be null | [AppConfig](schema-definitions-databaseconfig-properties-adminpassword.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/adminPassword") |
| [rdsCa](#rdsca)                 | `string`  | Optional | cannot be null | [AppConfig](schema-definitions-databaseconfig-properties-rdsca.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/rdsCa")                 |

### name

Defines the database name.


`name`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/name")

#### name Type

`string`

### username

Defines a username with standard access to the database.


`username`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig-properties-username.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/username")

#### username Type

`string`

### password

Defines the password for the standard user.


`password`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig-properties-password.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/password")

#### password Type

`string`

### hostname

Defines the hostname of the database configured for the ClowdApp.


`hostname`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/hostname")

#### hostname Type

`string`

### port

Defines the port of the database configured for the ClowdApp.


`port`

-   is required
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/port")

#### port Type

`integer`

### adminUsername

Defines the pgAdmin username.


`adminUsername`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig-properties-adminusername.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/adminUsername")

#### adminUsername Type

`string`

### adminPassword

Defines the pgAdmin password.


`adminPassword`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig-properties-adminpassword.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/adminPassword")

#### adminPassword Type

`string`

### rdsCa

Defines the CA used to access the database.


`rdsCa`

-   is optional
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-databaseconfig-properties-rdsca.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DatabaseConfig/properties/rdsCa")

#### rdsCa Type

`string`

## Definitions group ObjectStoreBucket

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket"}
```

| Property                          | Type     | Required | Nullable       | Defined by                                                                                                                                                                                      |
| :-------------------------------- | -------- | -------- | -------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [accessKey](#accesskey)           | `string` | Optional | cannot be null | [AppConfig](schema-definitions-objectstorebucket-properties-accesskey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket/properties/accessKey")         |
| [secretKey](#secretkey)           | `string` | Optional | cannot be null | [AppConfig](schema-definitions-objectstorebucket-properties-secretkey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket/properties/secretKey")         |
| [requestedName](#requestedname-1) | `string` | Required | cannot be null | [AppConfig](schema-definitions-objectstorebucket-properties-requestedname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket/properties/requestedName") |
| [name](#name-2)                   | `string` | Required | cannot be null | [AppConfig](schema-definitions-objectstorebucket-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket/properties/name")                   |

### accessKey

Defines the access key for specificed bucket.


`accessKey`

-   is optional
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstorebucket-properties-accesskey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket/properties/accessKey")

#### accessKey Type

`string`

### secretKey

Defines the secret key for the specified bucket.


`secretKey`

-   is optional
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstorebucket-properties-secretkey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket/properties/secretKey")

#### secretKey Type

`string`

### requestedName

The name that was requested for the bucket in the ClowdApp.


`requestedName`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstorebucket-properties-requestedname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket/properties/requestedName")

#### requestedName Type

`string`

### name

The actual name of the bucket being accessed.


`name`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstorebucket-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreBucket/properties/name")

#### name Type

`string`

## Definitions group ObjectStoreConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig"}
```

| Property                  | Type      | Required | Nullable       | Defined by                                                                                                                                                                              |
| :------------------------ | --------- | -------- | -------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [buckets](#buckets)       | `array`   | Optional | cannot be null | [AppConfig](schema-definitions-objectstoreconfig-properties-buckets.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/buckets")     |
| [accessKey](#accesskey-1) | `string`  | Optional | cannot be null | [AppConfig](schema-definitions-objectstoreconfig-properties-accesskey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/accessKey") |
| [secretKey](#secretkey-1) | `string`  | Optional | cannot be null | [AppConfig](schema-definitions-objectstoreconfig-properties-secretkey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/secretKey") |
| [hostname](#hostname-2)   | `string`  | Required | cannot be null | [AppConfig](schema-definitions-objectstoreconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/hostname")   |
| [port](#port-2)           | `integer` | Required | cannot be null | [AppConfig](schema-definitions-objectstoreconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/port")           |
| [tls](#tls)               | `boolean` | Required | cannot be null | [AppConfig](schema-definitions-objectstoreconfig-properties-tls.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/tls")             |

### buckets




`buckets`

-   is optional
-   Type: `object[]` ([Details](schema-definitions-objectstorebucket.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstoreconfig-properties-buckets.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/buckets")

#### buckets Type

`object[]` ([Details](schema-definitions-objectstorebucket.md))

### accessKey

Defines the access key for the Object Storage server configuration.


`accessKey`

-   is optional
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstoreconfig-properties-accesskey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/accessKey")

#### accessKey Type

`string`

### secretKey

Defines the secret key for the Object Storage server configuration.


`secretKey`

-   is optional
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstoreconfig-properties-secretkey.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/secretKey")

#### secretKey Type

`string`

### hostname

Defines the hostname for the Object Storage server configuration.


`hostname`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstoreconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/hostname")

#### hostname Type

`string`

### port

Defines the port for the Object Storage server configuration.


`port`

-   is required
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstoreconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/port")

#### port Type

`integer`

### tls

Details if the Object Server uses TLS.


`tls`

-   is required
-   Type: `boolean`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-objectstoreconfig-properties-tls.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/ObjectStoreConfig/properties/tls")

#### tls Type

`boolean`

## Definitions group FeatureFlagsConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/FeatureFlagsConfig"}
```

| Property                | Type      | Required | Nullable       | Defined by                                                                                                                                                                              |
| :---------------------- | --------- | -------- | -------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [hostname](#hostname-3) | `string`  | Required | cannot be null | [AppConfig](schema-definitions-featureflagsconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/FeatureFlagsConfig/properties/hostname") |
| [port](#port-3)         | `integer` | Required | cannot be null | [AppConfig](schema-definitions-featureflagsconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/FeatureFlagsConfig/properties/port")         |

### hostname

Defines the hostname for the FeatureFlags server


`hostname`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-featureflagsconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/FeatureFlagsConfig/properties/hostname")

#### hostname Type

`string`

### port

Defines the port for the FeatureFlags server


`port`

-   is required
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-featureflagsconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/FeatureFlagsConfig/properties/port")

#### port Type

`integer`

## Definitions group InMemoryDBConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig"}
```

| Property                | Type      | Required | Nullable       | Defined by                                                                                                                                                                          |
| :---------------------- | --------- | -------- | -------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [hostname](#hostname-4) | `string`  | Required | cannot be null | [AppConfig](schema-definitions-inmemorydbconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig/properties/hostname") |
| [port](#port-4)         | `integer` | Required | cannot be null | [AppConfig](schema-definitions-inmemorydbconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig/properties/port")         |
| [username](#username-1) | `string`  | Optional | cannot be null | [AppConfig](schema-definitions-inmemorydbconfig-properties-username.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig/properties/username") |
| [password](#password-1) | `string`  | Optional | cannot be null | [AppConfig](schema-definitions-inmemorydbconfig-properties-password.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig/properties/password") |

### hostname

Defines the hostname for the In Memory DB server configuration.


`hostname`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-inmemorydbconfig-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig/properties/hostname")

#### hostname Type

`string`

### port

Defines the port for the In Memory DB server configuration.


`port`

-   is required
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-inmemorydbconfig-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig/properties/port")

#### port Type

`integer`

### username

Defines the username for the In Memory DB server configuration.


`username`

-   is optional
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-inmemorydbconfig-properties-username.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig/properties/username")

#### username Type

`string`

### password

Defines the password for the In Memory DB server configuration.


`password`

-   is optional
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-inmemorydbconfig-properties-password.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/InMemoryDBConfig/properties/password")

#### password Type

`string`

## Definitions group LabelConfig

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/LabelConfig"}
```

| Property        | Type     | Required | Nullable       | Defined by                                                                                                                                                          |
| :-------------- | -------- | -------- | -------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [name](#name-3) | `string` | Required | cannot be null | [AppConfig](schema-definitions-labelconfig-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LabelConfig/properties/name")   |
| [value](#value) | `string` | Required | cannot be null | [AppConfig](schema-definitions-labelconfig-properties-value.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LabelConfig/properties/value") |

### name

The label name


`name`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-labelconfig-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LabelConfig/properties/name")

#### name Type

`string`

### value

The label value


`value`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-labelconfig-properties-value.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LabelConfig/properties/value")

#### value Type

`string`

## Definitions group DependencyEndpoint

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint"}
```

| Property                | Type      | Required | Nullable       | Defined by                                                                                                                                                                              |
| :---------------------- | --------- | -------- | -------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [name](#name-4)         | `string`  | Required | cannot be null | [AppConfig](schema-definitions-dependencyendpoint-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint/properties/name")         |
| [hostname](#hostname-5) | `string`  | Required | cannot be null | [AppConfig](schema-definitions-dependencyendpoint-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint/properties/hostname") |
| [port](#port-5)         | `integer` | Required | cannot be null | [AppConfig](schema-definitions-dependencyendpoint-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint/properties/port")         |
| [app](#app)             | `string`  | Required | cannot be null | [AppConfig](schema-definitions-dependencyendpoint-properties-app.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint/properties/app")           |

### name

The PodSpec name of the dependent service inside the ClowdApp.


`name`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-dependencyendpoint-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint/properties/name")

#### name Type

`string`

### hostname

The hostname of the dependent service.


`hostname`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-dependencyendpoint-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint/properties/hostname")

#### hostname Type

`string`

### port

The port of the dependent service.


`port`

-   is required
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-dependencyendpoint-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint/properties/port")

#### port Type

`integer`

### app

The app name of the ClowdApp hosting the service.


`app`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-dependencyendpoint-properties-app.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/DependencyEndpoint/properties/app")

#### app Type

`string`

## Definitions group PrivateDependencyEndpoint

Reference this group by using

```json
{"$ref":"https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint"}
```

| Property                | Type      | Required | Nullable       | Defined by                                                                                                                                                                                            |
| :---------------------- | --------- | -------- | -------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [name](#name-5)         | `string`  | Required | cannot be null | [AppConfig](schema-definitions-privatedependencyendpoint-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint/properties/name")         |
| [hostname](#hostname-6) | `string`  | Required | cannot be null | [AppConfig](schema-definitions-privatedependencyendpoint-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint/properties/hostname") |
| [port](#port-6)         | `integer` | Required | cannot be null | [AppConfig](schema-definitions-privatedependencyendpoint-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint/properties/port")         |
| [app](#app-1)           | `string`  | Required | cannot be null | [AppConfig](schema-definitions-privatedependencyendpoint-properties-app.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint/properties/app")           |

### name

The PodSpec name of the dependent service inside the ClowdApp.


`name`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-privatedependencyendpoint-properties-name.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint/properties/name")

#### name Type

`string`

### hostname

The hostname of the dependent service.


`hostname`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-privatedependencyendpoint-properties-hostname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint/properties/hostname")

#### hostname Type

`string`

### port

The port of the dependent service.


`port`

-   is required
-   Type: `integer`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-privatedependencyendpoint-properties-port.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint/properties/port")

#### port Type

`integer`

### app

The app name of the ClowdApp hosting the service.


`app`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-privatedependencyendpoint-properties-app.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/PrivateDependencyEndpoint/properties/app")

#### app Type

`string`
