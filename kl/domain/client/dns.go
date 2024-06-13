package client

func BackupDns() ([]string, error) {

	ed, err := GetExtraData()
	if err != nil {
		return nil, err
	}

	return ed.BackupDns, nil
}

func SetBackupDns(dns []string) error {

	ed, err := GetExtraData()
	if err != nil {
		return err
	}

	ed.BackupDns = dns

	return SaveExtraData(ed)
}
