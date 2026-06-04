package main

import "github.com/rhevin/apple-compose/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
