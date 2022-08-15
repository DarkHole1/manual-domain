package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type Configuration struct {
	File      string
	Domain    string
	Command   string
	BackupDir string
}

func ReadConfig() Configuration {
	file, _ := os.Open("conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal("Error reading config: ", err.Error())
	}
	return configuration
}

func main() {

	conf := ReadConfig()
	page := template.Must(template.ParseFiles("./index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			err := r.ParseForm()
			if err != nil {
				log.Printf("Error occured: %s\n", err.Error())
				page.Execute(w, "Случилась непонятная внутренняя ошибка")
				return
			}

			ips := r.PostForm.Get("ips")
			if ips == "" {
				page.Execute(w, "Запрос пустой или некорректный")
				return
			}
			ipsSplitted := strings.Split(ips, "\n")
			ipsParsed := make([]net.IP, 0)
			for _, ip := range ipsSplitted {
				ipCleaned := strings.Trim(ip, " \r\n")
				if ipCleaned == "" {
					continue
				}
				ipParsed := net.ParseIP(ipCleaned)
				if ipParsed == nil {
					page.Execute(w, "Я не смог разобрать IP \""+ipCleaned+"\"")
					return
				}
				ipsParsed = append(ipsParsed, ipParsed)
			}

			res := ""
			for _, ip := range ipsParsed {
				if ip.To4() != nil {
					res += fmt.Sprintf("%s\tIN\tA\t%s\n", conf.Domain, ip.String())
				} else {
					res += fmt.Sprintf("%s\tIN\tAAAA\t%s\n", conf.Domain, ip.String())
				}
			}

			data, err := os.ReadFile(conf.File)
			if err != nil {
				log.Printf("Error occured: " + err.Error())
				page.Execute(w, "Случилась непонятная внутренняя ошибка")
				return
			}
			err = os.WriteFile(path.Join(conf.BackupDir, path.Base(conf.File)+"."+time.Now().Format("2006-01-02-15-04-05")), data, 0644)
			if err != nil {
				log.Printf("Error occured: %s\n", err.Error())
				page.Execute(w, "Случилась ошибка при создании бэкапа")
			}
			newLines := []string{}
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, fmt.Sprintf("%s\tIN\tA\t", conf.Domain)) || strings.HasPrefix(line, fmt.Sprintf("%s\tIN\tAAAA\t", conf.Domain)) {
					continue
				}
				newLines = append(newLines, line)
			}
			err = os.WriteFile(conf.File, []byte(strings.Join(newLines, "\n")+res), 0644)
			if err != nil {
				log.Printf("Error occured: " + err.Error())
				page.Execute(w, "Случилась непонятная внутренняя ошибка")
				return
			}
			cmd := exec.Command("sh", "-c", conf.Command)
			if out, err := cmd.CombinedOutput(); err != nil {
				log.Printf("Error occured: %s", err)
				log.Printf("Output: %s", string(out))
				page.Execute(w, "Не получилось перезапустить DNS, смотрите в логи")
				return
			}
			page.Execute(w, "Всё прошло хорошо")
			return
		}

		page.Execute(w, "")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
