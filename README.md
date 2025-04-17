# acars-annotator

A simple daemon that listens to ACARS messages, hydrates them with additional
data via external lookups and then submits the message to a specified receiver
such as a webhook.

Enable annotators and receivers by filling in the required environment
variables for them.

## Available annotators

- ACARS: This will add key/value fields for all data in the original ACARS
  message
- VDLM2: Same as above but for VDLM2 messages
- ADS-B Exchange (Only SingleAircraftPositionByRegistration at the moment)
- Tar1090: Adds a lot of information from a tar1090 instance including location.
  It's advised to use one running in the same geographical location as the
  ACARS/VDLM2 receiver.

## Available receivers

- New Relic
- Discord
- Custom Webhook - See below for usage

### General Configuration

| Environment Variable                            | Value                                                                                                   |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| ACARSHUB_HOST                                   | The hostname or IP to your acarshub instance                                                            |
| ACARSHUB_PORT                                   | The ACARS port to connect to your acarshub instance on                                                  |
| ACARSHUB_MAX_CONCURRENT_REQUESTS_PER_SUBSCRIBER | Maximum number of requests to make to ACARSHub at once (per subscriber, i.e. 1 for ACARS & 1 for VDLM2) |
| ACARSHUB_VDLM2_HOST                             | The hostname or IP to your acarshub instance for VDLM2                                                  |
| ACARSHUB_PORT                                   | The VDLM2 port to connect to your acarshub instance on                                                  |
| LOGLEVEL | debug, info, warn, error (default "info") |

### Annotators

| Environment Variable               | Value                                                                  |
| ---------------------------------- | ---------------------------------------------------------------------- |
| ANNOTATE_ACARS                     | Include the original ACARS message, "true" or "false"                  |
| ANNOTATE_VDLM2                     | Include the original VDLM2 message, "true" or "false"                  |
| ADBSEXCHANGE_APIKEY                | **REQUIRED TO USE** Your API Key to adb-s exchange (lite tier is fine) |
| ADBSEXCHANGE_REFERENCE_GEOLOCATION | A geolocation to calulate distance from (ex: "0.1,-0.1") \*            |
| TAR1090_URL                        | **REQUIRED TO USE** URL to a tar1090 instance                          |
| TAR1090_REFERENCE_GEOLOCATION      | Geolocation to allow the annotator to provide distance metrics \*      |

### Filters

