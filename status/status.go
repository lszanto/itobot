package status

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// AddStatusTomorrow sets what a user is doing tomorrow and places it in the bucket
func AddStatusTomorrow(db *bolt.DB, user string, location string) {
	bucket := time.Now().AddDate(0, 0, 1).Format("02.01.2006")
	addStatusToBucket(db, user, location, bucket)
}

// AddStatusToday sets what a user is doing today and places it in the bucket
func AddStatusToday(db *bolt.DB, user string, location string) {
	bucket := time.Now().Format("02.01.2006")
	addStatusToBucket(db, user, location, bucket)
}

// GetLocationsFromBucket returns a list of what people are doing from a given bucket
func GetLocationsFromBucket(db *bolt.DB, bucket string) string {
	checkCreateBucket(db, bucket)

	locations := ""

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			locations = locations + string(k) + ": " + string(v) + "\n"
		}

		return nil
	})

	if locations == "" {
		locations = "Sorry! Nobody has let us know their locations for the requested day"
	}

	return locations
}

func addStatusToBucket(db *bolt.DB, user string, location string, bucket string) {
	checkCreateBucket(db, bucket)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(user), []byte(location))

		return err
	})
}

func checkCreateBucket(db *bolt.DB, bucket string) {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket %s", err)
		}

		return nil
	})
}
