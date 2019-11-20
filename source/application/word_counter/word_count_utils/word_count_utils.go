package word_count_utils

import(
	"strings"
	"unicode"
)

type WordCount struct {
	Word       string
	Occurrence int
}

func StringSplit(text string) []WordCount {
	text = strings.ToLower(text)
	text = strings.Replace(text, "\n", " ", -1)
	words := strings.Split(text, " ")
	
	var counted []WordCount
	for i := range words {
		 trimmed_word := strings.TrimFunc(words[i], func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		if trimmed_word != "" { 
			counted = append(counted, WordCount{trimmed_word, 1}) 
		}
	}
	return counted
}

func CountWords(words []WordCount) []WordCount {
	var counted []WordCount
	for i := range words {
		var j int
		for j = 0; j < len(counted); j++ {
			if words[i].Word == counted[j].Word {
				counted[j].Occurrence += words[j].Occurrence
				break
			}
		}
		if j == len(counted) {
			counted = append(counted, words[i])
		}
	}
	return counted
}

