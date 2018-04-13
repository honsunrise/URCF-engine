package account

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zhsyourai/URCF-engine/models"
)


var repo = NewAccountRepository()

func TestDoubleInsertAndFind(t *testing.T) {
	testUsername := "__test" + fmt.Sprint(rand.Int())
	err := repo.InsertAccount(models.Account{
		Username:      testUsername,
		Enabled: true,
		// Ignore other field, just for test
	})
	err = repo.InsertAccount(models.Account{
		Username:      testUsername,
		Enabled: true,
		// Ignore other field, just for test
	})
	if err == nil && err != ErrAccountExist {
		t.Fatal("Double Insert check error")
	}
	account, err := repo.FindAccountByUsername(testUsername)
	if err != nil {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	if account.Username != testUsername {
		t.Fatalf("%s(%s)", "Find error", "Username not equ")
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
}

func TestInsertAndUpdate(t *testing.T) {
	testUsername := "__test" + fmt.Sprint(rand.Int())
	err := repo.InsertAccount(models.Account{
		Username:      testUsername,
		Enabled: true,
		// Ignore other field, just for test
	})
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindAccountByUsername(testUsername)
	if err != nil {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.Username != testUsername {
		t.Fatalf("%s(%s)", "Find error", "Username not equ")
	}
	_, err = repo.UpdateAccountByUsername(testUsername, map[string]interface{}{
		"Enabled": false,
	})
	if err != nil {
		t.Fatalf("%s(%s)", "Update error", fmt.Sprint(err))
	}
	account, err = repo.FindAccountByUsername(testUsername)
	if err != nil {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.Username != testUsername {
		t.Fatalf("%s(%s)", "Find error", "Username not equ")
	}
	if account.Enabled {
		t.Fatalf("%s(%s)", "Find error", "Enabled should be false")
	}
	t.Logf("%s(%s)", "Update success", fmt.Sprint(account))
}

func TestInsertAndDelete(t *testing.T) {
	testUsername := "__test" + fmt.Sprint(rand.Int())
	err := repo.InsertAccount(models.Account{
		Username:      testUsername,
		Enabled: true,
		// Ignore other field, just for test
	})
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindAccountByUsername(testUsername)
	if err != nil {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.Username != testUsername {
		t.Fatalf("%s(%s)", "Find error", "Username not equ")
	}
	account, err = repo.DeleteAccountByUsername(testUsername)
	if err != nil {
		t.Fatalf("%s(%s)", "Delete error", fmt.Sprint(err))
	}
	if account.Username != testUsername {
		t.Fatalf("%s(%s)", "Delete error", "Username not equ")
	}
	account, err = repo.FindAccountByUsername(testUsername)
	if err != nil && err == leveldb.ErrNotFound {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s", "Delete success")
}
