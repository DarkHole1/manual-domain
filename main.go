package main

import (
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

type FT struct {
	IPs     string
	Message string
	Host    string
	Zone    string
}

func ReadCurrentIPs(conf Config) string {
	data, err := os.ReadFile(conf.File)
	if err != nil {
		log.Printf("Error occured: %s\n", err)
		return ""
	}
	res := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, fmt.Sprintf("%s\tIN\tA\t", conf.Host)) {
			res = append(res, strings.TrimPrefix(line, fmt.Sprintf("%s\tIN\tA\t", conf.Host)))
		}
		if strings.HasPrefix(line, fmt.Sprintf("%s\tIN\tAAAA\t", conf.Host)) {
			res = append(res, strings.TrimPrefix(line, fmt.Sprintf("%s\tIN\tAAAA\t", conf.Host)))
		}
	}
	return strings.Join(res, "\n")
}

func main() {

	conf := ReadConfigDefault().UnwrapMessage("Error reading config: %s")
	page := template.Must(template.ParseFiles("./index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		currentIps := ReadCurrentIPs(conf)
		if r.Method == "POST" {
			err := r.ParseForm()
			if err != nil {
				log.Printf("Error occured: %s\n", err.Error())
				page.Execute(w, FT{currentIps, "Случилась непонятная внутренняя ошибка", conf.Host, conf.Domain})
				return
			}

			ips := r.PostForm.Get("ips")
			if ips == "" {
				page.Execute(w, FT{currentIps, "Запрос пустой или некорректный", conf.Host, conf.Domain})
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
					page.Execute(w, FT{currentIps, "Я не смог разобрать IP \"" + ipCleaned + "\"", conf.Host, conf.Domain})
					return
				}
				ipsParsed = append(ipsParsed, ipParsed)
			}

			res := ""
			for _, ip := range ipsParsed {
				if ip.To4() != nil {
					res += fmt.Sprintf("%s\tIN\tA\t%s\n", conf.Host, ip.String())
				} else {
					res += fmt.Sprintf("%s\tIN\tAAAA\t%s\n", conf.Host, ip.String())
				}
			}

			data, err := os.ReadFile(conf.File)
			if err != nil {
				log.Printf("Error occured: " + err.Error())
				page.Execute(w, FT{currentIps, "Случилась непонятная внутренняя ошибка", conf.Host, conf.Domain})
				return
			}
			err = os.WriteFile(path.Join(conf.BackupDir, path.Base(conf.File)+"."+time.Now().Format("2006-01-02-15-04-05")), data, 0644)
			if err != nil {
				log.Printf("Error occured: %s\n", err.Error())
				page.Execute(w, FT{currentIps, "Случилась ошибка при создании бэкапа", conf.Host, conf.Domain})
			}
			newLines := []string{}
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, fmt.Sprintf("%s\tIN\tA\t", conf.Host)) || strings.HasPrefix(line, fmt.Sprintf("%s\tIN\tAAAA\t", conf.Host)) {
					continue
				}
				newLines = append(newLines, line)
			}
			err = os.WriteFile(conf.File, []byte(strings.Join(newLines, "\n")+res), 0644)
			if err != nil {
				log.Printf("Error occured: " + err.Error())
				page.Execute(w, FT{currentIps, "Случилась непонятная внутренняя ошибка", conf.Host, conf.Domain})
				return
			}
			currentIps = ""
			for _, ip := range ipsParsed {
				currentIps += ip.String() + "\n"
			}
			cmd := exec.Command("sh", "-c", conf.Command)
			if out, err := cmd.CombinedOutput(); err != nil {
				log.Printf("Error occured: %s", err)
				log.Printf("Output: %s", string(out))
				page.Execute(w, FT{currentIps, "Не получилось перезапустить DNS, смотрите в логи", conf.Host, conf.Domain})
				return
			}
			page.Execute(w, FT{currentIps, "Всё прошло хорошо", conf.Host, conf.Domain})
			return
		}

		page.Execute(w, FT{currentIps, "", conf.Host, conf.Domain})
	})

	ListenMultiplyFatal(conf.Addresses).Unwrap()
}
