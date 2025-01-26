package service

import xlog "go-auth/internal/util/utillog"

func mustCreateRepository(appService AppService) {

	db := appService.Repository() // not full inited

	for _, x := range []any{
		&UserAccount{},
		&VaultKey{},
		&UserSession{},
	} {

		if err := db.AutoMigrate(x); err != nil {
			panic(err)
		}

	}

	mustInitRepositoryMasterData(appService) // not full inited

}

func mustInitRepositoryMasterData(appService AppService) {

	// TODO add valut keys if not exists

	db := appService.Repository() // not full inited

	{
		key := &VaultKey{}
		res := db.Model(key).Limit(1).Find(key)
		if res.RowsAffected == 0 {
			key, err := NewVaultKey()
			if err != nil {
				panic(err)
			}
			res := db.Create(key)

			if res.Error != nil {
				xlog.Panic("new vault key: %v", res.Error)
			}
		}
	}

}
