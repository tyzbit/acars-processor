# acars-annotator

A simple daemon that listens to ACARS messages, hydrates them with additional
data via external lookups and then submits the message to a specified receiver
such as a webhook.

Enable annotators and receivers by filling in the required environment
variables for them.

## Available annotators

- ACARS: This will add key/value fields for all data in the original ACARS
  message
- ADS-B Exchange (Only SingleAircraftPositionByRegistration at the moment)

## Available receivers

- New Relic
- Discord
- Custom Webhook - See below for usage

| Environment Variable                | Value                                                       |
| ----------------------------------- | ----------------------------------------------------------- |
| ACARSHUB_HOST                       | The hostname or IP to your acarshub instance                |
| ACARSHUB_PORT                       | The port to connect to your acarshub instance on            |
| ANNOTATE_ACARS                      | Include the original ACARS message, "true" or "false"       |
| FILTER_CRITERIA_INCLUSIVE           | All filters must pass                                       |
| FILTER_CRITERIA_HAS_TEXT            | Include the original ACARS message, "true" or "false"       |
| FILTER_CRITERIA_MATCH_TAIL_CODE     | Include the original ACARS message, "true" or "false"       |
| FILTER_CRITERIA_MATCH_FLIGHT_NUMBER | Include the original ACARS message, "true" or "false"       |
| FILTER_CRITERIA_MATCH_FREQUENCY     | Include the original ACARS message, "true" or "false"       |
| FILTER_CRITERIA_ABOVE_SIGNAL_DBM    | Include the original ACARS message, "true" or "false"       |
| FILTER_CRITERIA_MATCH_STATION_ID    | Include the original ACARS message, "true" or "false"       |
| DISCORD_WEBHOOK_URL                 | URL to a Discord webhook to post messages in a channel      |
| ADBSEXCHANGE_APIKEY                 | Your API Key to adb-s exchange (lite tier is fine)          |
| ADBSEXCHANGE_REFERENCE_GEOLOCATION  | A geolocation to calulate distance from (ex: "0.1,-0.1") \* |
| LOGLEVEL                            | debug, info, warn, error (default "info")                   |
| NEW_RELIC_LICENSE_KEY               | Your New Relic Infra license key (ex: 123456NRAL)           |
| WEBHOOK_URL                         | URL to your custom webhook                                  |
| WEBHOOK_METHOD                      | GET, POST, etc                                              |
| WEBHOOK_HEADERS                     | Headers to send along with the webhook request\*\*          |

\* Required at this time for ADB-S, feel free to set it to "0,0"

\*\* The headers should be in the format `key=value,otherkey=value`

#### Webhooks

In order to define the content for your webhook, edit `receiver_webhook.tpl`
and add the fields and values that you need with
[valid Go template syntax](https://pkg.go.dev/text/template).
An example is provided which shows a very simple webhook payload
that uses annotations from the ACARS annotator.
