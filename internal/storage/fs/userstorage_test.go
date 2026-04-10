package fs

import (
	"os"
	"path"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/model"
)

func newTestStorage(t *testing.T) (*FileSystemStorage, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{
		DataDir: dir,
	}
	return NewStorage(cfg), dir
}

func TestRegisterUser(t *testing.T) {
	fs, dir := newTestStorage(t)

	user, err := model.NewUser("testuser@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.RegisterUser(user)
	if err != nil {
		t.Fatal(err)
	}

	// Profile file should exist
	profilePath := path.Join(dir, userDir, user.ID, profileName)
	if _, err := os.Stat(profilePath); err != nil {
		t.Fatalf("profile file not created: %v", err)
	}

	// Sync directory should exist
	syncPath := path.Join(dir, userDir, user.ID, SyncFolder)
	if _, err := os.Stat(syncPath); err != nil {
		t.Fatalf("sync directory not created: %v", err)
	}
}

func TestRegisterUser_EmptyID(t *testing.T) {
	fs, _ := newTestStorage(t)

	user := &model.User{ID: ""}
	err := fs.RegisterUser(user)
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestRegisterUser_Duplicate(t *testing.T) {
	fs, _ := newTestStorage(t)

	user, err := model.NewUser("dup@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}

	if err := fs.RegisterUser(user); err != nil {
		t.Fatal(err)
	}

	// Second registration should fail (O_EXCL)
	err = fs.RegisterUser(user)
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestGetUser(t *testing.T) {
	fs, _ := newTestStorage(t)

	original, err := model.NewUser("getuser@example.com", "pass123")
	if err != nil {
		t.Fatal(err)
	}
	original.Name = "Test User"
	original.Nickname = "testy"

	if err := fs.RegisterUser(original); err != nil {
		t.Fatal(err)
	}

	loaded, err := fs.GetUser(original.ID)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.ID != original.ID {
		t.Errorf("ID mismatch: got %q, want %q", loaded.ID, original.ID)
	}
	if loaded.Email != original.Email {
		t.Errorf("Email mismatch: got %q, want %q", loaded.Email, original.Email)
	}
	if loaded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", loaded.Name, original.Name)
	}
}

func TestGetUser_Empty(t *testing.T) {
	fs, _ := newTestStorage(t)

	_, err := fs.GetUser("")
	if err == nil {
		t.Fatal("expected error for empty user")
	}
}

func TestGetUser_NotFound(t *testing.T) {
	fs, _ := newTestStorage(t)

	_, err := fs.GetUser("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

func TestGetUsers(t *testing.T) {
	fs, _ := newTestStorage(t)

	for _, email := range []string{"user1@test.com", "user2@test.com", "user3@test.com"} {
		u, err := model.NewUser(email, "pass")
		if err != nil {
			t.Fatal(err)
		}
		if err := fs.RegisterUser(u); err != nil {
			t.Fatal(err)
		}
	}

	users, err := fs.GetUsers()
	if err != nil {
		t.Fatal(err)
	}

	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	fs, _ := newTestStorage(t)

	user, err := model.NewUser("update@test.com", "pass")
	if err != nil {
		t.Fatal(err)
	}

	if err := fs.RegisterUser(user); err != nil {
		t.Fatal(err)
	}

	user.Name = "Updated Name"
	user.Nickname = "updated"

	if err := fs.UpdateUser(user); err != nil {
		t.Fatal(err)
	}

	loaded, err := fs.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != "Updated Name" {
		t.Errorf("Name not updated: got %q", loaded.Name)
	}
	if loaded.Nickname != "updated" {
		t.Errorf("Nickname not updated: got %q", loaded.Nickname)
	}
}

func TestUpdateUser_EmptyID(t *testing.T) {
	fs, _ := newTestStorage(t)

	err := fs.UpdateUser(&model.User{ID: ""})
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestRemoveUser(t *testing.T) {
	fs, dir := newTestStorage(t)

	user, err := model.NewUser("remove@test.com", "pass")
	if err != nil {
		t.Fatal(err)
	}

	if err := fs.RegisterUser(user); err != nil {
		t.Fatal(err)
	}

	userPath := path.Join(dir, userDir, user.ID)
	if _, err := os.Stat(userPath); err != nil {
		t.Fatal("user dir should exist before removal")
	}

	if err := fs.RemoveUser(user.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(userPath); !os.IsNotExist(err) {
		t.Fatal("user dir should be removed")
	}
}

func TestRemoveUser_EmptyID(t *testing.T) {
	fs, _ := newTestStorage(t)

	err := fs.RemoveUser("")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}
