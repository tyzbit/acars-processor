package main

type ACARSHandlerAnnotator struct {
}

func (a ACARSHandlerAnnotator) Name() string {
	return "ACARS"
}

// Interface function to satisfy ACARSHandler
func (a ACARSHandlerAnnotator) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	annotation = Annotation{
		"acarsFrequencyMHz":     m.FrequencyMHz,
		"acarsChannel":          m.Channel,
		"acarsErrorCode":        m.ErrorCode,
		"acarsSignaldBm":        m.SignaldBm,
		"acarsTimestamp":        m.Timestamp,
		"acarsAppName":          m.App.Name,
		"acarsAppVersion":       m.App.Version,
		"acarsAppProxied":       m.App.Proxied,
		"acarsAppProxiedBy":     m.App.ProxiedBy,
		"acarsAppRouterVersion": m.App.ACARSRouterVersion,
		"acarsAppRouterUUID":    m.App.ACARSRouterUUID,
		"acarsStationID":        m.StationID,
		"acarsASSStatus":        m.ASSStatus,
		"acarsMode":             m.Mode,
		"acarsLabel":            m.Label,
		"acarsBlockID":          m.BlockID,
		"acarsAcknowledge":      m.Acknowledge,
		"acarsAircraftTailCode": m.AircraftTailCode,
		"acarsMessageText":      m.MessageText,
		"acarsMessageNumber":    m.MessageNumber,
		"acarsFlightNumber":     m.FlightNumber,
		"acarsExtraURL":         FlightAwareRoot + m.AircraftTailCode,
		"acarsExtraPhotos":      FlightAwarePhotos + m.AircraftTailCode,
	}
	return annotation
}
