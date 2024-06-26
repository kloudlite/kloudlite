package client

import "os"

func InsideBox() bool {
	s, ok := os.LookupEnv("IN_DEV_BOX")
	if !ok {
		return false
	}

	return s == "true"
}

// func GetWorkspacePath() (string, error) {
// 	s, ok := os.LookupEnv("KL_WORKSPACE")
// 	if !ok {
// 		dir, err := os.Getwd()
// 		if err != nil {
// 			return "", err
// 		}
// 		return dir, nil
// 	}
//
// 	return s, nil
// }
