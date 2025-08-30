# acars-processor

A daemon that processes steps you configure. Sources like ACARS or VDLM2 add
messages, while Annotators like Ollama or OpenAI add fields to those messages
after calculating or making an external call. Filters stop the message from
being processed further and Receivers are external things acars-processor sends
messages to.

You can use any combination of a Filter, Receiver (labeled as Send in your
config), or Annotator in a step. Annotators will add fields, Filters will
prevent messages from reaching further steps and Receivers will take the
resulting message and Send it somewhere external to ACARS-Processor. See
[Filters](#available-filters), [Annotators](#available-annotators), and
[Receivers](#available-receivers) for more information on what's available. You
can also use `SelectedFields` at the top level in any Filter step to limit what
fields are provided to subsequent steps, and you may also have a separate Filter
step with only `SelectedFields` if you wish with remove fields between other
steps

# General Configuration

By default, it loads config from [config.yaml](config.yaml) and there is a
well-documented schema to help you fill it out as well as a file with all
possible options and helpful comments at
[config_all_options.yaml](config_all_options.yaml).

You can use your own environment variables (ex: `${apikey}`) in the config and
they will be substituted from the environment before the app starts.

> [!NOTE]
>
> It's highly recommended to quote your string values to lower the chance of
> confusing errors in case substitution fails. Unquoted values will be typed by
> YAML implicitly, example `VALUE=1` in a line `SomeValue: ${VALUE}` will
> evaluate to `SomeValue: 1` which will be interpreted as an integer, not a
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

You may also add `Invert: true` to any filter to have the logic inverted. For
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
  - Processing text: "Process this text to remove excessive caps, spelling
    errors and remove anything that isn't prose so it reads naturally and
    logically"

## Available Receivers

- New Relic: Sends custom events to New Relic.

- Discord: Calls a Discord webhook to post messages in a channel.

- Custom Webhook: Calls a webhook however you want - See
  [below for usage](#available-filters)

### Tips for Using Large Language Models for Filtering and Annotating

The model you choose will greatly impact the quality of the filtering and
annotating.

OpenAI's gpt-4o-mini and higher do reasonably well with basic prompts. With
Ollama, I recommend `qwen3:30b-a3b` if you can run it. It uses about 20GB at
runtime for me but is similar in effectiveness to OpenAI's models due to it
being a Mixture of Experts model -- and it has the added advantage that it only
loads the parameters in memory it needs for the query so it is relatively very
fast.

If you're not seeing great results out of your model, be verbose, explicit and
include examples of what you want to see and not see. Try to imagine what the
model already knows (what "prose" is) versus what it was probably not trained on
(message content of ACARS or VDLM2 messages). It's discouraged but you can also
try a different system prompt.

> [!WARNING]
>
> Most users will never need to override the system prompt.

The User Prompt is where you should put your directions. If acars-processor
isn't able to parse a response from the provider, it'll log what it got from the
model at a `DEBUG` level for troubleshooting.

Also note that increasing concurrency can really hammer a configured LLM
provider (at 1 concurrency, there should only ever be one outstanding call from
acars-processor), so make sure your filters and annotations evaluate quickly and
in a consistent amount of time before increasing it. As self-hosted LLMs like
Ollama get overloaded, their response time increases which further contributes
to being overloaded.

#### Templating for Receivers

Discord, Mastodon and Webhook receivers support templating, which means you
provide a string with some value references and they'll replace those references
with the information from your messages.

Refer to [this page](https://pkg.go.dev/text/template#section-documentation) for
what you can do with Go templating. For example, for Mastodon you could set
`PostGoTemplate` like this:

```
  PostGoTemplate: |
    Tail: {{ index . "ACARSProcessor.TailCode" }}
    Step: {{ index . ".ACARSProcessor.StepNumber" }}
```

Make sure to annotate before this step and select all of the fields you want to
use in previous steps, otherwise the value will be `<no value>` when it's sent
off, **and there is no checking that the values are present before calling the
receiver**.

> [!NOTE]
>
> If you're familiar with Go templating ("text/template"), you'll know that
> `{{ .value }}` notation is possible and cleaner however since the fields
> always have periods in them, the parser will not correctly insert the values
> when used this way since it assumes periods means accessing a struct of the
> field, which is not how APMessages are structured (`map[string]any`)

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
