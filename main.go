package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/nlopes/slack"
)

func main() {
	dbFile := getenv("DBFILE")

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	token := getenv("SLACKTOKEN")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			fmt.Print("Event recieved: ")
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				info := rtm.GetInfo()
				botID := "<@" + strings.ToLower(info.User.ID) + ">"
				messageText := strings.TrimSpace(strings.ToLower(ev.Text))

				user, err := api.GetUserInfo(ev.User)
				if err != nil {
					continue
				}

				outGoingMessage := ""

				if ev.User != info.User.ID && strings.HasPrefix(messageText, botID) && strings.Contains(messageText, "where") {
					bucket := time.Now().Format("02.01.2006")
					if strings.Contains(messageText, "tomorrow") {
						bucket = time.Now().AddDate(0, 0, 1).Format("02.01.2006")
					}

					outGoingMessage = getLocationsFromBucket(db, bucket)
				} else if strings.Contains(messageText, "tomorrow") {
					addStatusTomorrow(db, user.Profile.DisplayNameNormalized, strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.Replace(messageText, botID, "", 1)), "tomorrow", "")))

					outGoingMessage = "Thanks for letting us know <@" + ev.User + ">"
				} else if strings.Contains(messageText, "today") {
					addStatusToday(db, user.Profile.DisplayNameNormalized, strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.Replace(messageText, botID, "", 1)), "today", "")))

					outGoingMessage = "Thanks for letting us know <@" + ev.User + ">"
				} else {
					continue
				}

				if outGoingMessage != "" {
					rtm.SendMessage(rtm.NewOutgoingMessage(outGoingMessage, ev.Channel))
				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				//
			}
		}
	}
}

func getenv(name string) string {
	v := os.Getenv(name)

	if v == "" {
		panic("missing required env variable " + name)
	}

	return v
}

func addStatusTomorrow(db *bolt.DB, user string, location string) {
	bucket := time.Now().AddDate(0, 0, 1).Format("02.01.2006")
	addStatusToBucket(db, user, location, bucket)
}

func addStatusToday(db *bolt.DB, user string, location string) {
	bucket := time.Now().Format("02.01.2006")
	addStatusToBucket(db, user, location, bucket)
}

func addStatusToBucket(db *bolt.DB, user string, location string, bucket string) {
	checkCreateBucket(db, bucket)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(user), []byte(location))

		return err
	})
}

func getLocationsFromBucket(db *bolt.DB, bucket string) string {
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

func checkCreateBucket(db *bolt.DB, bucket string) {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket %s", err)
		}

		return nil
	})
}
