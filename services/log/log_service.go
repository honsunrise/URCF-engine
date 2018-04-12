package log

import (
	"github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/services"
	"sync"
	"github.com/zhsyourai/URCF-engine/config"
	"github.com/zhsyourai/URCF-engine/repositories/log"
	"github.com/zhsyourai/URCF-engine/models"
	"time"
	"os"
	"encoding/json"
	"io"
	"bufio"
	"strings"
	"unicode"
)

type Service interface {
	services.ServiceLifeCycle
	GetLogger(name string) (*logrus.Entry, error)
	WarpReader(name string, r io.Reader) error
	ListAll(page uint64, size uint64, sort string, order string) (uint64, []models.Log, error)
	Clean(ids ...uint64) error
}

var instance *logService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &logService{
			repo: log.NewLogRepository(),
		}
	})
	return instance
}

type logService struct {
	services.InitHelper
	repo log.Repository
}

type logWriter struct{ *logService }

type logEntry struct {
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
}

func (l *logWriter) Write(p []byte) (int, error) {
	var entry map[string]interface{}
	err := json.Unmarshal(p, &entry)
	if err != nil {
		return 0, err
	}
	level, err := models.ParseLevel(entry["level"].(string))
	if err != nil {
		return 0, err
	}
	parseTime, err := time.Parse(time.RFC3339, entry["time"].(string))
	if err != nil {
		return 0, nil
	}
	_, err = l.repo.InsertLog(models.Log{
		Name:       entry["name"].(string),
		Message:    entry["msg"].(string),
		CreateDate: parseTime,
		Level:      level,
	})
	if err != nil {
		return 0, nil
	}
	return len(p), nil
}

func (s *logService) Initialize(arguments ...interface{}) error {
	return s.CallInitialize(func() error {
		return nil
	})
}

func (s *logService) UnInitialize(arguments ...interface{}) error {
	return s.CallUnInitialize(func() error {
		return nil
	})
}

func (s *logService) GetLogger(name string) (*logrus.Entry, error) {
	logger := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
	}
	if !config.PROD {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	logger.Out = &logWriter{
		s,
	}
	return logger.WithField("name", name), nil
}

func (s *logService) ListAll(page uint64, size uint64, sort string, order string) (total uint64, logs []models.Log,
	err error) {
	total, err = s.repo.Count()
	if err != nil {
		return 0, []models.Log{}, nil
	}
	logs, err = s.repo.FindAll()
	if err != nil {
		return 0, []models.Log{}, nil
	}
	return
}

func (s *logService) Clean(ids ...uint64) error {
	if len(ids) == 0 {
		return s.repo.DeleteAll()
	} else {
		for _, id := range ids {
			_, err := s.repo.DeleteLogByID(id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseJSON(input string) (*logEntry, map[string]interface{}, error) {
	var raw map[string]interface{}
	entry := &logEntry{}

	err := json.Unmarshal([]byte(input), &raw)
	if err != nil {
		return nil, nil, err
	}

	if v, ok := raw["message"]; ok {
		entry.Message = v.(string)
		delete(raw, "message")
	}

	if v, ok := raw["level"]; ok {
		entry.Level = v.(string)
		delete(raw, "level")
	}

	if v, ok := raw["timestamp"]; ok {
		t, err := time.Parse("2006-01-02T15:04:05.000000Z07:00", v.(string))
		if err != nil {
			return nil, nil, err
		}
		entry.Timestamp = t
		delete(raw, "timestamp")
	}

	return entry, raw, nil
}

func (s *logService) WarpReader(name string, r io.Reader) error {
	logger, err := s.GetLogger(name)
	if err != nil {
		return err
	}
	go func() {
		bufR := bufio.NewReader(r)
		for {
			line, err := bufR.ReadString('\n')
			if line != "" {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				entry, kvPairs, err := parseJSON(line)
				if err != nil {
					logger.Debug(line)
				} else {
					for k, v := range kvPairs {
						logger = logger.WithField(k, v)
					}
					level, err := logrus.ParseLevel(entry.Level)
					if err != nil {
						logger.Debug(line)
					} else {
						switch level {
						case logrus.DebugLevel:
							logger.Debug(entry.Message)
						case logrus.InfoLevel:
							logger.Info(entry.Message)
						case logrus.WarnLevel:
							logger.Warn(entry.Message)
						case logrus.ErrorLevel:
							logger.Error(entry.Message)
						case logrus.FatalLevel:
							logger.Fatal(entry.Message)
						}
					}
				}
			}

			if err == io.EOF {
				break
			}
		}
	}()
	return nil
}
