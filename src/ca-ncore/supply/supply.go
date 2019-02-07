package supply

import (
	"io"
	"bufio"
	"bytes"
	"path/filepath"
	"github.com/cloudfoundry/libbuildpack"
	"encoding/json"
	"strings"
	"strconv"
	"fmt"
	"regexp"
	"os"
	"io/ioutil"
)

type Stager interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/stager.go
	BuildDir() string
	DepDir() string
	DepsIdx() string
	DepsDir() string
	ProfileDir() string
}

type Manifest interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/manifest.go
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Installer interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/installer.go
	FetchDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Command interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/command.go
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(dir string, program string, args ...string) (string, error)
}

type Supplier struct {
	Manifest  Manifest
	Installer Installer
	Stager    Stager
	Command   Command
	Log       *libbuildpack.Logger
}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying ca-ncore")

	if err := DownloadAgent(s); err != nil {
		return err
	}
	
	s.Log.Info("Downloaded agent")
	
	// Resolve the EM URL
	var agentManagerURL string
	credentials := GetIntroscopeCredentials(s)
	if credentials != nil {

		if(credentials["url"] != nil) {
			agentManagerURL = credentials["url"].(string)
		} else if(credentials["agent_manager_url"] != nil) {
			agentManagerURL = credentials["agent_manager_url"].(string)
		}
	}
	
	if agentManagerURL == "" {
		s.Log.Error("Failed to determine EM URL. Please bind the app to an Introscope service.")
		
		// Log the error but don't fail
		//return errors.New("Failed to determine EM URL")
	}
	
	// Update all properties in credentials
	for key, valueObj := range credentials {
		if (key == "url" || key == "agent_manager_url") {
			key = "agentManager.url.1"
		}
		
		s.Log.Info("Setting profile property %s", key)
		if err := UpdateAgentProperty(s, key, valueObj.(string)); err != nil {
			return err
		}
	}
	
	if IsWindows() {
		if err := UpdateAgentProperty(s, "introscope.agent.dotnet.monitorApplications", "hwc.exe"); err != nil {
			return err
		}
	}
	
	var appName string
	applicationProps := GetApplicationProperties(s)
	if applicationProps != nil {
		// Set APMENV_AGENT_NAME to this app name
		appName = applicationProps["application_name"].(string)
		s.Log.Info("Setting APMENV_AGENT_NAME to %s", appName)
		if err := UpdateAgentProperty(s, "introscope.agent.agentNameSystemPropertyKey", "APMENV_AGENT_NAME"); err != nil {
			return err
		}
	
		// set the hostname to the first application URI
		appURIs := applicationProps["application_uris"].([]interface{})
		if appURIs != nil {
			appHostName := appURIs[0].(string)
			s.Log.Info("Setting introscope.agent.hostName to %s", appHostName)
			if err := UpdateAgentProperty(s, "introscope.agent.hostName", appHostName); err != nil {
				return err
			}
		}
	}
	
	if err := WriteProfileScript(s, appName); err != nil {
		return err
	}
	
	s.Log.Info("Written profile")
	
	return nil
}

func DownloadAgent(s *Supplier) error {
	
	// Download the agent zip
	agentZip := filepath.Join(s.Stager.DepDir(), "apm.zip")
	
	depName := "apm-linux"
	if IsWindows() {
		depName = "apm-windows"
	}
	
	if err := s.Installer.FetchDependency(libbuildpack.Dependency{Name: depName, Version: "99.99.0"}, agentZip); err != nil {
		return err
	}

	if err := libbuildpack.ExtractZip(agentZip, filepath.Join(s.Stager.DepDir(), "../../","apm")); err != nil {
		return err
	}
	
	return nil
}

