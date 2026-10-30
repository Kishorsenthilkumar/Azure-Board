package pipeline

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type PipelineConfig struct {
	Name    string          `yaml:"name"`
	Trigger PipelineTrigger `yaml:"trigger"`
	Jobs    map[string]Job  `yaml:"jobs"`
}

type PipelineTrigger struct {
	Branches []string `yaml:"branches"`
}

type Job struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

func ParseConfig(r io.Reader) (PipelineConfig, error) {

	var config PipelineConfig

	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)

	if err := decoder.Decode(&config); err != nil {
		return PipelineConfig{}, err
	}
	return config, nil
}

func ValidateConfig(config PipelineConfig) error {
	if strings.TrimSpace(config.Name) == "" {
		return fmt.Errorf("pipeline name is required")
	}

	if len(config.Jobs) == 0 {
		return fmt.Errorf("pipeline must contain at least one job")
	}

	for jobName, job := range config.Jobs {
		if strings.TrimSpace(jobName) == "" {
			return fmt.Errorf("job name is required")
		}

		if len(job.Steps) == 0 {
			return fmt.Errorf("job %q must contain at least one step", jobName)
		}

		for i, step := range job.Steps {
			if strings.TrimSpace(step.Name) == "" {
				return fmt.Errorf("job %q: step %d name is required", jobName, i+1)
			}

			if strings.TrimSpace(step.Run) == "" {
				return fmt.Errorf("job %q: step %q command is required", jobName, step.Name)
			}
		}
	}

	return nil
}

func LoadConfig(r io.Reader) (PipelineConfig, error) {
	config, err := ParseConfig(r)
	if err != nil {
		return PipelineConfig{}, fmt.Errorf("parsing pipeline.yml: %w", err)
	}
	err = ValidateConfig(config)
	if err != nil {
		return PipelineConfig{}, fmt.Errorf("pipeline validation failed: %w", err)
	}
	return config, nil
}

func LoadConfigFromFile(filepath string) (PipelineConfig, error) {

	file, err := os.Open(filepath)
	if err != nil {
		return PipelineConfig{}, fmt.Errorf("opening pipeline config %q: %w", filepath, err)
	}
	defer file.Close()

	return LoadConfig(file)
}
