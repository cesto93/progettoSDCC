package wordCountUtils

import(
	"strings"
	"unicode"
	"fmt"
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

/*func CountWords(words []WordCount) []WordCount {
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
}*/

func CountWords(words []WordCount) []WordCount {
	var counted []WordCount
	k := 0
	for i, _ := range words {
		if words[i].Occurrence != 0 {
			counted = append(counted, words[i])
			for j := i + 1; j < len(words); j++ {
				if counted[k].Word == words[j].Word {
					counted[k].Occurrence = counted[k].Occurrence + words[j].Occurrence
					words[j].Occurrence = 0
					//words = append(words[:i], words[i+1:]...)
				}
			}
			k++
		}
	}
	return counted
}

func ToString(wd []WordCount) string {
	text := ""
	for _, w:= range(wd) {
		text = text + fmt.Sprintf("%v %v\n", w.Word, w.Occurrence)
		
	}
	return text
}

func CountTotalWords(words []WordCount) int {
	res := 0
	for _,word := range words {
		res += word.Occurrence
	}
	return res
}

