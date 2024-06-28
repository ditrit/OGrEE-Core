package controllers

import (
	"cli/models"
	"fmt"
	"path"
	"strings"
)

func PWD() string {
	println(State.CurrPath)
	return State.CurrPath
}

func (controller Controller) UnfoldPath(path string) ([]string, error) {
	if strings.Contains(path, "*") || models.PathHasLayer(path) {
		_, subpaths, err := controller.GetObjectsWildcard(path, nil, nil)
		return subpaths, err
	}

	if path == "_" {
		return State.ClipBoard, nil
	}

	return []string{path}, nil
}

func (controller Controller) SplitPath(pathStr string) (models.Path, error) {
	for _, prefix := range models.PathPrefixes {
		if strings.HasPrefix(pathStr, string(prefix)) {
			var id string
			if prefix == models.VirtualObjsPath && strings.HasPrefix(pathStr, prefix+"#") {
				// virtual root layer, keep the virtual node
				id = pathStr[1:]
			} else {
				id = pathStr[len(prefix):]
			}
			id = strings.ReplaceAll(id, "/", ".")

			var layer models.Layer
			var err error

			id, layer, err = controller.GetLayer(id)
			if err != nil {
				return models.Path{}, err
			}

			return models.Path{
				Prefix:   prefix,
				ObjectID: id,
				Layer:    layer,
			}, nil
		}
	}

	return models.Path{}, fmt.Errorf("invalid object path")
}

func (controller Controller) GetParentFromPath(path string, ent int, isValidate bool) (string, map[string]any, error) {
	var parent map[string]any
	parentId := ""
	if ent != models.SITE && ent != models.STRAY_DEV {
		if isValidate {
			parentId = models.GetObjectIDFromPath(path)
		} else {
			var err error
			parent, err = controller.PollObject(path)
			if err != nil {
				return parentId, nil, err
			}
			if parent == nil && (ent != models.DOMAIN || path != "/Organisation/Domain") &&
				ent != models.VIRTUALOBJ {
				return parentId, nil, fmt.Errorf("parent not found")
			}
			if parent != nil {
				parentId = parent["id"].(string)
			}
		}
	}
	return parentId, parent, nil
}

func TranslatePath(p string, acceptSelection bool) string {
	if p == "" {
		p = "."
	}
	if p == "_" && acceptSelection {
		return "_"
	}
	if p == "-" {
		return State.PrevPath
	}
	var output_words []string
	if p[0] != '/' {
		outputBase := State.CurrPath
		if p[0] == '-' {
			outputBase = State.PrevPath
		}

		output_words = strings.Split(outputBase, "/")[1:]
		if len(output_words) == 1 && output_words[0] == "" {
			output_words = output_words[0:0]
		}
	} else {
		p = p[1:]
	}
	input_words := strings.Split(p, "/")
	for i, word := range input_words {
		if word == "." || (i == 0 && word == "-") {
			continue
		} else if word == ".." {
			if len(output_words) > 0 {
				output_words = output_words[:len(output_words)-1]
			}
		} else {
			output_words = append(output_words, word)
		}
	}
	if len(output_words) > 0 {
		if output_words[0] == "P" {
			output_words[0] = "Physical"
		} else if output_words[0] == "L" {
			output_words[0] = "Logical"
		} else if output_words[0] == "O" {
			output_words[0] = "Organisation"
		}
	}
	return path.Clean("/" + strings.Join(output_words, "/"))
}
