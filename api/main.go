package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// BoxInfo represents the input data from the user.
type BoxInfo struct {
	Name        string `json:"name"`
	Hostname    string `json:"hostname"`
	CPU         int    `json:"cpu"`
	Memory      int    `json:"memory"`
	IPAddress   string `json:"ip_address"`
	NetworkType string `json:"network_type"`
}

// VagrantfileTemplate is the template for generating a Vagrantfile.
const VagrantfileTemplate = `
Vagrant.configure("2") do |config|
  config.vm.box = "{{.Name}}"
  config.vm.hostname = "{{.Hostname}}"
  config.vm.network "{{.NetworkType}}", ip: "{{.IPAddress}}"
  config.vm.provider "virtualbox" do |vb|
    vb.name = "{{.Hostname}}"
    vb.memory = {{.Memory}}
    vb.cpus = {{.CPU}}
  end
end
`

func main() {
	http.HandleFunc("/generate", generateHandler)
	http.ListenAndServe(":8090", nil)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var boxInfo BoxInfo
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&boxInfo)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	vagrantfileContent, err := generateVagrantfile(boxInfo)
	if err != nil {
		http.Error(w, "Failed to generate Vagrantfile", http.StatusInternalServerError)
		return
	}

	err = writeToFile("Vagrantfile", vagrantfileContent)
	if err != nil {
		http.Error(w, "Failed to write Vagrantfile", http.StatusInternalServerError)
		return
	}

	cmd := exec.Command("vagrant", "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		http.Error(w, "Failed to start provisioning", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Provisioning started successfully")
}

func generateVagrantfile(info BoxInfo) (string, error) {
	tmpl, err := template.New("Vagrantfile").Parse(VagrantfileTemplate)
	if err != nil {
		return "", err
	}

	var result string
	writer := &strings.Builder{}
	err = tmpl.Execute(writer, info)
	if err != nil {
		return "", err
	}
	result = writer.String()

	return result, nil
}

func writeToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}
