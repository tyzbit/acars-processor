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

### General Configuration

| Environment Variable | Value                                            |
| -------------------- | ------------------------------------------------ |
| ACARSHUB_HOST        | The hostname or IP to your acarshub instance     |
| ACARSHUB_PORT        | The port to connect to your acarshub instance on |
| LOGLEVEL             | debug, info, warn, error (default "info")        |

### Annotators

| Environment Variable               | Value                                                       |
| ---------------------------------- | ----------------------------------------------------------- |
| ANNOTATE_ACARS                     | Include the original ACARS message, "true" or "false"       |
| ADBSEXCHANGE_APIKEY                | Your API Key to adb-s exchange (lite tier is fine)          |
| ADBSEXCHANGE_REFERENCE_GEOLOCATION | A geolocation to calulate distance from (ex: "0.1,-0.1") \* |

### Filters

| Environment Variable                | Value                                                                            |
| ----------------------------------- | -------------------------------------------------------------------------------- |
| ACARS_ANNOTATOR_SELECTED_FIELDS     | If this is set, receivers will only receive fields present in this variable \*\* |
| ADSB_ANNOTATOR_SELECTED_FIELDS      | If this is set, receivers will only receive fields present in this variable \*\* |
| FILTER_CRITERIA_INCLUSIVE           | All filters must pass                                                            |
| FILTER_CRITERIA_HAS_TEXT            | Message must have text                                                           |
| FILTER_CRITERIA_MATCH_TAIL_CODE     | Message must match tail code                                                     |
| FILTER_CRITERIA_MATCH_FLIGHT_NUMBER | Message must match flight number                                                 |
| FILTER_CRITERIA_MATCH_FREQUENCY     | Message must have been received on this frequency                                |
| FILTER_CRITERIA_ABOVE_SIGNAL_DBM    | Message must have signal above this                                              |
| FILTER_CRITERIA_MATCH_STATION_ID    | Message must have come from this station                                         |

### Receivers

| Environment Variable  | Value                                                  |
| --------------------- | ------------------------------------------------------ |
| DISCORD_WEBHOOK_URL   | URL to a Discord webhook to post messages in a channel |
| NEW_RELIC_LICENSE_KEY | Your New Relic Infra license key (ex: 123456NRAL)      |
| WEBHOOK_URL           | URL to your custom webhook                             |
| WEBHOOK_METHOD        | GET, POST, etc                                         |
| WEBHOOK_HEADERS       | Headers to send along with the webhook request\*\*\*   |

\* Required at this time for ADB-S, feel free to set it to "0,0"

\*\* Use whatever separator you want, the field just has to be present somewhere
in the variable.

\*\*\* The headers should be in the format `key=value,otherkey=value`

#### Webhooks

In order to define the payload for your webhook, edit `receiver_webhook.tpl`
and add the fields and values that you need with
[valid Go template syntax](https://pkg.go.dev/text/template).
An example is provided which shows a very simple webhook payload
that uses annotations from the ACARS annotator.

#### Example .env
```
ACARSHUB_HOST=192.168.0.100
ACARSHUB_PORT=15550
ANNOTATE_ACARS=true
LOGLEVEL=debug
WEBHOOK_URL=http://webhook
WEBHOOK_METHOD=POST
ACARS_ANNOTATOR_SELECTED_FIELDS=acarsAircraftTailCode,acarsExtraURL,acarsFlightNumber,acarsFrequencyMHz,acarsMessageText
```