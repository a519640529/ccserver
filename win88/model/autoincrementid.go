package model

import (
	"github.com/globalsign/mgo/bson"
)

func AutoIncGameLogId() (string, error) {
	return bson.NewObjectId().Hex(), nil
}
