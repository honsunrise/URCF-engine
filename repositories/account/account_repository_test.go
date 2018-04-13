package account

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zhsyourai/URCF-engine/models"
)

var testID = "__test" + fmt.Sprint(rand.Int())
var repo = NewAccountRepository()

func TestInsertAndFind(t *testing.T) {
	err := repo.InsertAccount(models.Account{
		ID:      testID,
		Enabled: true,
		// Ignore other field, just for test
	})
	err = repo.InsertAccount(models.Account{
		ID:      testID,
		Enabled: true,
		// Ignore other field, just for test
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindAccountByID(testID)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	if account.ID != testID {
		t.Errorf("%s(%s)", "Find error", "ID not equ")
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
}

func TestInsertAndUpdate(t *testing.T) {
	err := repo.InsertAccount(models.Account{
		ID:      testID,
		Enabled: true,
		// Ignore other field, just for test
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindAccountByID(testID)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.ID != testID {
		t.Errorf("%s(%s)", "Find error", "ID not equ")
	}
	err = repo.UpdateAccountByID(testID, map[string]interface{}{
		"Enabled": false,
	})
	if err != nil {
		t.Errorf("%s(%s)", "Update error", fmt.Sprint(err))
	}
	account, err = repo.FindAccountByID(testID)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.ID != testID {
		t.Errorf("%s(%s)", "Find error", "ID not equ")
	}
	if account.Enabled {
		t.Errorf("%s(%s)", "Find error", "Enabled should be false")
	}
	t.Logf("%s(%s)", "Update success", fmt.Sprint(account))
}

func TestInsertAndDelete(t *testing.T) {
	err := repo.InsertAccount(models.Account{
		ID:      testID,
		Enabled: true,
		// Ignore other field, just for test
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindAccountByID(testID)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.ID != testID {
		t.Errorf("%s(%s)", "Find error", "ID not equ")
	}
	account, err = repo.DeleteAccountByID(testID)
	if err != nil {
		t.Errorf("%s(%s)", "Delete error", fmt.Sprint(err))
	}
	if account.ID != testID {
		t.Errorf("%s(%s)", "Delete error", "ID not equ")
	}
	account, err = repo.FindAccountByID(testID)
	if err != nil && err == leveldb.ErrNotFound {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s", "Delete success")
}
