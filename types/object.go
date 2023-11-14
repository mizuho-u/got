package types

import "fmt"

type ObjectID string

func NewObjectID(s string) (ObjectID, error) {

	if len(s) != 40 {
		return "", fmt.Errorf("objectId should be 40 digits. got %d", len(s))
	}

	return ObjectID(s), nil
}

func (o ObjectID) Long() string {
	return o.String()
}

func (o ObjectID) Short() string {
	return o.String()[0:8]
}

func (o ObjectID) String() string {
	return string(o)
}
