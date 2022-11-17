package helpers

import "sort"

func GetUniqueStrings(items []string) []string {
	var (
		uniqueList []string
		keys       = make(map[string]bool)
	)

	for _, item := range items {
		if _, value := keys[item]; !value {
			keys[item] = true
			uniqueList = append(uniqueList, item)
		}
	}
	return uniqueList
}

func GetSortedMapKeys(inputMap map[string]string) []string {
	var keys []string
	for k := range inputMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func GetSortedMapKeysBool(inputMap map[string]bool) []string {
	var keys []string
	for k := range inputMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
