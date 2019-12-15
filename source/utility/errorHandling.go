package utility

import "log"

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func CheckErrorNonFatal(err error) {
	if err != nil {
		log.Print(err)
	}
}
