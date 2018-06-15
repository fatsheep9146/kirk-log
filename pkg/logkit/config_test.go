package logkit

import (
	"encoding/json"
	"testing"

	"github.com/fatsheep9146/kirklog/pkg/api"
)

func TestLoadConfig(t *testing.T) {
	testLogSource := &api.LogSource{
		Meta: api.Meta{
			Name: "test_applog_test-xxx-yyy",
		},
		Spec: api.LogSourceSpec{
			PodName:     "test-xxx-yyy",
			Namespace:   "test-ns",
			VolumeMount: "applog",
			Config: `
{
	"name": "applog",
	"batch_len": 2000,
	"batch_size": 1572864,
	"reader": {
		"mode": "tailx",
		"log_path": "/applog/*/*.log*",
		"read_from": "newest",
		"datasource_tag": "log_source",
		"expire": "240h",
		"stat_interval": "1m",
		"ignore_file_suffix":".pid, .swp, .xml, .txt"
	},
	"parser": {
		"name": "applog_parser",
		"type": "qiniulog"
	},
	"senders": [{
		"name": "applog_sendor",
		"sender_type": "pandora",
		"fault_tolerant":"true",
		"pandora_ak": "CKu2k1bdor3xbcHv3AOfShTOTy2n1rXFmt4Ukmdt",
			"pandora_sk": "tXwQPcFRg1zAspWuj9iNmn_qBBrRXALFHzKippri",
		"pandora_host": "https://pipeline.qiniu.com",
		"pandora_repo_name": "dora_boots_gate_applog",
		"pandora_region": "nb",
		"pandora_schema": "reqid,time logtime,level,file,log,pod_name machine",
		"ft_sync_every":"10000",
		"ft_strategy": "backup_only"
	}]
}`,
		},
	}

	configRaw, err := renderConfig(testLogSource)
	if err != nil {
		t.Errorf(err.Error())
	}
	var config LogkitConf

	err = json.Unmarshal([]byte(configRaw), &config)
	if err != nil {
		t.Errorf(err.Error())
	}

	if config.ReaderConfig["mode"] != "dir" {
		t.Errorf("config reader config mode is not dir, is %v", config.ReaderConfig["mode"])
	}

	if config.ReaderConfig["log_path"] != "/applog/test-ns_test-xxx-yyy" {
		t.Errorf("config reader config log_path is wrong, is %v", config.ReaderConfig["log_path"])
	}

	if config.ReaderConfig["meta_path"] != "/applog/test-ns_test-xxx-yyy/.meta" {
		t.Errorf("config reader config meta_path is wrong, is %v", config.ReaderConfig["meta_path"])
	}

	if len(config.Transforms) == 0 {
		t.Errorf("config transform config should exist")
	}

	t.Logf("%v", configRaw)
}
