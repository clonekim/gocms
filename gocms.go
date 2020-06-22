package main

import (
	"gocms/cmd"
	"os"
	"log"
)

func main() {
	//cli 파라미터 검증
	err := cmd.Setup().Run(os.Args);

	if err != nil {
		log.Fatal(err)
	}

}
