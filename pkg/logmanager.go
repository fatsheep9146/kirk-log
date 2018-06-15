package logmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/fatsheep9146/kirklog/pkg/agent"
	"github.com/fatsheep9146/kirklog/pkg/api"
	"github.com/fatsheep9146/kirklog/pkg/logkit"
)

type LogManagerConfig struct {
	LogConfigDir string `json:"log_config_dir"`
	Name         string `json:"name"`
	Namespace    string
	AgentType    string `json:"agent_type"`
	Cli          *kubernetes.Clientset
}

type LogManager struct {
	// map for all log config
	LogConfigs map[string]api.LogConfig

	// map for all log entries
	LogSources map[string]api.LogSource

	// map for all log agent
	LogAgents map[string]agent.Agent

	// map for the match relation between logSource and logAgent
	Match map[string]*Match

	// the working queue to store the logSource wait to be processed
	Queue string

	// the Agent used to manage the log agent components
	LogAgentManager agent.AgentManager

	// The kubernetes client used to query info from k8s
	Cli *kubernetes.Clientset
}

// This type is used to indicate the match relation between logSource and logAgent
type Match struct {
	PodName   string
	AgentName string
	ConfPath  string
}

func NewLogManagerConfig() *LogManagerConfig {
	return &LogManagerConfig{}
}

func NewLogManager(cfg *LogManagerConfig) *LogManager {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create logConfigs from files
	logConfigs, err := loadLogConfig(cfg.LogConfigDir)
	if err != nil {
		// error: load log config error
		panic(err)
	}
	logConfigsMap := logConfigConvertFromSliceToMap(logConfigs)

	// Create logsources spec map from logConfigs
	listLogSourcesFunc := getListLogSourcesFunc(cli, logConfigsMap)
	logSources, _ := listLogSourcesFunc()

	// Check and create LogAgentManager and the deployment of log collector if not exist
	logAgentManager := newAgentManager(agent.AgentType(cfg.AgentType), &agent.AgentManagerConfig{
		Name:       cfg.Name,
		Namespace:  cfg.Namespace,
		LogConfigs: logConfigs,
		Cli:        cli,
	})
	logAgents, _ := logAgentManager.List()
	if len(logAgents) == 0 {
		logAgentManager.Deploy()
		logAgents, _ = logAgentManager.List()
	}

	// Restore the logsources map status from current situations in case this is a restart

	return &LogManager{
		LogConfigs:      logConfigsMap,
		LogSources:      logSourceConvertFromSliceToMap(logSources),
		LogAgents:       logAgentConvertFromSliceToMap(logAgents),
		LogAgentManager: logAgentManager,
		Queue:           "",
		Cli:             cli,
	}
}

func (lm *LogManager) Run() {
	// This function choose whether to rearrange the match relations between logSource and logAgent
	lm.syncInfo()

	// Start worker to handle the message in queue

	// Start the goroutine to check the lag of each logSource

}

// Create an logAgentManager according to the type of agent.
func newAgentManager(agentType agent.AgentType, cfg *agent.AgentManagerConfig) agent.AgentManager {
	switch agentType {
	case agent.Logkit:
		return logkit.NewLogkitAgentManager(cfg)
	}
	return nil
}

// Loop function to sync the info about logSource and logAgent
// If logSource or logAgent changes, use scheduling algorithm to
func (lm *LogManager) syncInfo() {
	// Get current logSources
	listLogSourcesFunc := getListLogSourcesFunc(lm.Cli, lm.LogConfigs)
	// Get current logAgents
	listLogAgentsFunc := lm.LogAgentManager.List

	for {
		// Get current logSources
		logSources, _ := listLogSourcesFunc()
		// Get current logAgents
		logAgents, _ := listLogAgentsFunc()

		// Update the logSources and logAgents map
		updateLogSources(lm.LogSources, logSources, lm.Match)
		updateLogAgents(lm.LogAgents, logAgents, lm.Match)

		// Update the match relation between logSource and logAgent
		_ = schedule(lm.LogSources, lm.LogAgents, lm.Match)

		// Enqueue the LogSources that are needed to be synced
	}
}

func (lm *LogManager) sync() error {
	// Do different works according to different sitiations
	// Get key from queue
	key := ""
	var flag bool
	var err error
	// Get logSource entry of this key from map

	action := judgeAction(lm.Match[key])
	switch action {
	case LogSourceAdd:
		flag, err = lm.logSourceAddFunc(key)
	case LogSourceDel:
		flag, err = lm.logSourceDelFunc(key)
	case LogSourceMov:
		flag, err = lm.logSourceMovFunc(key)
	}

	if flag {
		// key done handling, forget the key
		return nil
	} else {
		if err != nil {
			return err
		} else {
			// key undone reenqueue
			return nil
		}
	}
	return nil
}

