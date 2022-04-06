package framework

func MakeFramework(cfg *Config) (fm *Framework, e error) {
	errors.HandleErr(&e)
	// ctx := context.Background()
	// mongoCli, e := db.MakeMongoClient(ctx, cfg.MongoDB.Uri, cfg.MongoDB.Db)

	// mongoCli.Collection("users").Drop(ctx)

	fm = &Framework{
		start: func() error {
			return nil
		},
	}
	return
}
