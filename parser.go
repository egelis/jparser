package jparser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type RawMessageSet map[string]json.RawMessage

type MetaData struct {
	Path    string
	ParamID string
}

type UnmarshalError struct {
	err     error
	paramID string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("error: %s, param_id: %s", e.err, e.paramID)
}

// nolint:wsl
func ParseParams(data json.RawMessage, meta []MetaData) ([]RawMessageSet, error) {
	if len(data) == 0 || len(meta) == 0 {
		return []RawMessageSet{{}}, nil
	}

	if len(meta) == 1 && meta[0].Path == "" {
		return []RawMessageSet{
			{meta[0].ParamID: data},
		}, nil
	}

	currentPathToNewMeta := make(map[string][]MetaData)
	for i := 0; i < len(meta); i++ {
		currentPath, restOfPath := splitPath(meta[i].Path)
		currentPathToNewMeta[currentPath] = append(currentPathToNewMeta[currentPath],
			MetaData{restOfPath, meta[i].ParamID})
	}

	res := []RawMessageSet{{}}
	for currentPath, newMeta := range currentPathToNewMeta {
		currentRes, err := unmarshalNextLevel(data, newMeta, currentPath)
		if err != nil {
			return nil, err
		}

		res = cartesianProduct(res, currentRes)
	}

	return res, nil
}

// nolint:nestif,gocognit,cyclop
func unmarshalNextLevel(data json.RawMessage, meta []MetaData, currentPath string) ([]RawMessageSet, error) {
	if currentPath == "[]" {
		metaBase, metaAll, metaIndex, metaCount := splitMeta(meta)

		var resAll, resList []RawMessageSet

		if metaAll == nil {
			resAll = []RawMessageSet{{}}
		} else {
			resAll = []RawMessageSet{{metaAll.ParamID: data}}
		}

		var sliceJSON []json.RawMessage
		if err := json.Unmarshal(data, &sliceJSON); err != nil {
			return nil, &UnmarshalError{err, meta[0].ParamID}
		}

		if metaCount != nil {
			resAll = cartesianProduct(resAll,
				[]RawMessageSet{{metaCount.ParamID: json.RawMessage(strconv.Itoa(len(sliceJSON)))}})
		}

		if len(sliceJSON) == 0 {
			resList = []RawMessageSet{{}}
		}

		if metaIndex != nil || len(metaBase) > 0 {
			for i, JSON := range sliceJSON {
				currentRes, err := ParseParams(JSON, metaBase)
				if err != nil {
					return nil, err
				}

				var ixRes []RawMessageSet
				if metaIndex == nil {
					ixRes = []RawMessageSet{{}}
				} else {
					ixRes = []RawMessageSet{{metaIndex.ParamID: json.RawMessage(strconv.Itoa(i))}}
				}

				currentRes = cartesianProduct(currentRes, ixRes)

				resList = append(resList, currentRes...)
			}
		} else {
			resList = []RawMessageSet{{}}
		}

		return cartesianProduct(resList, resAll), nil
	}

	var rawMessage RawMessageSet
	if err := json.Unmarshal(data, &rawMessage); err != nil {
		return nil, &UnmarshalError{err, meta[0].ParamID}
	}

	value, ok := rawMessage[currentPath]
	if !ok {
		return []RawMessageSet{{}}, nil
	}

	res, err := ParseParams(value, meta)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// nolint:gomnd
func splitPath(path string) (currentPath, restOfPath string) {
	res := strings.SplitN(path, ".", 2)
	if len(res) == 1 {
		return res[0], ""
	}

	return res[0], res[1]
}

func cartesianProduct(rawSets1, rawSets2 []RawMessageSet) []RawMessageSet {
	res := make([]RawMessageSet, len(rawSets1)*len(rawSets2))

	for i, set1 := range rawSets1 {
		for _, set2 := range rawSets2 {
			newMap := RawMessageSet{}

			for k, v := range set1 {
				newMap[k] = v
			}

			for k, v := range set2 {
				newMap[k] = v
			}

			res[i] = newMap
			i++
		}
	}

	return res
}

// nolint:revive
func splitMeta(meta []MetaData) (metaBase []MetaData, metaAll, metaIndex, metaCount *MetaData) {
	metaBase = []MetaData{}

	for _, v := range meta {
		v := v
		switch v.Path {
		case "@":
			metaIndex = &v
		case "#":
			metaCount = &v
		case "":
			metaAll = &v
		default:
			metaBase = append(metaBase, v)
		}
	}

	return metaBase, metaAll, metaIndex, metaCount
}
