package utility
import (
	"encoding/json"
	"io/ioutil"
	"fmt"
)
func import_json(path string, pointer interface{}) {
	file_json, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	err = json.Unmarshal(file_json, pointer)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func read_wordfiles(paths []string) []string {
	var texts []string
	for i := range paths {
		file, err := ioutil.ReadFile(paths[i])
		if err != nil {
			fmt.Println("File reading error", err)
			return nil
		}
		texts = append(texts, string(file))
	}
	return texts
}
