# acars-annotator
A simple daemon that listens to ACARS messages, hydrates them with additional
data via external lookups and then submits the message to a specified receiver
such as a webhook.

## Available annotators
- ADS-B Exchange (Only SingleAircraftPositionByRegistration at the moment)

## Available receivers
- Custom Webhook
- New Relic
