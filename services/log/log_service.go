package log

import (
	"bufio"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/zhsyourai/URCF-engine/config"
	"github.com/zhsyourai/URCF-engine/models"
	"github.com/zhsyourai/URCF-engine/repositories"
	"github.com/zhsyourai/URCF-engine/repositories/log"
	"github.com/zhsyourai/URCF-engine/services"
	"io"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
)

type Level uint32

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

type Service interface {
	services.ServiceLifeCycle
	GetLogger(name string) (*logrus.Entry, error)
	WarpReader(name string, r io.Reader) error
	ListAll(page uint32, size uint32, sort string, order string) (int64, []models.Log, error)
	Clean(ids ...int64) error
	ChangeLevel(level Level) error
}

var instance *logService
var once sync.Once

func GetInstance() Service {
	once.Do(func() {
		instance = &logService{
			level: InfoLevel,
			repo:  log.NewLogRepository(),
		}
	})
	return instance
}

type logService struct {
	services.InitHelper
	level Level
	repo  log.Repository
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
	err = l.repo.InsertLog(&models.Log{
		Name:       entry["name"].(string),
		Message:    entry["msg"].(string),
		CreateTime: parseTime,
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

func (s *logService) ChangeLevel(level Level) error {
	s.level = level
	return nil
}

func (s *logService) GetLogger(name string) (*logrus.Entry, error) {
	logger := &logrus.Logger{
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.Level(s.level),
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

func (s *logService) ListAll(page uint32, size uint32, sort string, order string) (total int64, logs []models.Log,
	err error) {
	total, err = s.repo.CountAll()
	if err != nil {
		return 0, []models.Log{}, err
	}
	if sort == "" {
		logs, err = s.repo.FindAll(page, size, nil)
		if err != nil {
			return 0, []models.Log{}, err
		}
	} else {
		o, err := repositories.ParseOrder(order)
		if err != nil {
			return 0, []models.Log{}, err
		}
		logs, err = s.repo.FindAll(page, size, []repositories.Sort{
			{
				Name:  sort,
				Order: o,
			},
		})
		if err != nil {
			return 0, []models.Log{}, err
		}
	}

	return
}

func (s *logService) Clean(ids ...int64) error {
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
			if err != nil {
				if err == io.EOF || err == syscall.EIO || strings.Contains(err.Error(), "file already closed") {
					break
				} else {
					logrus.Error(err)
					break
				}
			}

			if line != "" {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				entry, kvPairs, err := parseJSON(line)
				if err != nil {
					logger.Info(line)
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
		}
	}()
	return nil
}
