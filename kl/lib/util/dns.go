package util

func ActiveDns() ([]string, error) {
	file, err := GetContextFile()

	if err != nil {
		return nil, err
	}

	// if len(file.DNS) == 0 {
	// 	return nil,
	// 		errors.New("no active dns found")
	// }

	return file.DNS, nil
}

func SetActiveDns(dns []string) error {
	file, err := GetContextFile()
	if err != nil {
		return err
	}
	file.DNS = dns
	return WriteContextFile(*file)
}
