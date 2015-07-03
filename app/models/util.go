package models

import (
	"labix.org/v2/mgo/bson"
)

func ArrayContainObjectId(arr []bson.ObjectId, id bson.ObjectId) bool {
	for _, o := range arr {
		if o == id {
			return true
		}
	}
	return false
}
