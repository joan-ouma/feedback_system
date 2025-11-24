package utils

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StringToObjectID converts a string to ObjectID
func StringToObjectID(idStr string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(idStr)
}

// ObjectIDToString converts ObjectID to string
func ObjectIDToString(id primitive.ObjectID) string {
	return id.Hex()
}

// MustObjectID converts string to ObjectID, panics on error (use carefully)
func MustObjectID(idStr string) primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		panic(fmt.Sprintf("invalid ObjectID: %s", idStr))
	}
	return id
}

