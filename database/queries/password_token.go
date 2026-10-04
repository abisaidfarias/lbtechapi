package queries

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func InvalidatePasswordTokensFilter(userID primitive.ObjectID, purpose string) bson.M {
	return bson.M{
		"user_id":   userID,
		"purpose":   purpose,
		"used_at":   bson.M{"$exists": false},
		"expires_at": bson.M{"$gt": time.Now()},
	}
}

func InvalidatePasswordTokensUpdate(now time.Time) bson.M {
	return bson.M{
		"$set": bson.M{"used_at": now},
	}
}

func FindValidPasswordTokenFilter(tokenHash string) bson.M {
	return bson.M{
		"token_hash": tokenHash,
		"used_at":    bson.M{"$exists": false},
		"expires_at": bson.M{"$gt": time.Now()},
	}
}

func MarkPasswordTokenUsedFilter(id primitive.ObjectID) bson.M {
	return bson.M{"_id": id}
}

func MarkPasswordTokenUsedUpdate(now time.Time) bson.M {
	return bson.M{
		"$set": bson.M{"used_at": now},
	}
}

func ActivateUserPasswordFilter(userID primitive.ObjectID) bson.M {
	return bson.M{"_id": userID}
}

func ActivateUserPasswordUpdate(hashPassword string, activatedAt time.Time) bson.M {
	return bson.M{
		"$set": bson.M{
			"passwordHash":    hashPassword,
			"account_status":  "active",
			"password_set_at": activatedAt,
		},
	}
}
