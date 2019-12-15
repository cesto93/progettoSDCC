package utility

import "log"

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func CheckErrorNonFatal(err error) {
	if err != nil {
		log.Printf("NON FATAL ERROR: %v", err)
	}
}
