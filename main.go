package main

import "github.com/Rhevin/apple-compose/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
