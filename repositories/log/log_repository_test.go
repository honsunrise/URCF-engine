package log

import (
	"fmt"
	"testing"

	"github.com/zhsyourai/URCF-engine/models"
	"time"
)

var repo = NewLogRepository()

func TestInsertAndFind(t *testing.T) {
	log := &models.Log{
		Message:    "I am a log's message in TestInsertAndFind.",
		Level:      models.DebugLevel,
		CreateTime: time.Now(),
		Name:       "log_test",
	}
	err := repo.InsertLog(log)
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	logRet, err := repo.FindLogByID(log.ID)
	if err != nil {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(logRet))
}

func TestInsertAndDelete(t *testing.T) {
	log := &models.Log{
		Message:    "I am a log's message in TestInsertAndDelete.",
		Level:      models.DebugLevel,
		CreateTime: time.Now(),
		Name:       "log_test",
	}
	err := repo.InsertLog(log)
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	logRet, err := repo.FindLogByID(log.ID)
	if err != nil {
		t.Fatalf("%s(%s)", "Find error", fmt.Sprint(err))
	}
	t.Logf("%s(%s)", "Find success", fmt.Sprint(logRet))
	_, err = repo.DeleteLogByID(log.ID)
	if err != nil {
		t.Fatalf("%s(%s)", "Delete error", fmt.Sprint(err))
	}

}

func TestFindAll(t *testing.T) {
	logs, err := repo.FindAll(0, 100, nil)
	if err != nil {
		t.Fatalf("%s(%s)", "Find all error", fmt.Sprint(err))
	}
	if len(logs) != 1 {
		t.Fatalf("Find All error (len %d not equal 1)", len(logs))
	}
	t.Logf("%s", "Find All success")
}

func TestDeleteAll(t *testing.T) {
	log := &models.Log{
		Message:    "I am a log's message in TestDeleteAll.",
		Level:      models.DebugLevel,
		CreateTime: time.Now(),
		Name:       "log_test",
	}
	err := repo.InsertLog(log)
	if err != nil {
		t.Fatalf("%s(%s)", "Insert error", fmt.Sprint(err))
	}
	err = repo.DeleteAll()
	if err != nil {
		t.Fatalf("%s(%s)", "Delete All error", fmt.Sprint(err))
	}

	logs, err := repo.FindAll(0, 100, nil)
	if err != nil {
		t.Fatalf("%s(%s)", "Find all error", fmt.Sprint(err))
	}
	if len(logs) != 0 {
		t.Fatalf("Delete All error (len %d not equal 0)", len(logs))
	}
	t.Logf("%s", "Delete All success")
}