func WriteProfileScript(s *Supplier, appName string) error {
	
	// Write APM startup script to profile.d 
	if err := os.Mkdir(filepath.Join(s.Stager.DepDir(), "profile.d"), 0777); err != nil {
		return err
	}
	
	if IsLinux() {
		apmScriptPath := filepath.Join(s.Stager.DepDir(), "profile.d/apm.sh")
		if err := ioutil.WriteFile(apmScriptPath, []byte(`
			export CORECLR_ENABLE_PROFILING=1
			export CORECLR_PROFILER={5F048FC6-251C-4684-8CCA-76047B02AC98}
			export CORECLR_PROFILER_PATH=/home/vcap/apm/wily/bin/wily.NativeProfiler.so
			export APMENV_AGENT_PROFILE=/home/vcap/apm/wily/IntroscopeAgent.profile
			`), 0666); err != nil {
			return err
		}
		
		// If the app name is set append the export
		if appName != "" {
			// Append the new key/value
			fileHandle, err := os.OpenFile(apmScriptPath, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			defer fileHandle.Close()
			writer := bufio.NewWriter(fileHandle)

			fmt.Fprintln(writer, fmt.Sprintf("\nexport APMENV_AGENT_NAME=%s", appName))
			writer.Flush()
		}
	}
	
	if IsWindows() {
		apmScriptPath := filepath.Join(s.Stager.DepDir(), "profile.d/apm.bat")
		if err := ioutil.WriteFile(apmScriptPath, []byte(`
			set COR_ENABLE_PROFILING=1
			set COR_PROFILER={5F048FC6-251C-4684-8CCA-76047B02AC98}
			set COR_PROFILER_PATH=C:\Users\vcap\apm\wily\bin\wily.NativeProfiler.dll
			set CORECLR_ENABLE_PROFILING=1
			set CORECLR_PROFILER={5F048FC6-251C-4684-8CCA-76047B02AC98}
			set CORECLR_PROFILER_PATH=C:\Users\vcap\apm\wily\bin\wily.NativeProfiler.dll
			set com.wily.introscope.agentProfile=C:\Users\vcap\apm\wily\IntroscopeAgent.profile
			`), 0666); err != nil {
			return err
		}
		
		// If the app name is set append the export
		if appName != "" {
			// Append the new key/value
			fileHandle, err := os.OpenFile(apmScriptPath, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			defer fileHandle.Close()
			writer := bufio.NewWriter(fileHandle)

			fmt.Fprintln(writer, fmt.Sprintf("\nset APMENV_AGENT_NAME=%s", appName))
			writer.Flush()
		}
		
		// Support binary buildpack 
		// deps/profile.d scripts are not copied. Add script to app/.profile.d
		dest := filepath.Join(s.Stager.ProfileDir(), s.Stager.DepsIdx()+"_apm.bat")
		s.Log.Info("Copying apm.bat to %s", dest)

		if err := libbuildpack.CopyFile(apmScriptPath, dest); err != nil {
			return err
		}
	}
	
	return nil
}

func GetIntroscopeCredentials(s *Supplier) map[string]interface{} {
	// Parse Services
	var services map[string]interface{}
	serviceBytes := []byte(os.Getenv("VCAP_SERVICES"))
	if err := json.Unmarshal(serviceBytes, &services); err != nil {
		return nil
	}
	
	for _, serviceArrayObj  := range services {
		serviceArray := serviceArrayObj.([]interface{})
		for _, serviceObj := range serviceArray {
			service := serviceObj.(map[string]interface{})
			serviceName := service["name"].(string)
			
			// Match an introscope service name
			if strings.EqualFold(serviceName, "introscope") {
				emCredentials := service["credentials"].(map[string]interface{})
				
				return emCredentials
			}
		}
	}
	
	return nil
}

func GetApplicationProperties(s *Supplier) map[string]interface{} {
	// Parse Application JSON
	var appProps map[string]interface{}
	appJSONBytes := []byte(os.Getenv("VCAP_APPLICATION"))
	if err := json.Unmarshal(appJSONBytes, &appProps); err != nil {
		return nil
	}

	return appProps
}

func UpdateAgentProperty(s *Supplier, key string, value string) error {
	profilePath := filepath.Join(s.Stager.DepDir(), "../../apm/wily/IntroscopeAgent.profile")
	
	updated := false
	
	if IsLinux() {
		// Check if the key exists
		var grepBuff bytes.Buffer
		grepWriter := bufio.NewWriter(&grepBuff)
		_ = s.Command.Execute(s.Stager.DepDir(), grepWriter, os.Stderr, 
			"/bin/grep", "-c", fmt.Sprintf("^%s=", key), profilePath)
		
		keyCount, err := strconv.Atoi(strings.TrimSpace(grepBuff.String()))
		if err != nil {
			s.Log.Error("grep failed: %s", grepBuff.String())
			return err
		}
		
		//s.Log.Info("Count for %s = %d", key, keyCount)
		if keyCount > 0 {
			// Replace the existing value
			
			// Create a copy of the current profile
			tempProfilePath := filepath.Join(s.Stager.DepDir(), "temp_IntroscopeAgent.profile")
			
			if err := libbuildpack.CopyFile(profilePath, tempProfilePath); err != nil {
				return err
			}
		
			// Create a buffered writer for the profile output
			profileFile, err := os.Create(profilePath)
			if err != nil {
				return err
			}
		
			profileWriter := bufio.NewWriter(profileFile)
		
			// Replace the value
			escKey := strings.Replace(regexp.QuoteMeta(key), "/", "\\/", -1)
			escValue := strings.Replace(regexp.QuoteMeta(value), "/", "\\/", -1)
			s.Log.Debug("sed expr: %s", fmt.Sprintf("s/^%s=.*/%s=%s/", escKey, escKey, escValue))
			if err := s.Command.Execute(s.Stager.DepDir(), profileWriter, os.Stderr, 
				"/bin/sed", fmt.Sprintf("s/^%s=.*/%s=%s/", escKey, escKey, escValue), tempProfilePath); err != nil {
				return err
			}
		
			profileWriter.Flush()
			
			updated = true
		} 
	}
	
	if !updated {
		// Append the new key/value
		fileHandle, err := os.OpenFile(profilePath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer fileHandle.Close()
		writer := bufio.NewWriter(fileHandle)
		

		fmt.Fprintln(writer, fmt.Sprintf("\n%s=%s", key, value))
		writer.Flush()
	}
	
	return nil
}

func IsWindows() bool {
	cfStack := os.Getenv("CF_STACK")
	if strings.Contains(cfStack, "windows") {
		return true
	}
	
	return false
}

func IsLinux() bool {
	cfStack := os.Getenv("CF_STACK")
	if strings.Contains(cfStack, "cflinux") {
		return true
	}
	
	return false
}
