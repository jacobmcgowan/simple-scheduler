package common

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AppendBson[T any](document primitive.D, field string, value Undefinable[T]) primitive.D {
	if value.Defined {
		return append(document, bson.E{
			Key:   field,
			Value: value.Value,
		})
	}

	return document
}