| Environment Variable                             | Value                                                                                                                                                   |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ACARS_ANNOTATOR_SELECTED_FIELDS                  | If this is set, receivers will only receive fields present in this variable from ACARS annotator \*\*                                                   |
| VDLM2_ANNOTATOR_SELECTED_FIELDS                  | If this is set, receivers will only receive fields present in this variable from VDLM2 annotator \*\*                                                   |
| ADSB_ANNOTATOR_SELECTED_FIELDS                   | If this is set, receivers will only receive fields present in this variable from TAR1090 annotator \*\*                                                 |
| TAR1090_ANNOTATOR_SELECTED_FIELDS                | If this is set, receivers will only receive fields present in this variable \*\*                                                                        |
| FILTER_CRITERIA_HAS_TEXT                         | Message must have text                                                                                                                                  |
| FILTER_CRITERIA_MATCH_TAIL_CODE                  | Message must match tail code                                                                                                                            |
| FILTER_CRITERIA_MATCH_FLIGHT_NUMBER              | Message must match flight number                                                                                                                        |
| FILTER_CRITERIA_MATCH_FREQUENCY                  | Message must have been received on this frequency                                                                                                       |
| FILTER_CRITERIA_ABOVE_SIGNAL_DBM                 | Message must have signal above this                                                                                                                     |
| FILTER_CRITERIA_MATCH_STATION_ID                 | Message must have come from this station                                                                                                                |
| FILTER_CRITERIA_DICTIONARY_PHRASE_LENGTH_MINIMUM | Message must have at least this amount of consecutive words (English only at the moment)                                                                |
| FILTER_OLLAMA_URL                                | **REQUIRED TO USE** Full URL to your ollama instance (<scheme>://<host>:<port>)                                                                         |
| FILTER_OLLAMA_TIMEOUT                            | Max time to wait for results from Ollama in seconds (default 60)                                                                                        |
| FILTER_OLLAMA_MODEL                              | **REQUIRED TO USE** The model to use; ex: "llama3.2"                                                                                                    |
| FILTER_OLLAMA_PROMPT                             | **REQUIRED TO USE** Criteria for the model to evaluate the message against \*\*\*\*                                                                     |
| FILTER_OLLAMA_FILTER_ON_FAILURE                  | If Ollama fails, should the message be filtered?                                                                                                        |
| FILTER_OLLAMA_MAX_PREDICTION_TOKENS              | Maximum tokens to return for prediction. Don't go below about 40.                                                                                       |
| FILTER_OLLAMA_MAX_RETRY_ATTEMPTS                 | Maximum number of times to retry calling Ollama PER MESSAGE                                                                                             |
| FILTER_OLLAMA_RETRY_DELAY                        | Seconds to wait between Ollama retries (exponential backoff)                                                                                            |
| FILTER_OLLAMA_SYSTEM_PROMPT                      | By default, `acars-annotator` includes a system prompt that describes what the response should look like. This overrides that.                          |
| FILTER_OPENAI_PROMPT                             | **REQUIRED TO USE** Criteria to evaluate the message, sent to OpenAI \*\*\*\*                                                                           |
| FILTER_OPENAI_APIKEY                             | **REQUIRED TO USE** API key for OpenAI, required for functionality                                                                                      |
| FILTER_OPENAI_MODEL                              | Override the default model (`gpt-4o` by default), see [here](https://pkg.go.dev/github.com/openai/openai-go@v0.1.0-alpha.62#ChatModel) for your options |
| FILTER_OPENAI_SYSTEM_PROMPT                      | By default, `acars-annotator` includes a system prompt that describes what the response should look like. This overrides that.                          |

Filters fail **CLOSED** which means if they fail, they do not filter the message.

### A Note on Using Large Language Models for Filters

OpenAI's gpt3.5 and higher do well with the system prompt. With Ollama,
the model you choose will greatly impact the quality of the filtering.
I recommend `gemma3:4b`. It uses about 8GB at runtime but is similar in
effectiveness to OpenAI's models.

If you're not seeing great results out of your model, be verbose, explicit and
include examples of what you want to see and not see. Uou can also try
a different one or try overriding the system prompt with
`FILTER_OLLAMA_SYSTEM_PROMPT`. If acars-processor isn't able to pull a JSON
object from the response, it'll log what it got from the model at a
`DEBUG` level for troubleshooting.

### Receivers

| Environment Variable  | Value                                                                      |
| --------------------- | -------------------------------------------------------------------------- |
| DISCORD_WEBHOOK_URL   | **REQUIRED TO USE** URL to a Discord webhook to post messages in a channel |
| NEW_RELIC_LICENSE_KEY | **REQUIRED TO USE** Your New Relic Infra license key (ex: 123456NRAL)      |
| WEBHOOK_URL           | **REQUIRED TO USE** URL to your custom webhook                             |
| WEBHOOK_METHOD        | **REQUIRED TO USE** GET, POST, etc                                         |
| WEBHOOK_HEADERS       | Headers to send along with the webhook request \*\*\*                      |

\* If none provided, "0,0" is used.

\*\* Use whatever separator you want, the field just has to be present somewhere
in the variable.

\*\*\* The headers should be in the format `key=value,otherkey=value`

\*\*\*\* Yes or no question works best. Example:
"Does this contain what looks like a bespoke message from one person to another?"

#### Webhooks

In order to define the payload for your webhook, edit `receiver_webhook.tpl`
and add the fields and values that you need with
[valid Go template syntax](https://pkg.go.dev/text/template).
An example is provided which shows a very simple webhook payload
that uses annotations from the ACARS annotator.

#### Example .env

```env
ACARSHUB_HOST=192.168.0.100
ACARSHUB_PORT=15550
ACARSHUB_VDLM2_PORT=15555
ANNOTATE_ACARS=true
ANNOTATE_VDLM2=true
LOGLEVEL=debug
WEBHOOK_URL=http://webhook
WEBHOOK_METHOD=POST
TAR1090_URL=http://tar1090
TAR1090_REFERENCE_GEOLOCATION=34.1340549,-118.3264889
ACARS_ANNOTATOR_SELECTED_FIELDS=acarsAircraftTailCode,acarsExtraURL,acarsFlightNumber,acarsFrequencyMHz,acarsMessageText
VDLM_ANNOTATOR_SELECTED_FIELDS=acarsAircraftTailCode,acarsExtraURL,acarsFlightNumber,vdlm2FrequencyHz,acarsMessageText
```
