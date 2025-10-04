package mongocc

import "go.mongodb.org/mongo-driver/v2/bson"

type Options struct {
	Fields *[]string
	Limit  *int64
	Skip   *int64
	Sort   *[]bson.M
}
