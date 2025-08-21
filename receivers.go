package main

func (r ReceiverStep) Send(m APMessage) error {
	receivers := []Receiver{
		r.Discord,
		r.NewRelic,
		r.Webhook,
	}
	var err error
	for _, a := range receivers {
		if !a.Configured() {
			continue
		}
		err = a.Send(m)
		if err != nil {
			return err
		}
	}
	return nil
}
