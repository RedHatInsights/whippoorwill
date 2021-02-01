# KafkaLogConfig Schema

```txt
https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaLogConfig
```

Kafka based logging config


| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                    |
| :------------------ | ---------- | -------------- | ------------ | :---------------- | --------------------- | ------------------- | ------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Allowed               | none                | [schema.json\*](../../out/schema.json "open original schema") |

## KafkaLogConfig Type

`object` ([KafkaLogConfig](schema-definitions-kafkalogconfig.md))

# KafkaLogConfig Properties

| Property                | Type     | Required | Nullable       | Defined by                                                                                                                                                                        |
| :---------------------- | -------- | -------- | -------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [topicName](#topicname) | `string` | Required | cannot be null | [AppConfig](schema-definitions-kafkalogconfig-properties-topicname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaLogConfig/properties/topicName") |

## topicName

Kafka Logging Topic name


`topicName`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-kafkalogconfig-properties-topicname.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/KafkaLogConfig/properties/topicName")

### topicName Type

`string`
