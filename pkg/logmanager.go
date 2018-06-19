package logmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"

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
	LogConfigs map[string]*api.LogConfig

	// map for all log entries
	LogSources map[string]*api.LogSource

	// map for all log agent
	LogAgents map[string]*agent.Agent

	// map for the match relation between logSource and logAgent
	Match map[string]*Match

	// the working queue to store the logSource wait to be processed
	Queue workqueue.RateLimitingInterface

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
	logger := log.WithFields(log.Fields{
		"func": "NewLogManager",
	})

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
		logger.Fatalf("Load config file from dir %s failed, err: %v", cfg.LogConfigDir, err)
	}
	logConfigsMap := logConfigConvertFromSliceToMap(logConfigs)
	logger.Info("Successfully load log configs")

	// Create logsources spec map from logConfigs
	listLogSourcesFunc := getListLogSourcesFunc(cli, logConfigsMap)
	logSources, err := listLogSourcesFunc()
	if err != nil {
		logger.Fatalf("List log sources from the config of log failed, err:%+v", err)
	}
	logger.Info("Successfully get current log sources")

	// Check and create LogAgentManager and the deployment of log collector if not exist
	logAgentManager := newAgentManager(agent.AgentType(cfg.AgentType), &agent.AgentManagerConfig{
		Name:       cfg.Name,
		Namespace:  cfg.Namespace,
		LogConfigs: logConfigs,
		Cli:        cli,
	})
	logger.Infof("Successfully create AgentManager of type %s", cfg.AgentType)

	logAgents, err := logAgentManager.List()
	if err != nil {
		logger.Fatal("List agent pods failed, err: %+v", err)
	}
	if len(logAgents) == 0 {
		logger.Info("List no active log agent pods, then we should deploy a new log agent service")
		err = logAgentManager.Deploy()
		if err != nil {
			logger.Fatal("Deploy new log agent service failed, err: %v", err)
		}
		logAgents, err = logAgentManager.List()
		if err != nil {
			logger.Fatal("List agent pods failed, err: %+v", err)
		}
	}
	logger.Info("Successfully list the log agents instance")

	// ToDo: Restore the logsources map status from current situations in case this is a restart

	return &LogManager{
		LogConfigs:      logConfigsMap,
		LogSources:      logSourceConvertFromSliceToMap(logSources),
		LogAgents:       logAgentConvertFromSliceToMap(logAgents),
		LogAgentManager: logAgentManager,
		Match:           make(map[string]*Match),
		Queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "logsource"),
		Cli:             cli,
	}
}

func (lm *LogManager) Run() {
	stop := make(chan struct{})
	logger := log.WithFields(log.Fields{
		"func": "Run",
	})
	logger.Info("Start the LogManager main loop")

	// This function choose whether to rearrange the match relations between logSource and logAgent
	go lm.syncInfo()

	// Info: Start worker to handle the message in queuexx
	logger.Info("Start the worker function to deal with logSource")
	go wait.Until(lm.worker, time.Second, stop)
	// Start the goroutine to check the lag of each logSource

	<-stop
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
	logger := log.WithFields(log.Fields{
		"func": "syncInfo",
	})

	logger.Info("Start the main loop of sync info of logSources and logAgents")
	// Get current logSources
	listLogSourcesFunc := getListLogSourcesFunc(lm.Cli, lm.LogConfigs)
	// Get current logAgents
	listLogAgentsFunc := lm.LogAgentManager.List

	for {
		// Get current logSources
		logSources, err := listLogSourcesFunc()
		if err != nil {
			logger.Errorf("List newest log sources failed, err: %v", err)
			continue
		}
		logger.Infof("List newest log sources succeeded, list %d logSources", len(logSources))
		for i, logSource := range logSources {
			logger.Debugf("LogSource %d: %s, detail: %v", i, logSource.Meta.Name, logSource)
		}

		// Get current logAgents
		logAgents, err := listLogAgentsFunc()
		if err != nil {
			logger.Errorf("List newest log agents failed, err: %v", err)
			continue
		}
		logger.Infof("List newest log agents succeeded, list %d agents", len(logAgents))
		for i, logAgent := range logAgents {
			logger.Debugf("LogAgent %d: %s, detail: %v", i, logAgent.Name, logAgent)
		}

		// Update the logSources and logAgents map
		updateLogSources(lm.LogSources, logSources, lm.Match)
		updateLogAgents(lm.LogAgents, logAgents, lm.Match)
		logger.Info("Update logSources and logAgents succeeded")

		// Update the match relation between logSource and logAgent
		logsources := updateMatch(lm.LogSources, lm.LogAgents, lm.Match)
		logger.Info("Update match succeeded")

		// Enqueue the LogSources that are needed to be synced
		for _, logsource := range logsources {
			lm.Queue.Add(logsource.Meta.Name)
			log.Debugf("Logsource %s is added to queue", logsource.Meta.Name)
		}

		time.Sleep(3 * time.Second)
	}
}

