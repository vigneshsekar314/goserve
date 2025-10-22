package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	some_uuid := uuid.New()
	signedToken, err := MakeJWT(some_uuid, "my_secret", time.Second*60)
	if err != nil {
		t.Errorf("got error generating signed token in MakeJWT(), %s\n", err)
	}
	returned_uuid, err := ValidateJWT(signedToken, "my_secret")
	if err != nil {
		t.Errorf("got error retrieving uuid in ValidateJWT(), %s\n", err)
	}
	if some_uuid != returned_uuid {
		t.Errorf("original uuid is not present in parsed JWT. Original UUID: %s and Parsed UUID: %s\n", some_uuid, returned_uuid)
	}
}
