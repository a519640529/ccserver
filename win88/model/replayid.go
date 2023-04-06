package model

import (
	"github.com/globalsign/mgo/bson"
)

func GetOneReplayId() (pid string, err error) {
	return bson.NewObjectId().Hex(), nil
}
