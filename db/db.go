package db

import (
	"context"
	"log"

	"github.com/WhiCu/mongoRedisFiber/db/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DB struct {
	DB          *mongo.Database
	Collections map[string]*mongo.Collection
}

func Client(ctx context.Context, uri string) *mongo.Client {

	res := make(chan *mongo.Client)

	go func() {
		clientOptions := options.Client().ApplyURI(uri)

		client, err := mongo.Connect(clientOptions)

		if err != nil {
			//TODO: add logging
			log.Fatal("Connect error:", err)
		}

		err = client.Ping(context.TODO(), nil)
		if err != nil {
			log.Fatal(err)
		}
		res <- client
	}()

	select {
	case client := <-res:
		return client
	case <-ctx.Done():
		return nil
	}

}

func NewDB(client *mongo.Client, dbName string) *DB {
	return &DB{
		DB:          client.Database(dbName),
		Collections: make(map[string]*mongo.Collection),
	}
}

func (db *DB) Collection(name string) *mongo.Collection {
	//TODO: if value, ok := db.Collections[name]; !ok -?
	if db.Collections[name] == nil {
		//TODO: add logging
		log.Println("New collection")
		db.Collections[name] = db.DB.Collection(name)
	}
	return db.Collections[name]
}

func (db *DB) FindToken(ctx context.Context, collectionName string, token string) *types.User {

	var user types.User

	filter := bson.D{{Key: "token", Value: token}}

	err := db.Collection(collectionName).FindOne(ctx, filter).Decode(&user)

	if err != nil {
		log.Println(err)
		return nil
	}

	return &user
}

func (db *DB) AddUser(ctx context.Context, collectionName string, user *types.User) (id string, token string) {

	log.Println(user)

	result, err := db.Collection(collectionName).InsertOne(ctx, user)
	if err != nil {
		return "", ""
	}

	log.Println(result.InsertedID)

	return result.InsertedID.(bson.ObjectID).Hex(), user.GetToken()
}
