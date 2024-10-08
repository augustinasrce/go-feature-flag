package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/model/dto"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"gopkg.in/yaml.v3"
)

// FlagConverter is a cli to convert your old file to a new format
type FlagConverter struct {
	InputFile    string
	InputFormat  string
	OutputFormat string
}

func (f *FlagConverter) Migrate() ([]byte, error) {
	// Read content of the file
	content, err := os.ReadFile(f.InputFile)
	if err != nil {
		return nil, fmt.Errorf("file %v is impossible to find", f.InputFile)
	}

	flags, err := f.unmarshall(content)
	if err != nil {
		return nil, err
	}

	convertedFlag := f.convert(flags)
	newFileContent, err := f.marshall(convertedFlag)
	return newFileContent, err
}

func (f *FlagConverter) unmarshall(content []byte) (map[string]dto.DTO, error) {
	var flags map[string]dto.DTO
	var err error
	switch strings.ToLower(f.InputFormat) {
	case "toml":
		err = toml.Unmarshal(content, &flags)
	case "json":
		err = json.Unmarshal(content, &flags)
	case "yaml":
		err = yaml.Unmarshal(content, &flags)
	default:
		err = fmt.Errorf("invalid input format %s", f.InputFormat)
	}
	if err != nil {
		return nil, err
	}

	return flags, nil
}

func (f *FlagConverter) convert(flags map[string]dto.DTO) map[string]dto.DTO {
	convertedFlags := make(map[string]dto.DTO, len(flags))
	for k, v := range flags {
		// we don't set a logger on purpose here, because this is not accurate in the migration context.
		logger := fflog.FFLogger{}
		convertedFlags[k] = convertToDto(v.Convert(&logger, k))
	}
	return convertedFlags
}

func (f *FlagConverter) marshall(convertedFlags map[string]dto.DTO) ([]byte, error) {
	switch strings.ToLower(f.OutputFormat) {
	case "toml":
		buf := new(bytes.Buffer)
		_ = toml.NewEncoder(buf).Encode(convertedFlags)
		return buf.Bytes(), nil
	case "json":
		return json.MarshalIndent(convertedFlags, "", "  ")
	default:
		return yaml.Marshal(convertedFlags)
	}
}

func convertToDto(internalFlag flag.InternalFlag) dto.DTO {
	var experimentation *dto.ExperimentationDto
	if internalFlag.Experimentation != nil {
		experimentation = &dto.ExperimentationDto{
			Start: internalFlag.Experimentation.Start,
			End:   internalFlag.Experimentation.End,
		}
	}

	return dto.DTO{
		TrackEvents: internalFlag.TrackEvents,
		Disable:     internalFlag.Disable,
		Version:     internalFlag.Version,
		DTOv1: dto.DTOv1{
			BucketingKey:    internalFlag.BucketingKey,
			Variations:      internalFlag.Variations,
			Rules:           internalFlag.Rules,
			DefaultRule:     internalFlag.DefaultRule,
			Scheduled:       internalFlag.Scheduled,
			Experimentation: experimentation,
		},
	}
}
