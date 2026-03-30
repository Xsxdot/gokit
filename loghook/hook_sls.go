package loghook

import (
	"fmt"
	"time"

	config "github.com/xsxdot/gokit/config"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
)

type SlsHook struct {
	levels   []logrus.Level
	client   sls.ClientInterface
	appName  string
	host     string
	project  string
	logstore string
}

func NewSlsHook(appName, host string, cfg config.LogConfig) *SlsHook {
	provider := sls.NewStaticCredentialsProvider(cfg.AccessKey, cfg.AccessSecret, "")
	client := sls.CreateNormalInterfaceV2(cfg.Endpoint, provider)

	return &SlsHook{
		levels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
		client:   client,
		appName:  appName,
		host:     host,
		project:  cfg.Project,
		logstore: cfg.Logstore,
	}
}

func (s *SlsHook) Fire(entry *logrus.Entry) error {
	var content []*sls.LogContent
	for k, v := range entry.Data {
		content = append(content, &sls.LogContent{
			Key:   proto.String(k),
			Value: proto.String(fmt.Sprintf("%v", v)),
		})
	}

	content = append(content, &sls.LogContent{
		Key:   proto.String("message"),
		Value: proto.String(entry.Message),
	})

	logGroup := &sls.LogGroup{
		Topic:  proto.String(s.appName),
		Source: proto.String(s.host),
		Logs: []*sls.Log{{
			Time:     proto.Uint32(uint32(time.Now().Unix())),
			Contents: content,
		}},
	}

	return s.client.PutLogs(s.project, s.logstore, logGroup)
}

func (s *SlsHook) Levels() []logrus.Level {
	return s.levels
}
