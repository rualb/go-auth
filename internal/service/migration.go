package service

func mustCreateRepository(appService AppService) {

	db := appService.Repository() // not full inited

	for _, x := range []any{
		&UserAccount{},
		&VaultKey{},
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
		res := db.Driver().First(key)
		if res.RowsAffected == 0 {
			key, err := NewVaultKey()
			if err != nil {
				panic(err)
			}
			db.Create(key)
		}
	}

	// {
	// 	user := UserAccount{}
	// 	// result := db.Limit(1).Find(&user)
	// 	result := db.First(&user)
	// 	if result.RowsAffected < 1 {
	// 		user, err := NewUserAccount()
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		user.Username = "+99450123456789"
	// 		user.PhoneNumber = "+99450123456789"
	// 		user.Email = "+99450123456789"
	// 		user.NormalizedEmail = "+99450123456789"
	// 		err = user.SetPassword("Qq123456")

	// 		if err != nil {
	// 			panic(fmt.Errorf("error on password hash: %v", err))
	// 		}

	// 		db.Create(user)
	// 	}
	// }

}
