package common

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/kballard/go-shellquote"
	"gopkg.in/yaml.v3"
	"strings"
)

func SplitCommand(input string) ([]string, error) {
	return shellquote.Split(input)
}

func UnmarshalBuildDefinition(content []byte, vars []entity.UserVariable) (entity.BuildDefinitionContent, error) {
	s := string(content)
	ReplaceVariables(&s, vars)

	var bdc entity.BuildDefinitionContent
	if err := yaml.Unmarshal(content, &bdc); err != nil {
		return entity.BuildDefinitionContent{}, err
	}

	return bdc, nil
}

func ReplaceVariables(content *string, variables []entity.UserVariable) {
	if len(variables) == 0 {
		return
	}

	for _, v := range variables {
		*content = strings.ReplaceAll(*content, fmt.Sprintf("${%s}", v.Variable), v.Value)
	}
}
