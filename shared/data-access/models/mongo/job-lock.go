package mongoModels

import (
	"time"

	repositoryErrors "github.com/jacobmcgowan/simple-scheduler/shared/data-access/repositories/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func JobLock(managerId string, heartbeat time.Time) (bson.D, error) {
	objId, err := bson.ObjectIDFromHex(managerId)
	if err != nil {
		return nil, &repositoryErrors.InvalidIdError{
			Value: managerId,
		}
	}

	setDoc := bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "managerId",
			Value: objId,
		}, {
			Key:   "heartbeat",
			Value: heartbeat,
		}},
	}}

	return setDoc, nil
}

func JobUnlock() bson.D {
	return bson.D{{
		Key:   "$unset",
		Value: "managerId",
	}}
}
