package client

func ActiveDns() ([]string, error) {

	ed, err := GetExtraData()
	if err != nil {
		return nil, err
	}

	return ed.DNS, nil
}

func SetActiveDns(dns []string) error {

	ed, err := GetExtraData()
	if err != nil {
		return err
	}

	ed.DNS = dns

	return SaveExtraData(ed)
}
