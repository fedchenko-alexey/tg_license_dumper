// Fedchenko Alexey (fedchenko.alexey@r7-office.ru) R7-Office. 2024

package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/cavaliergopher/grab/v3"
)

// ConfigData contains config data. Api keys, paths
type ConfigData struct {
	botAPIKey string
}

type FileProperties struct {
	FileId     string `json:"file_id"`
	FileUniqId string `json:"file_unique_id"`
	FileSize   int    `json:"file_size"`
	FilePath   string `json:"file_path"`
}

type GetFileResponse struct {
	Ok         bool           `json:"ok"`
	Properties FileProperties `json:"result"`
}

func loadConfigFromEnv() ConfigData {

	botAPIKey, isSet := os.LookupEnv("TELEGRAM_BOT_API_KEY_LICENSE")
	if !isSet {
		log.Panic("No bot api key found. Please set TELEGRAM_BOT_API_KEY_LICENSE env")
	}

	log.Println(botAPIKey + " ")

	return ConfigData{botAPIKey}
}

func getLicenseDump(fileId string) string {
	filePath := getFilePath(fileId)
	if len(filePath) <= 0 {
		return "Ошибка получения файла."

	}

	fileSystemPath := downloadFile(filePath)
	if len(fileSystemPath) <= 0 {
		return "Ошибка скачивания файла."
	}

	dump := getDumpString(fileSystemPath)
	removeLocalFile(fileSystemPath)

	return dump
}

func removeLocalFile(fileSystemPath string) {
	err := os.Remove(fileSystemPath)
	if err != nil {
		log.Println("Error delete file", fileSystemPath, err)
	}
}

func fileIsLicense(fileSize int, fileName string) bool {
	if fileSize > 500 {
		return true
	}

	if strings.HasSuffix(fileName, "lickey") {
		return true
	}

	return false
}

func getFilePath(fileId string) string {
	builder := strings.Builder{}
	builder.WriteString("https://api.telegram.org/bot")
	builder.WriteString(SystemConfig.botAPIKey)
	builder.WriteString("/getFile?file_id=")
	builder.WriteString(fileId)

	resp, err := http.Get(builder.String())
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var js GetFileResponse
	err = json.Unmarshal(body, &js)
	if err != nil {
		return ""
	}

	return js.Properties.FilePath
}

func downloadFile(filePath string) string {
	builder := strings.Builder{}
	builder.WriteString("https://api.telegram.org/file/bot")
	builder.WriteString(SystemConfig.botAPIKey)
	builder.WriteString("/")
	builder.WriteString(filePath)

	resp, err := grab.Get("/tmp", builder.String())
	if err != nil {
		return ""
	}

	return resp.Filename
}

func getDumpString(fileSystemPath string) string {
	out, err := exec.Command("./dumper", fileSystemPath).Output()
	if err != nil {
		return "Ошибка дампа файла."
	}

	re := regexp.MustCompile(`License info:`)
	split := re.Split(string(out), -1)

	return split[len(split)-1]
}
