package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

type Config struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	Name              string      `json:"name"`
	Token             string      `json:"token"`
	DefaultKey        string      `json:"default_key"`
	DefaultUser       string      `json:"default_user"`
	DefaultAdditional string      `json:"default_additional"`
	CustomMap         []CustomMap `json:"custom_map"`
}

type CustomMap struct {
	Name       string `json:"name"`
	User       string `json:"user"`
	Key        string `json:"key"`
	Additional string `json:"additional"`
}

var token string
var marker string
var configFile string
var sshconfigFile string
var printonly bool
var config Config

func main() {
	flag.StringVar(&configFile, "config-file", "config.json", "configuration file")
	flag.StringVar(&sshconfigFile, "ssh-config-file", "~/.ssh/config", "ssh configuration file")
	flag.StringVar(&marker, "marker", "HCLOUD-REPLACE", "hcloud replacement marker")
	flag.BoolVar(&printonly, "printonly", false, "don't write to file print out only")
	flag.Parse()

	config := LoadConfiguration(configFile)

	var replaceContent string

	for _, project := range config.Projects {
		replaceContent += sshconfigContent(project)
	}

	setSshConfigFile(replaceContent)
}

func findCustomConfig(customMaps []CustomMap, serverName string) CustomMap {
	for _, customMap := range customMaps {
		if customMap.Name == serverName {
			return customMap
		}
	}
	return CustomMap{}
}

func getReplacementValue(project Project, CustomMap CustomMap, name string) string {
	var ret string
	switch name {
	case "user":
		ret = project.DefaultUser
		if CustomMap.User != "" {
			ret = CustomMap.User
		}
	case "key":
		ret = project.DefaultKey
		if CustomMap.Key != "" {
			ret = CustomMap.Key
		}
	case "additional":
		ret = project.DefaultAdditional
		if CustomMap.Additional != "" {
			ret = CustomMap.Additional
		}
	}
	return ret
}

func sshconfigContent(project Project) string {
	client := hcloud.NewClient(hcloud.WithToken(project.Token))

	servers, err := client.Server.All(context.Background())
	if err != nil {
		log.Fatalf("Error retrieving server for project '%s': %s\n", project.Name, err)
	}
	var configContent string

	for _, server := range servers {
		if server != nil {
			customMap := findCustomConfig(project.CustomMap, server.Name)

			configContent += "Host " + server.Name + "\n"
			configContent += "\tUser " + getReplacementValue(project, customMap, "user") + "\n"
			configContent += "\tHostName " + server.PublicNet.IPv4.IP.String() + "\n"

			key := getReplacementValue(project, customMap, "key")
			if key != "" {
				configContent += "\tIdentityFile " + key + "\n"
			}
			additional := getReplacementValue(project, customMap, "additional")
			if additional != "" {
				configContent += "\t" + additional + "\n"
			}
		}
		configContent += "\n"
	}

	return configContent
}

func setSshConfigFile(replaceContent string) {
	sshconfigFile = replaceHomeDir(sshconfigFile)
	content, err := ioutil.ReadFile(sshconfigFile)
	if err != nil {
		log.Fatal(err)
	}
	originalContent := string(content)

	var newContent string
	if strings.Contains(originalContent, getReplacementToken("start", marker)) {
		newContent = replaceInFileContent(marker, originalContent, replaceContent)
		fmt.Println("Replaced config in " + sshconfigFile)
	} else {
		newContent = addToFileContent(marker, originalContent, replaceContent)
		fmt.Println("Added new config to " + sshconfigFile)
	}
	if printonly {
		fmt.Println(newContent)
	} else {
		WriteToFile(sshconfigFile, newContent)
	}
}

func addToFileContent(replaceToken string, originalContent string, replaceContent string) string {
	originalContent += "\n\n"
	originalContent += getReplacementToken("start", replaceToken) + "\n\n"
	originalContent += replaceContent
	originalContent += getReplacementToken("end", replaceToken) + "\n"
	return originalContent
}

func getReplacementToken(kind string, replaceToken string) string {
	if kind == "start" {
		return "##START " + replaceToken + " ##"
	} else {
		return "##END " + replaceToken + " ##"
	}
}

func replaceInFileContent(replaceToken string, originalContent string, replaceContent string) string {
	var Myregex = "(?s)" + getReplacementToken("start", replaceToken) + ".*" + getReplacementToken("end", replaceToken)
	var re = regexp.MustCompile(Myregex)
	r := getReplacementToken("start", replaceToken) + "\n\n" + replaceContent + getReplacementToken("end", replaceToken)
	s := re.ReplaceAllString(originalContent, r)
	return s
}

func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

func replaceHomeDir(filename string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Can't detect home dir: %s\n", err)
	}
	if strings.Contains(filename, "~/") {
		filename = strings.Replace(filename, "~/", home, 1)

	}
	fmt.Println(filename)
	return filename
}
