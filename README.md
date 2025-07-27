# acars-processor

A simple daemon that, in order:

1.  Listens to ACARS/VDLM2 messages from ACARSHub,
    optionally saving them to resume or review.
2.  Filters them according to various criteria from various providers, including
    LLMs
3.  Adds additional information (annotates) using lookups from messages, like
    aircraft tail numbers.
4.  Submits the message to one or more specified receivers such as a Discord
    webhook or others.

Configuration is done with `config.yaml` and there is a schema to help you
fill it out. See below.

## Available filters

See the configuration section, but at a high level:

- ACARS and VDLM2: Just message similarity at the moment. Many more in `Generic`
- Generic: Filter on aspects of the message such as if an emergency was
  specified or if the message has additional message text.
- Ollama: Provide a yes/no or affirmative/negative prompt and Ollama will
  evalutate the message and decide if it should be filtered.
- OpenAI: Provide a yes/no or affirmative/negative prompt and OpenAI will
  evalutate the message and decide if it should be filtered.

### A Note on Filters

Filters fail **CLOSED** by default which means if they fail (only when something
goes wrong), **by default they do not filter the message**.

## Available annotators

- ACARS: This will add key/value fields for all data in the original ACARS
  message

- VDLM2: Same as above but for VDLM2 messages

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

Check `config_example.yaml` for all possible settings and illustrative values.
You can duplicate it to `config.yaml` and edit it or copy it but only keep the
first line. This will let you auto-complete the file if your editor supports it.

You can use environment variables (`${apikey}`) in the config and they will
be substituted from the environment before the app starts. It's highly
recommended to quote your values in case substitution fails so you don't
chase misleading errors.

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
- LLM output is not permitted to be added to this codebase. Any PR that seems
  like it has LLM output will be closed. Any PR where you do not understand the
  code you're trying to commit will be closed.

Colors for log messages should follow the general guide below to help with
accessibility.

| Function   | Type of event                         | Color         | Example message                                      |
| ---------- | ------------------------------------- | ------------- | ---------------------------------------------------- |
| Success    | A likely desired state is achieved.   | Green         | Connected successfully                               |
| Content    | Meaningful info to the user.          | Magenta       | 10 filters enabled                                   |
| Note       | Important info.                       | Cyan          | No info back from annotators                         |
| Attention  | Information about a possible problem. | Yellow        | No receivers configured                              |
| Aside      | Less important, perhaps verbose info. | Grey          | "++86501,N8867Q,B7378MAX,250608,WN0393...."          |
| Emphasized | Output or results.                    | Bold & Italic | "Filtering due to excessive use of exclamations!!!!" |
| Custom     | Special and specific. Discouraged.    | Any           | N/A                                                  |
