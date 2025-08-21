# acars-processor

A Daemon that processes ACARS or VDLM2 messages from ACARSHub in any numeber
of steps. You can use any combination of a Filter, Receiver(refferred to
as Send in your config), or Annotator in a step. Annotators will add fields,
Filters will prevent messages from reaching further steps and Receivers
will take the resulting message and Send it somewhere external to
ACARS-Processor. See
[Filters](#available-filters),
[Annotators](#available-annotators), and
[Receivers](#available-receivers) for more information on what's available.

Default config it reads from is [config.yaml](config.yaml) and there is a
well-documented schema to help you fill it out as well as a
[config_all_options.yaml](config_all_options.yaml) file with every possible
option demonstrated, with comments. See
[General Configuration](#general-configuration).

## Available filters

See the configuration section, but at a high level:

- Builtin: Filter on aspects of the message such as if an emergency was
  specified or if the message has additional message text.
- Ollama: Provide a yes/no or affirmative/negative prompt and Ollama will
  evalutate the message and decide if it should be filtered.
- OpenAI: Provide a yes/no or affirmative/negative prompt and OpenAI will
  evalutate the message and decide if it should be filtered.

### A Note on Filters

Filters fail **CLOSED** by default which means if they fail (only when something
goes wrong, not if the message doesn't pass the filter),
**by default they do not filter the message**. This could overwhelm subsequent
steps.

## Available annotators

- ADS-B Exchange: Adds geolocation information for the transmitter of the
  message and optionally calculates distance to a configurable geolocation.

- Tar1090: Adds a lot of information from a tar1090 instance including location.
  It's advised to use one running in the same geographical location as the
  ACARS/VDLM2 receiver.

- Ollama: Uses Ollama and it will return a processed response based on your
  instructions. You can ask it to make a numerical evaluation, such as "How
  much danger does this message indicate?". You can also ask a question about
  the message. Both the question and the numerical evaluation can annotate and
  filter in one step (if filtered, it won't produce any fields from the
  annotator). This is likely not as effective as just using the Ollama filter
  itself.

## Available receivers

- New Relic: Sends custom events to New Relic
- Discord: Calls a Discord webhook to post messages in a channel.
- Custom Webhook: Calls a webhook however you want - See below for usage

### General Configuration

Check [config_all_options.yaml](config_all_options.yaml) for all possible
settings and illustrative values.

You can use environment variables (`${apikey}`) in the config and they will
be substituted from the environment before the app starts. It's highly
recommended to quote your values in case substitution fails so you don't
chase misleading errors with blank values that cause YAML parsing issues.

### A Note on Using Large Language Models for Filters

OpenAI's gpt3.5 and higher do well with the system prompt. With OllamaFilter,
the model you choose will greatly impact the quality of the filtering.
I recommend `gemma3:4b`. It uses about 8GB at runtime but is similar in
effectiveness to OpenAI's models.

If you're not seeing great results out of your model, be verbose, explicit and
include examples of what you want to see and not see. you can also try
a different one or try overriding the system prompt with
`FILTER_OLLAMA_SYSTEM_PROMPT`. If acars-processor isn't able to pull a JSON
object from the response, it'll log what it got from the model at a
`DEBUG` level for troubleshooting. If your performance isn't great,
reduce `FILTER_OLLAMA_MAX_PREDICTION_TOKENS` and/or increase
`ACARSHUB_MAX_CONCURRENT_REQUESTS_PER_SUBSCRIBER` as well as review your OllamaFilter
configuration for improvements (such as number of parallel requests)

#### Webhooks

In order to define the payload for your webhook, edit `receiver_webhook.tpl`
and add the fields and values that you need with
[valid Go template syntax](https://pkg.go.dev/text/template).
An example is provided which shows a very simple webhook payload
that uses annotations from the ACARS annotator.

# Contributing

First off, thanks for your interest! All contributions are welcome. Here's
some info to help you get a good start:

- Check go.mod for which Go version we're using. As of this writing, that's 1.24
- Install the pre-commit hook by installing
  [pre-commit](https://pre-commit.com/#install) and then running
  `pre-commit install` in the root directory of the repo.
- If you use VSCode, there's already an example launch config for debugging.
- Any PR that seems like it has LLM output will be closed. Any PR where you are
  somehow not familiar with the code you're trying to commit will be closed.

Colors for log messages should follow the general guide below to help with
accessibility.

| Function   | Type of event                         | Color         | Example message                                      |
| ---------- | ------------------------------------- | ------------- | ---------------------------------------------------- |
| Success    | A likely desired state is achieved.   | Green         | Connected successfully                               |
| Content    | Meaningful info to the user.          | Magenta       | 10 filters enabled                                   |
| Note       | Important info.                       | Cyan          | No info back from annotators                         |
| Attention  | Information about a possible problem. | Yellow        | No receivers configured                              |
| Aside      | Less important, perhaps verbose info. | Grey          | "++86501,N8867Q,B7378MAX,250608,WN0393...."          |
| Emphasised | Output or results.                    | Bold & Italic | "Filtering due to excessive use of exclamations!!!!" |
| Custom     | Special and specific. Discouraged.    | Any           | N/A                                                  |
