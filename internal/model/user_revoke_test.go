package model

import (
	"testing"
	"time"
)

func TestDeviceTokenNotRevokedByDefault(t *testing.T) {
	u := &User{}
	if u.IsDeviceTokenRevoked("dev-1", time.Now().Unix()) {
		t.Fatal("token should not be revoked when no revocation exists")
	}
}

func TestRevokeDeviceRejectsOlderTokens(t *testing.T) {
	u := &User{}
	issued := time.Now().Add(-time.Hour)
	u.RevokeDevice("dev-1", time.Now())

	if !u.IsDeviceTokenRevoked("dev-1", issued.Unix()) {
		t.Fatal("token issued before revocation should be revoked")
	}
	if !u.IsDeviceTokenRevoked("dev-1", 0) {
		t.Fatal("legacy token without iat should be revoked")
	}
	if u.IsDeviceTokenRevoked("dev-2", issued.Unix()) {
		t.Fatal("revocation must only affect the revoked device")
	}
}

func TestRepairedDeviceIsAccepted(t *testing.T) {
	u := &User{}
	revokedAt := time.Now().Add(-time.Minute)
	u.RevokeDevice("dev-1", revokedAt)

	if u.IsDeviceTokenRevoked("dev-1", time.Now().Unix()) {
		t.Fatal("token issued after revocation (re-pairing) should be accepted")
	}
}

func TestRevokedDevicesSurviveSerialization(t *testing.T) {
	u := &User{ID: "user1"}
	u.RevokeDevice("dev-1", time.Now())

	b, err := u.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DeserializeUser(b)
	if err != nil {
		t.Fatal(err)
	}
	if !restored.IsDeviceTokenRevoked("dev-1", 0) {
		t.Fatal("revocation lost after serialize/deserialize")
	}
}
