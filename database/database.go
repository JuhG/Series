package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	_"fmt"
)

func Series_get_rate(name string) {

	var (
		header string
		resp   string
	)

	//username,dbconnect,dbname
	dbinfo := "root:@tcp(127.0.0.1:3306)/series"
	db, err := sql.Open("mysql",
		dbinfo)

	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("select series_name, rating from series where series_name = ?", name)

	if err != nil {
		log.Fatal(err)
	}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&header, &resp)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(header, resp)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}


	defer db.Close()
}
