package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-squads/reuni-cli/helper"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

type configView struct {
	Version       int               `json:"version"`
	Configuration map[string]string `json:"configuration"`
	Created_by    string            `json:"created_by"`
}

type versionView struct {
	Version int `json:"version"`
}

var version int
var configKey, configVal string

var configurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Manage Configuration of your namespace",
	Long:  `Manage Configuration of your namespace. Organization, service and namespace name are required`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !strings.EqualFold(cmd.CalledAs(), "namespace") {
			key = getToken()
			if len(organizationName) < 1 {
				fmt.Println("Invalid Organization")
				os.Exit(1)
			}

			if len(serviceName) < 1 {
				fmt.Println("Invalid Service")
				os.Exit(1)
			}

			if len(namespaceName) < 1 {
				fmt.Println("Invalid Namespace")
				os.Exit(1)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listConfigurationCmd = &cobra.Command{
	Use:   "list",
	Short: "Display version list of configuration",
	Run: func(cmd *cobra.Command, args []string) {
		displayListVersions()
	},
}

var displayConfugurationCmd = &cobra.Command{
	Use:   "show",
	Short: "Display detail of configuration",
	Run: func(cmd *cobra.Command, args []string) {
		displayConfig()
	},
}

var updateAllConfigurationCmd = &cobra.Command{
	Use:   "update",
	Short: "Update configuration",
	Run: func(cmd *cobra.Command, args []string) {
		updateAllConfig()
	},
}

var setConfigurationCmd = &cobra.Command{
	Use:   "set",
	Short: "Update configuration",
	Run: func(cmd *cobra.Command, args []string) {
		setConfig()
	},
}

var unsetConfigurationCmd = &cobra.Command{
	Use:   "unset",
	Short: "Update configuration",
	Run: func(cmd *cobra.Command, args []string) {
		unsetConfig()
	},
}

func init() {
	rootCmd.AddCommand(configurationCmd)
	configurationCmd.AddCommand(listConfigurationCmd)
	configurationCmd.AddCommand(displayConfugurationCmd)
	configurationCmd.AddCommand(updateAllConfigurationCmd)
	configurationCmd.AddCommand(setConfigurationCmd)
	configurationCmd.AddCommand(unsetConfigurationCmd)

	configurationCmd.PersistentFlags().StringVarP(&organizationName, "organization", "o", "", "Your organization name")
	configurationCmd.PersistentFlags().StringVarP(&serviceName, "service", "s", "", "Your service name")
	configurationCmd.PersistentFlags().StringVarP(&namespaceName, "namespace", "n", "", "Your namespace name")
	configurationCmd.PersistentFlags().StringVarP(&configurationsData, "configurations", "c", "", "Your configurations")
	configurationCmd.PersistentFlags().IntVarP(&version, "versions", "v", 0, "Version")
	configurationCmd.PersistentFlags().StringVarP(&configKey, "key", "", "", "Configuration key")
	configurationCmd.PersistentFlags().StringVarP(&configVal, "value", "", "", "Configuration value")
}

//base func
func displayConfig() {
	if version < 1 {
		version = fetchLatestVersion(organizationName, serviceName, namespaceName)
	}

	config := fetchConfiguration(organizationName, serviceName, namespaceName, version)

	fmt.Println(displayHeader(serviceName, namespaceName, config.Version))
	fmt.Println(displayKeyVal(config.Configuration))
}

func unsetConfig() {
	version = fetchLatestVersion(organizationName, serviceName, namespaceName)
	config := fetchConfiguration(organizationName, serviceName, namespaceName, version)

	newConfig := config.Configuration
	if _, exists := newConfig[configKey]; !exists {
		fmt.Println("key not found!")
		return
	}

	delete(newConfig, configKey)

	data := make(map[string]interface{})
	data["configuration"] = newConfig
	dataJSON, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	updateConfig(dataJSON)
}

func setConfig() {
	version = fetchLatestVersion(organizationName, serviceName, namespaceName)
	config := fetchConfiguration(organizationName, serviceName, namespaceName, version)

	newConfig := config.Configuration

	newConfig[configKey] = configVal

	data := make(map[string]interface{})
	data["configuration"] = newConfig
	dataJSON, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	updateConfig(dataJSON)
}

func displayListVersions() {
	httphelper := &helper.HttpHelper{
		URL:           fmt.Sprintf("%v/%v/%v/%v/versions", "http://127.0.0.1:8080", organizationName, serviceName, namespaceName),
		Method:        "GET",
		Authorization: key,
	}

	data := make(map[string][]int)
	err := helper.FetchData(httphelper, &data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	output := []string{
		"Service|" + serviceName,
		"Namespace|" + namespaceName,
		"Versions|" + strings.Trim(strings.Replace(fmt.Sprint(data["versions"]), " ", ",", -1), "[]"),
	}
	result := columnize.SimpleFormat(output)
	fmt.Println(result)
}

func updateAllConfig() {
	version = fetchLatestVersion(organizationName, serviceName, namespaceName)
	config := fetchConfiguration(organizationName, serviceName, namespaceName, version)

	fmt.Println(displayHeader(serviceName, namespaceName, config.Version))
	fmt.Println(displayKeyVal(config.Configuration))

	newConfig := config.Configuration

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Do you want to edit a key (Y/N): ")
		editFlag, _ := reader.ReadString('\n')
		editFlag = string(editFlag[0 : len(editFlag)-1])
		if strings.EqualFold(editFlag, "N") {
			break
		}

		fmt.Print("Which key: ")
		key, _ := reader.ReadString('\n')
		key = string(key[0 : len(key)-1])
		if _, exists := newConfig[key]; exists {
			fmt.Print("New Value: ")
			val, _ := reader.ReadString('\n')
			val = string(val[0 : len(val)-1])
			newConfig[key] = val
		} else {
			fmt.Println("key not found!")
		}
	}

	for {
		fmt.Print("Do you want to delete a key (Y/N): ")
		deleteFlag, _ := reader.ReadString('\n')
		deleteFlag = string(deleteFlag[0 : len(deleteFlag)-1])
		if strings.EqualFold(deleteFlag, "n") {
			break
		}

		fmt.Print("Which key: ")
		key, _ := reader.ReadString('\n')
		key = string(key[0 : len(key)-1])
		if _, exists := newConfig[key]; exists {
			delete(newConfig, key)
		} else {
			fmt.Println("key not found!")
		}
	}

	for {
		fmt.Print("Do you want to add a key (Y/N): ")
		addFlag, _ := reader.ReadString('\n')
		addFlag = string(addFlag[0 : len(addFlag)-1])
		if strings.EqualFold(addFlag, "n") {
			break
		}

		var key string
		for {
			fmt.Print("Key: ")
			key, _ = reader.ReadString('\n')
			key = string(key[0 : len(key)-1])
			if key != "" {
				break
			}
		}

		fmt.Print("Value: ")
		val, _ := reader.ReadString('\n')
		val = string(val[0 : len(val)-1])
		newConfig[key] = val
	}

	fmt.Println("supposed to display new configuration")
	fmt.Println("supposed to ask edit again")

	fmt.Println("sending new data")

	fmt.Println(displayKeyVal(newConfig))
	data := make(map[string]interface{})
	data["configuration"] = newConfig
	dataJSON, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	updateConfig(dataJSON)
}

//ui helper
func displayKeyVal(config map[string]string) string {
	templateBody := []string{
		"#|Key|Value",
	}

	i := 1
	for k, v := range config {
		templateBody = append(templateBody, fmt.Sprintf("%v|%v|%v", i, k, v))
		i++
	}

	return columnize.SimpleFormat(templateBody)
}

func displayHeader(serviceName, namespaceName string, version int) string {
	templateHeader := []string{
		"Service|" + serviceName,
		"Namespace|" + namespaceName,
		"Version|" + strconv.Itoa(version),
	}

	return columnize.SimpleFormat(templateHeader)
}

//model helper
func fetchLatestVersion(organizationName, serviceName, namespaceName string) int {
	httphelper := &helper.HttpHelper{
		URL:           fmt.Sprintf("%v/%v/%v/%v/latest", "http://127.0.0.1:8080", organizationName, serviceName, namespaceName),
		Method:        "GET",
		Authorization: key,
	}

	data := make(map[string]int)
	err := helper.FetchData(httphelper, &data)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}

	return data["version"]
}

func fetchConfiguration(organizationName, serviceName, namespaceName string, version int) configView {
	httphelper := &helper.HttpHelper{
		URL:           fmt.Sprintf("%v/%v/%v/%v/%v", "http://127.0.0.1:8080", organizationName, serviceName, namespaceName, version),
		Method:        "GET",
		Authorization: key,
	}

	var config configView
	err := helper.FetchData(httphelper, &config)
	if err != nil {
		fmt.Println(err.Error())
		return config
	}
	return config
}

func updateConfig(payload []byte) {
	httphelper := &helper.HttpHelper{
		URL:           fmt.Sprintf("%v/%v/%v/%v", "http://127.0.0.1:8080", organizationName, serviceName, namespaceName),
		Method:        "POST",
		Authorization: key,
		Payload:       payload,
	}

	res, err := httphelper.SendRequest()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if res.StatusCode == http.StatusCreated {
		fmt.Println("New Configuration Created")
	} else {
		data := make(map[string]interface{})
		err = json.NewDecoder(res.Body).Decode(&data)
		res.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("HTTP Error " + fmt.Sprint(data["status"]) + ": " + fmt.Sprint(data["message"]))
	}
}
