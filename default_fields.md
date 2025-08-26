# Default Fields

> [!NOTE] 
> This file is generated like [schema.json](schema.json) and
> [config_all_options.yaml](config_all_options.yaml). Every available field for
> every step that adds fields is listed.

> [!TIP]
> Any fields starting with "ACARSProcessor." are common fields added by
> ACARSProcessor to make it easier to work with different modules that do the
> same thing, for example ADSB-Exchange and Tar1090 both generating aircraft
> distances.

## Sources

### ACARS Messages

- ACARSMessage.ASSStatus
- ACARSMessage.Acknowledge
- ACARSMessage.AircraftTailCode
- ACARSMessage.App.ACARSRouterUUID
- ACARSMessage.App.ACARSRouterVersion
- ACARSMessage.App.Name
- ACARSMessage.App.Proxied
- ACARSMessage.App.ProxiedBy
- ACARSMessage.App.Version
- ACARSMessage.BlockID
- ACARSMessage.Channel
- ACARSMessage.ErrorCode
- ACARSMessage.FlightNumber
- ACARSMessage.FrequencyMHz
- ACARSMessage.Label
- ACARSMessage.MessageNumber
- ACARSMessage.MessageText
- ACARSMessage.Mode
- ACARSMessage.Model.DeletedAt.Valid
- ACARSMessage.Model.ID
- ACARSMessage.Processed
- ACARSMessage.SignaldBm
- ACARSMessage.StationID
- ACARSMessage.Timestamp
- ACARSProcessor.ACARSDramaTailNumberLink
- ACARSProcessor.Channel
- ACARSProcessor.FlightNumber
- ACARSProcessor.FrequencyHz
- ACARSProcessor.FrequencyMHz
- ACARSProcessor.From
- ACARSProcessor.ImageLink
- ACARSProcessor.MessageText
- ACARSProcessor.PhotosLink
- ACARSProcessor.SignalLeveldBm
- ACARSProcessor.StationId
- ACARSProcessor.TailCode
- ACARSProcessor.ThumbnailLink
- ACARSProcessor.TrackingLink
- ACARSProcessor.TranslateLink
- ACARSProcessor.UnixTimestamp

### VDLM2 Messages

- ACARSProcessor.ACARSDramaTailNumberLink
- ACARSProcessor.FlightNumber
- ACARSProcessor.FrequencyHz
- ACARSProcessor.FrequencyMHz
- ACARSProcessor.From
- ACARSProcessor.ImageLink
- ACARSProcessor.MessageText
- ACARSProcessor.PhotosLink
- ACARSProcessor.SignalLeveldBm
- ACARSProcessor.StationId
- ACARSProcessor.TailCode
- ACARSProcessor.ThumbnailLink
- ACARSProcessor.TrackingLink
- ACARSProcessor.TranslateLink
- ACARSProcessor.UnixTimestamp
- VDLM2Message.Model.DeletedAt.Valid
- VDLM2Message.Model.ID
- VDLM2Message.Processed
- VDLM2Message.VDL2.AVLC.ACARS.Acknowledge
- VDLM2Message.VDL2.AVLC.ACARS.BlockID
- VDLM2Message.VDL2.AVLC.ACARS.CRCOK
- VDLM2Message.VDL2.AVLC.ACARS.Error
- VDLM2Message.VDL2.AVLC.ACARS.FlightNumber
- VDLM2Message.VDL2.AVLC.ACARS.Label
- VDLM2Message.VDL2.AVLC.ACARS.MessageNumber
- VDLM2Message.VDL2.AVLC.ACARS.MessageNumberSequence
- VDLM2Message.VDL2.AVLC.ACARS.MessageText
- VDLM2Message.VDL2.AVLC.ACARS.Mode
- VDLM2Message.VDL2.AVLC.ACARS.More
- VDLM2Message.VDL2.AVLC.ACARS.Registration
- VDLM2Message.VDL2.AVLC.CR
- VDLM2Message.VDL2.AVLC.Destination.Address
- VDLM2Message.VDL2.AVLC.Destination.Type
- VDLM2Message.VDL2.AVLC.FrameType
- VDLM2Message.VDL2.AVLC.Poll
- VDLM2Message.VDL2.AVLC.RSequence
- VDLM2Message.VDL2.AVLC.SSequence
- VDLM2Message.VDL2.AVLC.Source.Address
- VDLM2Message.VDL2.AVLC.Source.Status
- VDLM2Message.VDL2.AVLC.Source.Type
- VDLM2Message.VDL2.App.ACARSRouterUUID
- VDLM2Message.VDL2.App.ACARSRouterVersion
- VDLM2Message.VDL2.App.Name
- VDLM2Message.VDL2.App.Proxied
- VDLM2Message.VDL2.App.ProxiedBy
- VDLM2Message.VDL2.App.Version
- VDLM2Message.VDL2.BurstLengthOctets
- VDLM2Message.VDL2.FrequencyHz
- VDLM2Message.VDL2.FrequencySkew
- VDLM2Message.VDL2.HDRBitsFixed
- VDLM2Message.VDL2.Index
- VDLM2Message.VDL2.NoiseLevel
- VDLM2Message.VDL2.OctetsCorrectedByFEC
- VDLM2Message.VDL2.SignalLevel
- VDLM2Message.VDL2.Station
- VDLM2Message.VDL2.Timestamp.Microseconds
- VDLM2Message.VDL2.Timestamp.UnixTimestamp

## Annotators

### ADSBExchangeAnnotator

- ACARSProcessor.AircraftDistanceKm
- ACARSProcessor.AircraftDistanceMi
- ACARSProcessor.AircraftGeolocation
- ACARSProcessor.AircraftLatitude
- ACARSProcessor.AircraftLongitude
- ADSBExchangeAnnotator.APITimestamp
- ADSBExchangeAnnotator.AircraftDistanceKm
- ADSBExchangeAnnotator.AircraftDistanceMi
- ADSBExchangeAnnotator.AircraftGeolocation
- ADSBExchangeAnnotator.AircraftGeolocationLatitude
- ADSBExchangeAnnotator.AircraftGeolocationLongitude
- ADSBExchangeAnnotator.CacheTime
- ADSBExchangeAnnotator.Message
- ADSBExchangeAnnotator.ServerProcessingTime
- ADSBExchangeAnnotator.TotalAircraftResults

### OllamaAnnotator

- ACARSProcessor.LLMModelFeedbackText
- ACARSProcessor.LLMProcessedNumber
- ACARSProcessor.LLMProcessedText
- ACARSProcessor.LLMYesNoQuestionAnswer
- OllamaAnnotator.ModelFeedbackText
- OllamaAnnotator.ProcessedNumber
- OllamaAnnotator.ProcessedText
- OllamaAnnotator.YesNoQuestionAnswer

### Tar1090Annotator

- ACARSProcessor.AircraftDistanceKm
- ACARSProcessor.AircraftDistanceMi
- ACARSProcessor.AircraftGeolocation
- ACARSProcessor.AircraftLatitude
- ACARSProcessor.AircraftLongitude
- Tar1090.AircraftDistanceKm
- Tar1090.AircraftDistanceMi
- Tar1090.AircraftGeolocation
- Tar1090.AircraftGeolocationLatitude
- Tar1090.AircraftGeolocationLongitude
- Tar1090.Messages
- Tar1090.Now
