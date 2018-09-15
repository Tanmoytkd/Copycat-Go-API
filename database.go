package main

import (
	"database/sql"
	"fmt"
)

func connectDatabase() *sql.DB {
	const (
		host                    = "localhost"
		database                = "copycat"
		user                    = "tanmoy"
		password                = "jwjHr4RqGq0MOxpu@"
		accountsTableName       = "accounts"
		collabSessionsTableName = "collab_sessions"
	)

	// Initialize connection string.
	var connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user, password, host, database)

	db, dbErr := sql.Open("mysql", connectionString);
	checkError(dbErr)

	err := db.Ping()
	checkError(err)

	_, err = db.Exec("CREATE TABLE  IF NOT EXISTS  `" + database +
		"`.`" + accountsTableName + "` ( `id` INT NOT NULL AUTO_INCREMENT , " +
		"`name` VARCHAR(40) NOT NULL , `email` VARCHAR(40) NOT NULL ," +
		" `password` TEXT NOT NULL , `token` TEXT NULL , `token_expiration` DATETIME ," +
		" PRIMARY KEY (`id`)) ENGINE = InnoDB CHARSET=utf8 COLLATE utf8_general_ci;")
	checkError(err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `" + database + "`.`" + collabSessionsTableName +
		"` ( `id` INT NOT NULL AUTO_INCREMENT , `name` VARCHAR(40) , " +
		"`code` VARCHAR(40) NOT NULL , `password` TEXT NOT NULL, `data` TEXT NOT NULL ," +
		" `hash` TEXT, PRIMARY KEY (`id`)) ENGINE = InnoDB CHARSET=utf8 COLLATE " +
		"utf8_general_ci;")
	checkError(err)

	return db;
}
