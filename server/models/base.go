package models

import "pmail/db"

func SyncTables() {
	err := db.Instance.Sync2(&User{})
	if err != nil {
		panic(err)
	}
	err = db.Instance.Sync2(&Email{})
	if err != nil {
		panic(err)
	}
	err = db.Instance.Sync2(&Group{})
	if err != nil {
		panic(err)
	}
	err = db.Instance.Sync2(&Rule{})
	if err != nil {
		panic(err)
	}
	err = db.Instance.Sync2(&UserAuth{})
	if err != nil {
		panic(err)
	}
}
