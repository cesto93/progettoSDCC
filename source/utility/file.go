package utility
import (
	"encoding/json"
	"io/ioutil"
	"fmt"
)

func ImportJson(path string, pointer interface{}) error{
	file_json, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file_json, pointer)
	if err != nil {
		return fmt.Errorf("Json Unmarshal error %v", err)
	}
	return nil
}

func ExportJson(path string, data interface{}) error {
	file, _ := json.MarshalIndent(data, "", " ")
	return ioutil.WriteFile(path, file, 0644)
}

func readWordfiles(paths []string) ([]string, error) {
	var texts []string
	for i := range paths {
		file, err := ioutil.ReadFile(paths[i])
		if err != nil {
			return nil, fmt.Errorf("File reading error %v", err)
		}
		texts = append(texts, string(file))
	}
	return texts, nil
}
