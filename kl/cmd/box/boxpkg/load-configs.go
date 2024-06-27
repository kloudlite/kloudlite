package boxpkg

type EnvironmentVariable struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}

// func (*fileclient) loadConfig(mm apiclient.MountMap, envs map[string]string) (*mclient.DevboxKlFile, error) {
// 	kf, err := mclient.GetKlFile("")
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}
//
// 	// read kl.yml into struct
// 	klConfig := &mclient.DevboxKlFile{
// 		Packages: kf.Packages,
// 	}
//
// 	kt, err := mclient.GetKlFile("")
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}
//
// 	fm := map[string]string{}
//
// 	for _, fe := range kt.Mounts.GetMounts() {
// 		pth := fe.Path
// 		if pth == "" {
// 			pth = fe.Key
// 		}
//
// 		fm[pth] = mm[pth]
// 	}
//
// 	// return fm, nil
//
// 	ev := map[string]string{}
// 	for k, v := range envs {
// 		ev[k] = v
// 	}
//
// 	for _, ne := range kf.EnvVars.GetEnvs() {
// 		ev[ne.Key] = ne.Value
// 	}
//
// 	klConfig.Env = ev
// 	klConfig.Mounts = fm
//
// 	return klConfig, nil
// }
