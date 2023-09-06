package database

import (
	"encoding/json"
	"github.com/jerbe/jim/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
	"testing"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/15 01:18
  @describe :
*/

func TestMain(t *testing.M) {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
		return
	}
	if _, err = Init(cfg); err != nil {
		panic(err)
		return
	}
	t.Run()
}

func TestMongodb(t *testing.T) {
	room := new(ChatRoom)
	err := GlobDB.Mongo.Database(DatabaseMongodbIM).Collection(CollectionRoom).FindOne(GlobCtx, bson.M{
		"_id": "00000001_00000002",
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"last_message": bson.M{
						"$eq": nil,
					}},
					bson.M{"last_message.message_id": bson.M{"$lt": 9999}},
				},
			},
		},
	}).Decode(room)
	if err != nil {
		log.Println(err)
		return
	}

	j, err := json.Marshal(room)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(j))
}

func BenchmarkNewObjectID(b *testing.B) {
	f, err := os.OpenFile("./objectid.txt", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	//b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := f.WriteString(primitive.NewObjectID().Hex() + "\n")
			if err != nil {
				log.Println(err)
			}
		}
	})
}
