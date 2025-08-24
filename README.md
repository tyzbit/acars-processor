# acars-processor

A Daemon that processes Sources like ACARS or VDLM2 add messages, while
Annotators add fields to those messages after calculating or making a lookup.
Filters stop the message from being processed further and Receivers are external
things acars-processor can send a completed message to.

You can use any combination of a Filter, Receiver (labeled as Send in your
config), or Annotator in a step. Annotators will add fields, Filters will
prevent messages from reaching further steps and Receivers will take the
resulting message and Send it somewhere external to ACARS-Processor. See
[Filters](#available-filters), [Annotators](#available-annotators), and
[Receivers](#available-receivers) for more information on what's available. You
can also use `SelectedFields` in any Filter or Annotate step to limit what
fields are provided to subsequent steps, and you may also have a separate step
with only `SelectedFields` if you wish.

# General Configuration

By default, it loads config from [config.yaml](config.yaml) and there is a
well-documented schema to help you fill it out as well as a file with all
possible options and helpful comments at
[config_all_options.yaml](config_all_options.yaml).

You can use your own environment variables (ex: `${apikey}`) in the config and
they will be substituted from the environment before the app starts.

> [!NOTE] It's highly recommended to quote your string values to lower the
> chance of confusing errors in case substitution fails. Unquoted values will be
> typed by YAML implicitly, example `VALUE=1` in a line `SomeValue: ${VALUE}`
> will evaluate to `SomeValue: 1` which will be interpreted as an integer, not a
> string.

## Available Fields

See [default_fields.md](default_fields.md) for a list of all fields available
from all Sources (eg. ACARS/VDLM2) and Annotators. These are the only modules
that produce fields.

## Available Filters

- Builtin: Filter on aspects of the message such as if an emergency was
  specified or if the message has additional message text.

- Ollama: Provide a yes/no or affirmative/negative prompt and Ollama will
  evalutate the message and decide if it should be filtered - always filtering
  if yes.

- OpenAI: Provide a yes/no or affirmative/negative prompt and OpenAI will
  evalutate the message and decide if it should be filtered - always filtering
  if yes.

### A Note on Filters

Filters fail **CLOSED** by default which means if they fail (only when something
goes wrong, not if the message doesn't pass the filter), **by default they do
not filter the message**. This could overwhelm subsequent steps. Use
`FilterOnFailure` to change this behavior.

You an also add `Invert: true` to any filter to have the logic inverted. For
example, a filter that has `DictionaryPhraseLengthMinimum: 4` and `Invert: true`
would only forward messages that **do not** have 4 sequential dictionary words
in it.

## Available Annotators

- ADS-B Exchange: Adds geolocation information for the transmitter of the
  message and optionally calculates distance to a configurable geolocation.

- Tar1090: Adds information from a tar1090 instance including location. It's
  advised to use one running in the same geographical location as the
  ACARS/VDLM2 receiver so it's more likely location information will be
  available for any given aircraft.

- Ollama: Uses Ollama with a model of your choosing and it will return a set of
  fields with different purposes:

  - LLMModelFeedbackText: The model will add any comments resulting from the
    prompt here
  - LLMProcessedNumber: If you ask for a numerical evaluation, this will be the
    answer. It must be an integer without decimals.
  - LLMProcessedText: If you ask to transform the message, it will be returned
    here.
  - LLMYesNoQuestionAnswer: Yes or no questions, true or false respectively.

  Examples:

  - Numerical evaluation: "On a scale of 1-100, how angry or frustrated does
    this message seem?"

  - Yes/no question about the message: "Is this message above 50 in terms of the
    anger rating?" (combined with the first example).
  - Processing text: Process his text to remove excessive caps, spelling errors
    and remove anything that isn't prose so it reads naturally and logically"

## Available Receivers

- New Relic: Sends custom events to New Relic.

- Discord: Calls a Discord webhook to post messages in a channel.

- Custom Webhook: Calls a webhook however you want - See
  [below for usage](#available-filters)

### A Note on Using Large Language Models for Filtering and Annotating

The model you choose will greatly impact the quality of the filtering.

OpenAI's gpt-4o-mini and higher do reasonably well with basic prompts. With the
Ollama filter, I recommend `qwen3:30b-a3b` if you can run it. It uses about 20GB
at runtime but is similar in effectiveness to OpenAI's models due to it being a
Mixture of Experts model -- and it has the added advantage that it only loads
the parameters in memory it needs for the query so it is relatively very fast.

If you're not seeing great results out of your model, be verbose, explicit and
include examples of what you want to see and not see. Try to guess at what the
model knows (what prose is) versus what it was probably not trained on (message
content of ACARS or VDLM2 messages). It's discouraged but you can also try a
different system prompt take note -- **most users will never need to mess with
the system prompt**. The User Prompt is where you should put your directions. If
acars-processor isn't able to parse a response from the provider, it'll log what
it got from the model at a `DEBUG` level for troubleshooting.

Also note that increasing concurrency can really hammer a configured LLM
provider (at 1 concurrency, there should only ever be one outstanding call from
acars-processor), so make sure your filters are robust before increasing it.

#### Webhooks

In order to define the payload for your webhook, edit `receiver_webhook.tpl` and
add the fields and values that you need with
[valid Go template syntax](https://pkg.go.dev/text/template). An example is
provided which shows a very simple webhook payload that uses annotations from
the ACARS annotator.

# Contributing

First off, thanks for your interest! All contributions are welcome. Here's some
info to help you get a good start:

- Check `go.mod` for which Go version we're using. As of this writing, that's
  `1.24`
- Install the pre-commit hook by installing
  [pre-commit](https://pre-commit.com/#install) and then running
  `pre-commit install` in the root directory of the repo.
- If you use VSCode, there's already an example launch config for debugging.
- Any PR that seems like it has ANY unrefined LLM output will be closed. Any PR
  where you can't answer questions about the code you're trying to commit will
  be closed.

# Colors

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
