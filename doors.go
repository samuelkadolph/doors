package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func config(p string, cmd string, args []string) {
	var c *Config
	var err error

	if c, err = LoadConfigFromFile(p); err != nil {
		if _, y := err.(*os.PathError); y {
			c = NewConfig()
		} else {
			fmt.Printf("Error while loading config - %s\n", err)
			os.Exit(1)
		}
	}

	switch cmd {
	case "add":
	case "set":
		configSet(c, args)
	case "remove":
	}

	if err = c.SaveToFile(p); err != nil {
		fmt.Printf("Error while saving config - %s\n", err)
		os.Exit(1)
	}
}

func configSet(c *Config, args []string) {
	for _, a := range args {
		parts := strings.SplitN(a, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("argument is not in the format of name=value: %s\n", a)
			configUsage()
			os.Exit(1)
		}
		switch parts[0] {
		case "feedbacktimeout":
			c.FeedbackTimeout, _ = strconv.Atoi(parts[1])
		case "secret":
			c.Secret = parts[1]
		}
	}
}

func configUsage() {
}

func server(p string, args []string) {
	var c *Config
	var err error
	var host string
	var l net.Listener
	var port int

	set := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	set.Usage = serverUsage
	set.StringVar(&host, "host", "", "Host for the server to listen on")
	set.IntVar(&port, "port", 4567, "Port for the server to listen on")
	set.Parse(args)

	if c, err = LoadConfigFromFile(p); err != nil {
		fmt.Printf("Error while loading config - %s\n", err)
		os.Exit(1)
	}

	for n, i := range c.InterfaceKits {
		if err := i.Load(); err != nil {
			fmt.Printf("Error while loading interface kit (#%d) - %s\n", n+1, err)
			os.Exit(1)
		}
	}

	a := fmt.Sprintf("%s:%d", host, port)
	s := NewServer(c)

	if l, err = s.Listen(a); err != nil {
		fmt.Printf("Error while trying to listen on %s - %s\n", a, err)
		os.Exit(1)
	}

	defer l.Close()
	fmt.Printf("Listening on %s\n", a)

	select {}
}

func serverUsage() {
}

func main() {
	var configPath string

	set := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	set.Usage = mainUsage
	set.StringVar(&configPath, "config", "config.json", "Path to the config file")
	set.Parse(os.Args[1:])

	switch set.Arg(0) {
	case "config:add":
		config(configPath, "add", set.Args()[1:])
	case "config:remove":
		config(configPath, "remove", set.Args()[1:])
	case "config:set":
		config(configPath, "set", set.Args()[1:])
	default:
		server(configPath, set.Args())
	}
}

func mainUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
}