// Remove the logSource from logSourcesMap and logSource log dir and Match
func (lm *LogManager) removeLogSource(*api.LogSource) error {

	fmt.Printf("Not Implemented yet")

	return nil
}

// Create logconfig objects from the files under the path dir
func loadLogConfig(path string) ([]api.LogConfig, error) {
	logConfigs := make([]api.LogConfig, 0)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return logConfigs, err
	}

	for _, file := range files {
		raw, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, file.Name()))
		if err != nil {
			// warning the file is not read successfully
			continue
		}
		logConfig := &api.LogConfig{}
		err = json.Unmarshal(raw, logConfig)
		if err != nil {
			// warning the file is not legal json
			continue
		}
		logConfigs = append(logConfigs, *logConfig)
	}

	// debug: show all log config

	return logConfigs, nil
}

// Update the map of logSources and match according the newest logsource list
// If there is a new logSource, then add it to logSourcesMap, and create a new match with podname, no conf, no agentname
// If there is a deleted logSource, then modify the match of this logSource, remove the PodName
func updateLogSources(logSourcesMap map[string]api.LogSource, logSources []api.LogSource, match map[string]*Match) {
	visited := make(map[string]bool)

	for _, logSource := range logSources {
		if _, exist := logSourcesMap[logSource.Meta.Name]; !exist {
			// If this is a new added logSource
			logSourcesMap[logSource.Meta.Name] = logSource
			match[logSource.Meta.Name] = &Match{
				PodName: logSource.Spec.PodName,
			}
		} else {
			visited[logSource.Meta.Name] = true
		}
	}

	for logSourceName, _ := range logSourcesMap {
		// If there is a deleted logSource, the modify the match of this logSource, remove the PodName
		if _, exist := visited[logSourceName]; !exist {
			match[logSourceName].PodName = ""
		}
	}
}

// Update the map of logAgents and match according the newest logsource list
// If there is a deleted logAgent, then modify the match which has this logAgent, remove the AgentName
func updateLogAgents(logAgentsMap map[string]agent.Agent, logAgents []agent.Agent, match map[string]*Match) {
	logAgentsMap = logAgentConvertFromSliceToMap(logAgents)

	for k, m := range match {
		// If the match's agent does not exist any longer, the remove it record in corresponding match
		if _, exist := logAgentsMap[m.AgentName]; !exist {
			match[k].AgentName = ""
		}
	}
}

// Schedule Algorithm which is used to schedule the match relation between logSources and logAgents
// Return the key of LogSource whose match relation is changed
func schedule(logSourcesMap map[string]api.LogSource, logAgentsMap map[string]agent.Agent, match map[string]*Match) []string {
	keys := make([]string, 0)

	return keys
}

// Return the function that can be used to return newest logSources info from existing logConfigs
func getListLogSourcesFunc(cli *kubernetes.Clientset, logConfigs map[string]api.LogConfig) func() ([]api.LogSource, error) {

	for _, logConfig := range logConfigs {
		name := logConfig.Name
		kind := logConfig.Kind
		namespace := logConfig.Namespace
		labelSelector, _ := getLabelSelector(cli, name, namespace, kind)
		logConfig.LabelSelector = labelSelector
	}

	return func() ([]api.LogSource, error) {
		logSources := make([]api.LogSource, 0)

		for _, logConfig := range logConfigs {
			podList, _ := cli.CoreV1().Pods(logConfig.Namespace).List(metav1.ListOptions{
				LabelSelector: logConfig.LabelSelector,
			})
			for _, pod := range podList.Items {
				logSources = append(logSources, *api.NewLogSource(&pod, &logConfig))
			}
		}

		return logSources, nil
	}
}

func getLabelSelector(cli *kubernetes.Clientset, name, namespace, kind string) (string, error) {
	labels := make(map[string]string)
	switch kind {
	case "deployment":
		obj, err := cli.Extensions().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		labels = obj.Spec.Template.ObjectMeta.Labels
	case "statefulset":
		obj, err := cli.AppsV1beta1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		labels = obj.Spec.Template.ObjectMeta.Labels
	default:
		return "", nil
	}

	kvList := make([]string, 0)
	for k, v := range labels {
		kv := fmt.Sprintf("%s=%s", k, v)
		kvList = append(kvList, kv)
	}

	labelSelector := strings.Join(kvList, ",")
	return labelSelector, nil
}
