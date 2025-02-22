# acars-annotator

A simple daemon that listens to ACARS messages, hydrates them with additional
data via external lookups and then submits the message to a specified receiver
such as a webhook.

Enable annotators and receivers by filling in the required environment
variables for them.

## Available annotators

- ADS-B Exchange (Only SingleAircraftPositionByRegistration at the moment)
  Set

## Available receivers

- New Relic
- Custom Webhook (WIP)

| Environment Variable               | Value                                                       |
| ---------------------------------- | ----------------------------------------------------------- |
| ACARSHUB_HOST                      | The hostname or IP to your acarshub instance                |
| ACARSHUB_PORT                      | The port to connect to your acarshub instance on            |
| ADBSEXCHANGE_APIKEY                | Your API Key to adb-s exchange (lite tier is fine)          |
| ADBSEXCHANGE_REFERENCE_GEOLOCATION | A geolocation to calulate distance from (ex: "0.1,-0.1") \* |
| LOGLEVEL                           | debug, info, warn, error (default "info")                   |
| NEW_RELIC_LICENSE_KEY              | Your New Relic Infra license key (ex: 123456NRAL)           |
| WEBHOOK_URL                        | URL to your custom webhook                                  |
| WEBHOOK_METHOD                     | GET, POST, etc                                              |
| WEBHOOK_HEADERS                    | Headers to send along with the webhook request\*\*          |

\* Required at this time for ADB-S, feel free to set it to "0,0"

\*\* The headers should be in the format `key=value,otherkey=value`
