package plugin

import (
	"testing"
	"github.com/zhsyourai/URCF-engine/models"
	"fmt"
)

var repo = NewPluginRepository()

func TestInsertAndFind(t *testing.T) {
	plug, err := repo.InsertPlugin(models.Plugin{
		Enabled: true,
		// Ignore other field, just for test
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindPluginByID(plug.ID)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	if account.ID != plug.ID {
		t.Errorf("%s(%s)", "Find error", "ID not equ")
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
}

func TestInsertAndUpdate(t *testing.T) {
	plug, err := repo.InsertPlugin(models.Plugin{
		Enabled: true,
		// Ignore other field, just for test
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindPluginByID(plug.ID)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.ID != plug.ID {
		t.Errorf("%s(%s)", "Find error", "ID not equ")
	}
	err = repo.UpdatePluginByID(plug.ID, map[string]interface{}{
		"Enabled": false,
	})
	if err != nil {
		t.Errorf("%s(%s)", "Update error", fmt.Sprint(err))
	}
	account, err = repo.FindPluginByID(plug.ID)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.ID != plug.ID {
		t.Errorf("%s(%s)", "Find error", "ID not equ")
	}
	if account.Enabled {
		t.Errorf("%s(%s)", "Find error", "Enabled should be false")
	}
	t.Logf("%s(%s)", "Update success", fmt.Sprint(account))
}

func TestInsertAndDelete(t *testing.T) {
	plug, err := repo.InsertPlugin(models.Plugin{
		Enabled: true,
		// Ignore other field, just for test
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindPluginByID(plug.ID)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	if account.ID != plug.ID {
		t.Errorf("%s(%s)", "Find error", "ID not equ")
	}
	account, err = repo.DeletePluginByID(plug.ID)
	if err != nil {
		t.Errorf("%s(%s)", "Delete error", fmt.Sprint(err))
	}
	if account.ID != plug.ID {
		t.Errorf("%s(%s)", "Delete error", "ID not equ")
	}
	account, err = repo.FindPluginByID(plug.ID)
	if err != nil && err.Error() != "leveldb: not found" {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s", "Delete success")
}