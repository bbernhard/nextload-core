package misc

import (
	"gopkg.in/yaml.v2"
)

type Task struct {
	Url string `yaml:"url"`
	Format string `yaml:"format"`
}

func ToTask(in []byte) (Task, error) {
	task := Task{}
    
	err := yaml.Unmarshal(in, &task)
	if err != nil {
		return task, err
	}

	return task, nil
}