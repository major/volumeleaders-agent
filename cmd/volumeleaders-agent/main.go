package main

import (
	"log"

	"github.com/leodido/structcli"
	internalcmd "github.com/major/volumeleaders-agent/internal/cmd"
)

func main() {
	log.SetFlags(0)

	rootCmd, err := internalcmd.NewRootCmd()
	if err != nil {
		log.Fatalln(err)
	}

	structcli.ExecuteOrExit(rootCmd)
}