func (lm *LogManager) worker() {
	for lm.processNextWorkItem() {
	}
}

func (lm *LogManager) processNextWorkItem() bool {
	logger := log.WithFields(log.Fields{
		"func": "processNextWorkItem",
	})
	key, quit := lm.Queue.Get()
	if quit {
		return false
	}

	logger.Infof("Ready to handle the item %s", key)

	defer lm.Queue.Done(key)
	if flag, err := lm.sync(key.(string)); err != nil {
		// utilruntime.HandleError(fmt.Errorf("Error syncing StatefulSet %v, requeuing: %v", key.(string), err))
		lm.Queue.AddRateLimited(key)
	} else {
		if flag {
			lm.Queue.Forget(key)
		} else {
			lm.Queue.AddRateLimited(key)
		}
	}
	logger.Infof("Finish handling the item %s", key)
	return true
}

// flag indicates wheter the logsource is done processing.
func (lm *LogManager) sync(key string) (flag bool, err error) {
	logger := log.WithFields(log.Fields{
		"func": "sync",
		"key":  key,
	})

	// Get logSource entry of this key from map
	action := judgeAction(lm.Match[key])
	logger.Infof("Start handle with the logSource %s, action is %v", key, action)
	// Info: Handle the logSource action
	switch action {
	case LogSourceAdd:
		flag, err = lm.logSourceAddFunc(key)
	case LogSourceDel:
		flag, err = lm.logSourceDelFunc(key)
	case LogSourceMov:
		flag, err = lm.logSourceMovFunc(key)
	}
	logger.Infof("Handle logSource %s done", key)

	return flag, err
}

// Remove the logSource from logSourcesMap and logSource log dir and Match
func (lm *LogManager) removeLogSource(logSource *api.LogSource) error {
	logger := log.WithFields(log.Fields{
		"func": "removeLogSource",
		"key":  logSource.Meta.Name,
	})
	// Remove log dir
	err := os.RemoveAll(logSource.GetLogDir())
	if err != nil {
		logger.Errorf("Remove log dir failed, err: %v", err)
		return err
	}
	logger.Infof("Remove log dir %s succeeded", logSource.GetLogDir())

	key := logSource.Meta.Name
	delete(lm.LogSources, key)
	delete(lm.Match, key)
	logger.Info("Remove logSource meta data from logSources map and match")

	return nil
}

// Create logconfig objects from the files under the path dir
func loadLogConfig(path string) ([]api.LogConfig, error) {
	logger := log.WithFields(log.Fields{
		"func": "loadLogConfig",
	})

	logConfigs := make([]api.LogConfig, 0)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Errorf("Read dir failed %s, err: %v", path, err)
		return logConfigs, err
	}

	for _, file := range files {
		raw, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, file.Name()))
		if err != nil {
			logger.Errorf("Read file %s failed, err: %v", fmt.Sprintf("%s/%s", path, file.Name()), err)
			continue
		}
		logConfig := &api.LogConfig{}
		err = json.Unmarshal(raw, logConfig)
		if err != nil {
			logger.Errorf("Unmarshal file %s failed, err: %v", fmt.Sprintf("%s/%s", path, file.Name()), err)
			continue
		}
		logConfigs = append(logConfigs, *logConfig)
	}

	logger.Info("Load all log configs")
	for _, logConfig := range logConfigs {
		logger.Debugf("Successfully load log config %+v", logConfig)
	}

	return logConfigs, nil
}

// Update the map of logSources and match according the newest logsource list
// If there is a new logSource, then add it to logSourcesMap, and create a new match with podname, no conf, no agentname
// If there is a deleted logSource, then modify the match of this logSource, remove the PodName
func updateLogSources(logSourcesMap map[string]*api.LogSource, logSources []api.LogSource, match map[string]*Match) {
	logger := log.WithFields(log.Fields{
		"func": "updateLogSources",
	})

	visited := make(map[string]bool)

	for i, logSource := range logSources {
		if _, exist := logSourcesMap[logSource.Meta.Name]; !exist {
			logger.Infof("Found a new logSource %s, add it to logSources map", logSource.Meta.Name)
			logSourcesMap[logSource.Meta.Name] = &logSources[i]
		}
		if _, exist := match[logSource.Meta.Name]; !exist {
			logger.Infof("Found a new not matched logSource %s, add it to match", logSource.Meta.Name)
			match[logSource.Meta.Name] = &Match{
				PodName: logSource.Spec.PodName,
			}
		}
		visited[logSource.Meta.Name] = true
	}

	for logSourceName, _ := range logSourcesMap {
		if _, exist := visited[logSourceName]; !exist {
			logger.Infof("Found a deleted logSource %s, delete it from logSources map", logSourceName)
			match[logSourceName].PodName = ""
		}
	}
}

