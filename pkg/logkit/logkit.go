package logkit

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/fatsheep9146/kirklog/pkg/agent"
	"github.com/fatsheep9146/kirklog/pkg/api"
)

type LogkitAgentManagerImpl struct {
	Cli       *kubernetes.Clientset
	Name      string
	Namespace string
}

func NewLogkitAgentManager(cfg *agent.AgentManagerConfig) agent.AgentManager {
	return &LogkitAgentManagerImpl{
		Cli:       cfg.Cli,
		Name:      cfg.Name,
		Namespace: cfg.Namespace,
	}
}

func NewLogkitAgent(pod *v1.Pod) *agent.Agent {
	return &agent.Agent{
		Name:     pod.Name,
		ConfPath: getLogkitAgentConfDir(pod.Name),
	}
}

func (l *LogkitAgentManagerImpl) List() ([]agent.Agent, error) {
	agents := make([]agent.Agent, 0)
	namespace := l.Namespace
	name := l.Name
	deployname := getDeployName(name)

	deploy, err := l.Cli.ExtensionsV1beta1().Deployments(namespace).Get(deployname, metav1.GetOptions{})
	if err != nil {
		return agents, err
	}

	labels := deploy.Spec.Template.Labels
	labelSelectors := make([]string, 0)
	for k, v := range labels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
	}
	labelSelectorsStr := strings.Join(labelSelectors, ",")

	pods, err := l.Cli.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelectorsStr,
	})

	for _, pod := range pods.Items {
		agents = append(agents, *NewLogkitAgent(&pod))
	}

	return agents, nil
}

// Add the log config file of one logSource to logAgent agent
func (l *LogkitAgentManagerImpl) AddConfig(logSource *api.LogSource, agent string) error {
	// Generate the config file from logSource info
	config, err := renderConfig(logSource)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", getLogkitAgentConfDir(agent), getConfigFileName(logSource))

	err = ioutil.WriteFile(filePath, []byte(config), 0644)
	if err != nil {
		return err
	}
	// Create log config to this log agent
	return nil
}

// Delete the log config file of one logSource from logAgent agent
func (l *LogkitAgentManagerImpl) DelConfig(logSource *api.LogSource, agent string) error {
	filePath := fmt.Sprintf("%s/%s", getLogkitAgentConfDir(agent), getConfigFileName(logSource))

	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

func (l *LogkitAgentManagerImpl) CheckLag(logSource *api.LogSource, agent string) bool {
	return true
}

func (l *LogkitAgentManagerImpl) GetAgentNameFromConf(confpath string) string {
	return ""
}
