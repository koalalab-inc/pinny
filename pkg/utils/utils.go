package utils

import "regexp"

func MatchNamedRegex(regex *regexp.Regexp, str string) (bool, map[string]string) {
	matchesArray := regex.FindStringSubmatch(str)
	matches := make(map[string]string)
	if len(matchesArray) > 0 {
		for i, name := range regex.SubexpNames() {
			if i != 0 && name != "" {
				matches[name] = matchesArray[i]
			}
		}
		return true, matches
	}
	return false, nil
}
