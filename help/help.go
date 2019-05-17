package main

import (
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tomorrowDate := time.Now().AddDate(0, 0, 1).Format("02.01.2006")

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(tomorrowDate))
		if err != nil {
			return fmt.Errorf("create bucket %s", err)
		}

		return nil
	})

	user := "Luke"
	location := "home"

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tomorrowDate))
		err := b.Put([]byte(user), []byte(location))

		return err
	})

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tomorrowDate))
		v := b.Get([]byte(user))
		fmt.Printf("The answer is %s\n", v)
		return nil
	})
}
