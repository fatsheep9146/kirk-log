package logkit

import "fmt"

const (
	LogkitManagerVolumeMountPath = "/logkit"
	LogkitAgentVolumeMountPath   = "/logkit"
)

func (l *LogkitAgentManagerImpl) Deploy() error {

	fmt.Printf("Not Implemented yet")

	return nil
}

func getDeployName(name string) string {
	return fmt.Sprintf("logkit-%s", name)
}

func getLogkitAgentConfDir(podname string) string {
	return fmt.Sprintf("%s/%s", LogkitAgentVolumeMountPath, podname)
}
