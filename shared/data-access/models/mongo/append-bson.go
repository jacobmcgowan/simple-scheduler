package mongoModels

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

func AppendBson[T any](doc bson.D, field string, value *T) bson.D {
	if value != nil {
		return append(doc, bson.E{
			Key:   field,
			Value: value,
		})
	}

	return doc
}

func AppendBsonCondition[T any](doc bson.D, field string, condition string, value *T) bson.D {
	if value != nil {
		return append(doc, bson.E{
			Key: field,
			Value: bson.M{
				condition: value,
			},
		})
	}

	return doc
}
