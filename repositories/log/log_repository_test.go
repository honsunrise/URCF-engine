package log

import (
	"fmt"
	"testing"

	"github.com/zhsyourai/URCF-engine/models"
	"time"
)

var repo = NewLogRepository()

func TestInsertAndFind(t *testing.T) {
	id, err := repo.InsertLog(models.Log{
		Message: "I am a log's message.",
		Level: models.DebugLevel,
		CreateDate: time.Now(),
		Name: "log_test",
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindLogByID(id)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
}

func TestInsertAndDelete(t *testing.T) {
	id, err := repo.InsertLog(models.Log{
		Message: "I am a log's message.",
		Level: models.DebugLevel,
		CreateDate: time.Now(),
		Name: "log_test",
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err := repo.FindLogByID(id)
	if err != nil {
		t.Errorf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(account))
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	account, err = repo.DeleteLogByID(id)
	if err != nil {
		t.Errorf("%s(%s)", "Delete error", fmt.Sprint(err))
	}

}

func TestDeleteAll(t *testing.T) {
	_, err := repo.InsertLog(models.Log{
		Message: "I am a log's message.",
		Level: models.DebugLevel,
		CreateDate: time.Now(),
		Name: "log_test",
	})
	if err != nil {
		t.Errorf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	err = repo.DeleteAll()
	if err != nil {
		t.Errorf("%s(%s)", "Delete All error", fmt.Sprint(err))
	}
	t.Logf("%s", "Delete All success")
}