// Update the map of logAgents and match according the newest logsource list
// If there is a deleted logAgent, then modify the match which has this logAgent, remove the AgentName
func updateLogAgents(logAgentsMap map[string]*agent.Agent, logAgents []agent.Agent, match map[string]*Match) {
	logger := log.WithFields(log.Fields{
		"func": "updateLogAgents",
	})
	curLogAgentsMap := logAgentConvertFromSliceToMap(logAgents)
	deletedAgentName := make(map[string]bool)

	for k, m := range match {
		if m.AgentName != "" {
			if _, exist := curLogAgentsMap[m.AgentName]; !exist {
				logger.Infof("Found a logSource %s with no-more-existed agent %s, delete its info", k, match[k].AgentName)
				deletedAgentName[match[k].AgentName] = true
				match[k].AgentName = ""
			}
		}
	}

	for k, _ := range curLogAgentsMap {
		if _, exist := logAgentsMap[k]; !exist {
			logger.Infof("Found a new log agent %s", k)
			logAgentsMap[k] = curLogAgentsMap[k]
		}
	}

	// delete the lost agent name from logAgentsMap
	for k, _ := range deletedAgentName {
		logger.Infof("Delete a non-existed log agent %s", k)
		delete(logAgentsMap, k)
	}
}

// Schedule Algorithm which is used to schedule the match relation between logSources and logAgents
// Return the key of LogSource whose match relation is changed
func updateMatch(logSourcesMap map[string]*api.LogSource, logAgentsMap map[string]*agent.Agent, match map[string]*Match) []api.LogSource {
	logger := log.WithFields(log.Fields{
		"func": "updateMatch",
	})
	logsources := make([]api.LogSource, 0)

	// First visit all match found all match need to be added into the queue
	for k, m := range match {
		needAdded := false
		needSchedule := false
		if m.PodName != "" && m.AgentName == "" && m.ConfPath == "" {
			logger.Infof("LogSource %s is a new added logSource", k)
			// Info: A added logSource
			needAdded = true
			needSchedule = true
		} else if m.PodName == "" && m.AgentName != "" && m.ConfPath != "" {
			logger.Infof("LogSource %s is a deleted logSource", k)
			// Info
			needAdded = true
		} else if m.PodName != "" && m.AgentName == "" && m.ConfPath != "" {
			logger.Infof("LogSource %s is a agent-changed logSource", k)
			needAdded = true
			needSchedule = true
		}

		if needSchedule {
			logger.Infof("LogSource %s needs to be scheduled or re-scheduled", k)
			schedule(logSourcesMap[k], logAgentsMap, match)
			logger.Infof("LogSource %s is scheduled or re-scheduled to agent %s", logSourcesMap[k].Meta.Name, match[k].AgentName)
		}
		if needAdded {
			logger.Infof("LogSource %s is needs to be enqueued", logSourcesMap[k].Meta.Name)
			logsources = append(logsources, *logSourcesMap[k])
		}
	}

	return logsources
}

// Return the function that can be used to return newest logSources info from existing logConfigs
func getListLogSourcesFunc(cli *kubernetes.Clientset, logConfigs map[string]*api.LogConfig) func() ([]api.LogSource, error) {
	logger := log.WithFields(log.Fields{
		"func": "getListLogSourcesFunc",
	})

	for k, logConfig := range logConfigs {
		name := logConfig.Name
		kind := logConfig.Kind
		namespace := logConfig.Namespace
		labelSelector, err := getLabelSelector(cli, name, namespace, kind)
		if err != nil {
			logger.Errorf("getLabelSelector for logConfig %s failed, err: %v", logConfig.Name, err)
		}
		logConfigs[k].LabelSelector = labelSelector
	}

	return func() ([]api.LogSource, error) {
		logSources := make([]api.LogSource, 0)

		for _, logConfig := range logConfigs {
			podList, _ := cli.CoreV1().Pods(logConfig.Namespace).List(metav1.ListOptions{
				LabelSelector: logConfig.LabelSelector,
			})
			for _, pod := range podList.Items {
				logSources = append(logSources, *api.NewLogSource(&pod, logConfig))
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
		labels = obj.Spec.Selector.MatchLabels
	case "statefulset":
		obj, err := cli.AppsV1beta1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		labels = obj.Spec.Selector.MatchLabels
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
