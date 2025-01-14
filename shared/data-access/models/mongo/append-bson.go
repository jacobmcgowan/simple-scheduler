package mongoModels

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AppendBson[T any](doc primitive.D, field string, value *T) primitive.D {
	if value != nil {
		return append(doc, bson.E{
			Key:   field,
			Value: value,
		})
	}

	return doc
}

func AppendBsonCondition[T any](doc primitive.D, field string, condition string, value *T) primitive.D {
	if value != nil {
		return append(doc, bson.E{
			Key: field,
			Value: bson.D{{
				Key:   condition,
				Value: value,
			}},
		})
	}

	return doc
}
