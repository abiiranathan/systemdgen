package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// SystemdUnit represents the configuration for a systemd unit
type SystemdUnit struct {
	ServiceName string
	Description string
	ExecStart   string
	User        string
	Group       string
	WorkingDir  string
}

// GenerateUnitFile generates the content for a systemd unit file
func GenerateUnitFile(unit SystemdUnit) string {
	const unitTemplate = `[Unit]
Description={{.Description}}

[Service]
ExecStart={{ .ExecStart }}
Restart=always
User={{ .User }}
Group={{ .Group }}
WorkingDirectory={{ .WorkingDir }}

[Install]
WantedBy=multi-user.target
`

	tmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		fmt.Println("Error creating template:", err)
		os.Exit(1)
	}

	var result = new(bytes.Buffer)
	err = tmpl.Execute(result, unit)
	if err != nil {
		fmt.Println("Error executing template:", err)
		os.Exit(1)
	}
	return result.String()
}

// InstallUnitFile installs the unit file to the systemd directory
func InstallUnitFile(filename string) {
	// copy the unit file to the systemd directory
	cmd := exec.Command("sudo", "cp", filename, "/etc/systemd/system/")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error installing unit file:", err)
		os.Exit(1)
	}

	cmd = exec.Command("sudo", "systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error reloading systemd daemon:", err)
		os.Exit(1)
	}

	fmt.Println("Systemd unit file installed and daemon reloaded.")
}

func required(value string, message string) {
	if value == "" {
		fmt.Println(message)
		os.Exit(1)
	}
}

func main() {
	serviceName := flag.String("name", "", "Name of the systemd service")
	description := flag.String("description", "", "Description of the systemd service")
	execStart := flag.String("exec", "", "Command to start the service")
	user := flag.String("user", "root", "User to run the service as")
	group := flag.String("group", "root", "Group to run the service as")
	workDir := flag.String("workdir", "/", "Working directory for the service")
	install := flag.Bool("install", false, "Install the unit file")
	enable := flag.Bool("enable", false, "Enable service at boot")

	flag.Parse()

	required(*serviceName, "Service name is required")
	required(*description, "Description is required")
	required(*execStart, "Exec command is required")
	required(*user, "User is required")
	required(*group, "Group is required")
	required(*workDir, "Working directory is required")

	// verify that the execStart command is valid
	executable := strings.Split(*execStart, " ")[0]
	_, err := exec.LookPath(executable)
	if err != nil {
		fmt.Printf("Error finding executable: %s\n", executable)
		os.Exit(1)
	}

	unit := SystemdUnit{
		ServiceName: *serviceName,
		Description: *description,
		ExecStart:   *execStart,
		User:        *user,
		Group:       *group,
		WorkingDir:  *workDir,
	}

	unitFileContent := GenerateUnitFile(unit)
	unitFileName := fmt.Sprintf("/tmp/%s.service", *serviceName)

	err = os.WriteFile(unitFileName, []byte(unitFileContent), 0644)
	if err != nil {
		fmt.Println("Error writing unit file:", err)
		os.Exit(1)
	}

	fmt.Printf("Systemd unit file generated at: %s\n", unitFileName)

	if *install {
		InstallUnitFile(unitFileName)

		if *enable {
			EnableUnitFile(unitFileName)
		}
	}
}

func EnableUnitFile(unitFileName string) {
	cmd := exec.Command("sudo", "systemctl", "enable", unitFileName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		os.Exit(1)
	}

	fmt.Println(string(out))
	fmt.Println("Systemd unit file enabled.")

	// start the service
	cmd = exec.Command("sudo", "systemctl", "start", unitFileName)
	out, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		os.Exit(1)
	}

	fmt.Println(string(out))
	fmt.Println("Systemd unit file started.")
}
