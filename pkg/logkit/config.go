package logkit

import (
	"encoding/json"
	"fmt"

	"github.com/fatsheep9146/kirklog/pkg/api"
)

type MetricConfig struct {
	MetricType string                 `json:"type"`
	Attributes map[string]bool        `json:"attributes"`
	Config     map[string]interface{} `json:"config"`
}

type RouterConfig struct {
	KeyName      string         `json:"router_key_name"`
	MatchType    string         `json:"router_match_type"`
	DefaultIndex int            `json:"router_default_sender"`
	Routes       map[string]int `json:"router_routes"`
}

type RunnerInfo struct {
	RunnerName       string `json:"name"`
	CollectInterval  int    `json:"collect_interval,omitempty"` // metric runner收集的频率
	MaxBatchLen      int    `json:"batch_len,omitempty"`        // 每个read batch的行数
	MaxBatchSize     int    `json:"batch_size,omitempty"`       // 每个read batch的字节数
	MaxBatchInterval int    `json:"batch_interval,omitempty"`   // 最大发送时间间隔
	MaxBatchTryTimes int    `json:"batch_try_times,omitempty"`  // 最大发送次数，小于等于0代表无限重试
	CreateTime       string `json:"createtime"`
	EnvTag           string `json:"env_tag,omitempty"`
	ExtraInfo        bool   `json:"extra_info,omitempty"`
}

type LogkitConf struct {
	RunnerInfo
	MetricConfig  []MetricConfig           `json:"metric,omitempty"`
	ReaderConfig  map[string]string        `json:"reader"`
	CleanerConfig map[string]string        `json:"cleaner,omitempty"`
	ParserConf    map[string]string        `json:"parser"`
	Transforms    []map[string]interface{} `json:"transforms,omitempty"`
	SenderConfig  []map[string]string      `json:"senders"`
	Router        RouterConfig             `json:"router,omitempty"`
	IsInWebFolder bool                     `json:"web_folder,omitempty"`
	IsStopped     bool                     `json:"is_stopped,omitempty"`
}

func renderConfig(logSource *api.LogSource) (string, error) {
	configRaw := logSource.Spec.Config
	config := LogkitConf{}

	err := json.Unmarshal([]byte(configRaw), &config)
	if err != nil {
		return "", err
	}

	config.RunnerInfo.RunnerName = getRunnerName(logSource)

	// Fix the readerconfig into mode dir and the right dir for log_path and meta_path
	config.ReaderConfig["mode"] = "dir"
	config.ReaderConfig["log_path"] = logSource.GetLogDir()
	config.ReaderConfig["meta_path"] = logSource.GetLogMetaDir()
	config.ReaderConfig["datasource_tag"] = "log_source"

	// Add the k8stag transformer if neccessary
	k8sTagAdd := true

	if len(config.Transforms) != 0 {
		for _, transform := range config.Transforms {
			if transform["type"] == "k8sdir" {
				k8sTagAdd = false
				break
			}
		}
	}

	if k8sTagAdd {
		k8sdirTransform := make(map[string]interface{})
		k8sdirTransform["type"] = "k8sdir"
		k8sdirTransform["sourcefilefield"] = "log_source"
		config.Transforms = append(config.Transforms, k8sdirTransform)
	}

	newConfigRaw, err := json.Marshal(config)
	return string(newConfigRaw), err
}

func getRunnerName(logSource *api.LogSource) string {
	return fmt.Sprintf("%s_%s", logSource.Spec.VolumeMount, logSource.Spec.PodName)
}

func getConfigFileName(logSource *api.LogSource) string {
	return fmt.Sprintf("%s.json", getRunnerName(logSource))
}
