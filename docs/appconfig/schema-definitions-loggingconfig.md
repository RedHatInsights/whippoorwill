# LoggingConfig Schema

```txt
https://cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig
```

Logging Configuration


| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                    |
| :------------------ | ---------- | -------------- | ------------ | :---------------- | --------------------- | ------------------- | ------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Allowed               | none                | [schema.json\*](../../out/schema.json "open original schema") |

## LoggingConfig Type

`object` ([LoggingConfig](schema-definitions-loggingconfig.md))

# LoggingConfig Properties

| Property                  | Type     | Required | Nullable       | Defined by                                                                                                                                                                |
| :------------------------ | -------- | -------- | -------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [type](#type)             | `string` | Required | cannot be null | [AppConfig](schema-definitions-loggingconfig-properties-type.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/type")     |
| [cloudwatch](#cloudwatch) | `object` | Optional | cannot be null | [AppConfig](schema-definitions-cloudwatchconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/cloudwatch")            |
| [kafka](#kafka)           | `object` | Optional | cannot be null | [AppConfig](schema-definitions-kafkalogconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/kafka")                   |
| [tags](#tags)             | `array`  | Optional | cannot be null | [AppConfig](schema-definitions-loggingconfig-properties-tags.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/tags")     |
| [labels](#labels)         | `array`  | Optional | cannot be null | [AppConfig](schema-definitions-loggingconfig-properties-labels.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/labels") |

## type

Defines the type of logging configuration


`type`

-   is required
-   Type: `string`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-loggingconfig-properties-type.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/type")

### type Type

`string`

## cloudwatch

Cloud Watch configuration


`cloudwatch`

-   is optional
-   Type: `object` ([CloudWatchConfig](schema-definitions-cloudwatchconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-cloudwatchconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/cloudwatch")

### cloudwatch Type

`object` ([CloudWatchConfig](schema-definitions-cloudwatchconfig.md))

## kafka

Kafka based logging config


`kafka`

-   is optional
-   Type: `object` ([KafkaLogConfig](schema-definitions-kafkalogconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-kafkalogconfig.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/kafka")

### kafka Type

`object` ([KafkaLogConfig](schema-definitions-kafkalogconfig.md))

## tags

List of tags


`tags`

-   is optional
-   Type: `string[]`
-   cannot be null
-   defined in: [AppConfig](schema-definitions-loggingconfig-properties-tags.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/tags")

### tags Type

`string[]`

## labels

List of Labels


`labels`

-   is optional
-   Type: `object[]` ([Details](schema-definitions-labelconfig.md))
-   cannot be null
-   defined in: [AppConfig](schema-definitions-loggingconfig-properties-labels.md "https&#x3A;//cloud.redhat.com/schemas/clowder-appconfig#/definitions/LoggingConfig/properties/labels")

### labels Type

`object[]` ([Details](schema-definitions-labelconfig.md))
