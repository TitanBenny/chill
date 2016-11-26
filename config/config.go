package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	colorlog "chill/log"
)

type Configs struct {
	Directory string
	Patterns  []string
	Command   []string
}

var log = colorlog.NewLog()

var configfile = flag.String("config", ".chill.json", "Config file")
var directory = flag.String("dir", "", "Directory to watch")
var pattern = flag.String("pattern", "", "Patterns to filter filenames")
var saveconf = flag.Bool("save", false, "Save options to conf")

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:        %s [options] [command]\n", os.Args[0])
		flag.PrintDefaults()

	}
}

func GetConfigs() Configs {
	flag.Parse()

	conf := readConfigFile()

	if dir := parseDirectory(); dir != "" {
		conf.Directory = dir
	}

	if patterns := parsePatterns(); patterns != nil {
		conf.Patterns = patterns
	}

	if command := parseCommand(); command != nil {
		conf.Command = command
	}

	return conf
}

func readConfigFile() Configs {
	file, err := os.Open(*configfile)
	defer file.Close()
	var conf Configs
	if err != nil {
		log.Fatalf("Faild to open config file: %s", err.Error())
	} else {
		log.Infof("Reading options from %s", *configfile)
		if err := json.NewDecoder(file).Decode(&conf); err != nil {
			log.Fatalf("Failed to parse config file: %s", err.Error())
		} else {
			return conf
		}
	}

	return Configs{".", []string{"*"}, []string{}}
}

func parseDirectory() string {
	dir := *directory
	if info, err := os.Stat(dir); err == nil {
		if !info.IsDir() {
			log.Fatalf("%s is not a directory", dir)
		}
	}
	return dir
}

func parsePatterns() []string {
	pat := strings.Trim(*pattern, " ")
	if pat == "" {
		return nil
	}

	patternSep, _ := regexp.Compile("[,\\s]+")

	patternMap := make(map[string]bool)
	ret := []string{}

	for _, part := range patternSep.Split(pat, -1) {
		patternMap[part] = true
	}
	for part := range patternMap {
		ret = append(ret, part)
	}

	return ret
}

func parseCommand() []string {
	if flag.NArg() == 0 {
		return nil
	}
	return flag.Args()
}

func saveConfigFile(conf Configs) {
	log.Infof("Saving options to %s", *configfile)
	file, err := os.Create(*configfile)
	defer file.Close()

	if err != nil {
		log.Fatalf("Failed to open config file:", err)
	}
	if bytes, err := json.MarshalIndent(conf, "", "  "); err == nil {
		if _, err := file.Write(bytes); err != nil {
			log.Fatalf("Failed to write config file: %s", err.Error())
		}
	} else {
		log.Fatalf("Failed to encode options: %s", err.Error())
	}
}
