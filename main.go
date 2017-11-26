package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/kr/pretty"
)

type Default struct {
	SynFlood bool   `uci:"syn_flood"`
	Input    string `uci:"input"`
	Output   string `uci:"output"`
	Forward  string `uci:"forward"`
}
type Zone struct {
	Name    string `uci:"name"`
	Input   string `uci:"input"`
	Output  string `uci:"output"`
	Forward string `uci:"forward"`
}

type Forwarding struct {
	Source      string `uci:"src"`
	Destination string `uci:"dest"`
}

type Firewall struct {
	Default    Default      `uci:"default"`
	Zones      []Zone       `uci:"zone"`
	Forwarding []Forwarding `uci:"forwarding"`
}

type Config struct {
	Name    string
	Type    string
	Options map[string]string
	Lists   map[string][]string
}

func unmarshal(data []byte, v interface{}) error {
	// config := parse(string(data))

	a := reflect.ValueOf(v)
	rv := a.Type()
	// Iterate over the types
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldPwet := a.Field(i)
		tag := field.Tag.Get("uci")
		log.Println("uci : ", tag)
		log.Println("name : ", field.Name)
		log.Println("type : ", field.Type)
		log.Println("kind : ", field.Type.Kind())
		typePwet := field.Type
		if field.Type.String() == "string" {
			log.Println("setting stuff!")
			fieldPwet.SetString("pwet")
		}

		// If it's not a struct, if it's a slice, change the object to the type
		// of the Element
		if typePwet.Kind().String() != "struct" {
			typePwet = typePwet.Elem()
		}

		// Iterate over the inner types
		for j := 0; j < typePwet.NumField(); j++ {
			innerField := typePwet.Field(j)
			// innerFieldPwet := typePwet.Field(j)
			innerTag := innerField.Tag.Get("uci")
			log.Println("\tuci : ", innerTag)
			log.Println("\tname : ", innerField.Name)
			log.Println("\ttype : ", innerField.Type)
			log.Println("-----------------")
			// if innerField.Type.String() == "string" {
			// 	log.Println("setting stuff!")
			// 	innerFieldPwet.SetString("pwet")
			// }
		}
		log.Println("=================")
	}

	return nil
}

func main() {
	dat, err := ioutil.ReadFile("./firewall_small")
	if err != nil {
		panic(err)
	}

	var f Firewall
	err = unmarshal(dat, f)
	log.Println("After the stuff :")
	pretty.Println(f)

	// render(config)
}

func render(configs []*Config) {
	var str string
	for _, conf := range configs {
		str = fmt.Sprintf("%sconfig %s", str, conf.Type)
		if conf.Name != "" {
			str = fmt.Sprintf("%s '%s'", str, conf.Name)
		}
		str = fmt.Sprintf("%s", str)
		for key, value := range conf.Options {
			str = fmt.Sprintf("%s\n\toption %s '%s'", str, key, value)
		}
		for key, value := range conf.Lists {
			for _, list := range value {
				str = fmt.Sprintf("%s\n\tlist %s '%s'", str, key, list)
			}
		}
		str = fmt.Sprintf("%s\n\n", str)
	}
	log.Println("Got the rendered stuff :\n\n", str)
}

func parse(section string) []*Config {
	var comment, err = regexp.Compile("(?:\\s+)?#")
	var config, _ = regexp.Compile("config\\s+([a-z0-9_]+)(?:\\s+'(.+)')?")
	var option, _ = regexp.Compile("option\\s+([a-z0-9_]+)\\s+'(.+)'")
	var list, _ = regexp.Compile("list\\s+([a-z0-9_]+)\\s+'(.+)'")

	if err != nil {
		log.Println("ERR :", err)
		panic("oups")
	}

	if len(section) == 0 {
		return nil
	}

	lines := strings.Split(section, "\n")
	FullConfig := []*Config{}
	var currentConfig *Config

	for _, line := range lines {
		if pwet := comment.FindStringSubmatch(line); len(pwet) > 0 {
			log.Println("Got comment on line", line)
		} else if pwet := config.FindStringSubmatch(line); len(pwet) > 0 {
			if currentConfig != nil {
				FullConfig = append(FullConfig, currentConfig)
			}

			currentConfig = new(Config)
			currentConfig.Options = make(map[string]string)
			currentConfig.Lists = make(map[string][]string)
			currentConfig.Type = pwet[1]
			if len(pwet) > 2 {
				currentConfig.Name = pwet[2]
			}
		} else if pwet := option.FindStringSubmatch(line); len(pwet) > 0 {
			currentConfig.Options[pwet[1]] = pwet[2]
		} else if pwet := list.FindStringSubmatch(line); len(pwet) > 0 {
			currentConfig.Lists[pwet[1]] = append(currentConfig.Lists[pwet[1]], pwet[2])
		} else {
			log.Println("Got nothing on line", line)
		}
	}
	if currentConfig != nil {
		FullConfig = append(FullConfig, currentConfig)
	}

	log.Println("==========================================")
	pretty.Println(FullConfig)
	return FullConfig
}
