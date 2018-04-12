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
		Message: "I am a log's message in TestInsertAndFind.",
		Level: models.DebugLevel,
		CreateDate: time.Now(),
		Name: "log_test",
	})
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	log, err := repo.FindLogByID(id)
	if err != nil {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(log))
}

func TestInsertAndDelete(t *testing.T) {
	id, err := repo.InsertLog(models.Log{
		Message: "I am a log's message in TestInsertAndDelete.",
		Level: models.DebugLevel,
		CreateDate: time.Now(),
		Name: "log_test",
	})
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	log, err := repo.FindLogByID(id)
	if err != nil {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(log))
	_, err = repo.DeleteLogByID(id)
	if err != nil {
		t.Fatalf("%s(%s)", "Delete error", fmt.Sprint(err))
	}

}

func TestFindAll(t *testing.T) {
	logs, err := repo.FindAll()
	if err != nil {
		t.Fatalf("%s(%s)", "Find all error", fmt.Sprint(err))
	}
	if len(logs) != 1 {
		t.Fatalf("Find All error (len %d not equal 1)", len(logs))
	}
	t.Logf("%s", "Find All success")
}

func TestDeleteAll(t *testing.T) {
	_, err := repo.InsertLog(models.Log{
		Message: "I am a log's message in TestDeleteAll.",
		Level: models.DebugLevel,
		CreateDate: time.Now(),
		Name: "log_test",
	})
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	err = repo.DeleteAll()
	if err != nil {
		t.Fatalf("%s(%s)", "Delete All error", fmt.Sprint(err))
	}

	logs, err := repo.FindAll()
	if err != nil {
		t.Fatalf("%s(%s)", "Find all error", fmt.Sprint(err))
	}
	if len(logs) != 0 {
		t.Fatalf("Delete All error (len %d not equal 0)", len(logs))
	}
	t.Logf("%s", "Delete All success")
}
