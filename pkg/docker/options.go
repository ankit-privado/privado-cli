package docker

import (
	"fmt"
	"os"

	"github.com/Privado-Inc/privado-cli/pkg/config"
)

type containerVolumes struct {
	userKeyVolumeEnabled, dockerKeyVolumeEnabled, sourceCodeVolumeEnabled,
	externalRulesVolumeEnabled, userConfigVolumeEnabled, packageCacheVolumeEnabled bool
	userKeyVolumeHost, dockerKeyVolumeHost, sourceCodeVolumeHost,
	externalRulesVolumeHost, userConfigVolumeHost, packageCacheVolumeHost string
}

type EnvVar struct {
	Key, Value string
}

type RunImageOption func(opts *runImageHandler)

type runImageHandler struct {
	pullLatestImage                     bool
	args                                []string
	volumes                             containerVolumes
	environmentVars                     []string
	setupInterrupt                      bool
	attachOutput                        bool
	spawnWebBrowserOnURLMessage         bool
	spawnWebBrowserOnURLTriggerMessages []string
	exitOnError                         bool
	exitOnErrorTriggerMessages          []string
}

func newRunImageHandler(opts []RunImageOption) runImageHandler {
	// defaults here
	rh := runImageHandler{}
	for _, opt := range opts {
		opt(&rh)
	}
	return rh
}

// Prepend option functions with "Option"

func OptionWithLatestImage(pullImage bool) RunImageOption {
	return func(rh *runImageHandler) {
		rh.pullLatestImage = pullImage
	}
}

func OptionWithArgs(args []string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.args = args
	}
}

func OptionWithUserKeyVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.volumes.userKeyVolumeEnabled = true
		rh.volumes.userKeyVolumeHost = volumeHost
	}
}

func OptionWithDockerKeyVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.volumes.dockerKeyVolumeEnabled = true
		rh.volumes.dockerKeyVolumeHost = volumeHost
	}
}

func OptionWithUserConfigVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.volumes.userConfigVolumeEnabled = true
		rh.volumes.userConfigVolumeHost = volumeHost
	}
}

func OptionWithSourceVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.volumes.sourceCodeVolumeEnabled = true
		rh.volumes.sourceCodeVolumeHost = volumeHost
	}
}

func OptionWithExternalRulesVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		if volumeHost != "" {
			rh.volumes.externalRulesVolumeEnabled = true
			rh.volumes.externalRulesVolumeHost = volumeHost
			rh.args = append(rh.args, "-er", config.AppConfig.Container.ExternalRulesVolumeDir)
		}
	}
}

// eventually, volumes for all packages for all languages will come here
// unless another approach for cache is decided. Therefore, suggest to not
// make any specific changes related to M2 package volume cache
func OptionWithPackageCacheVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		if err := os.MkdirAll(volumeHost, os.ModePerm); err == nil {
			rh.volumes.packageCacheVolumeEnabled = true
			rh.volumes.packageCacheVolumeHost = volumeHost
		} else {
			fmt.Println("[WARN]: Could not create package cache volume on host. skipping volume mount: ", err)
		}
	}
}

func OptionWithIgnoreDefaultRules(ignoreDefaultRules bool) RunImageOption {
	return func(rh *runImageHandler) {
		if ignoreDefaultRules {
			rh.args = append(rh.args, "-i")
		}
	}
}

func OptionWithSkipDependencyDownload(skipDependencyDownload bool) RunImageOption {
	return func(rh *runImageHandler) {
		if skipDependencyDownload {
			rh.args = append(rh.args, "-sdd")
		}
	}
}

func OptionWithEnvironmentVariables(envVars []EnvVar) RunImageOption {
	return func(rh *runImageHandler) {
		if len(envVars) > 0 {
			processedEnvStrings := []string{}
			for _, envVar := range envVars {
				if envVar.Key != "" {
					processedEnvStrings = append(processedEnvStrings, fmt.Sprintf("%s=%s", envVar.Key, envVar.Value))
				}
			}
			rh.environmentVars = processedEnvStrings
		}
	}
}

func OptionWithInterrupt() RunImageOption {
	return func(rh *runImageHandler) {
		rh.setupInterrupt = true
	}
}

func OptionWithAttachedOutput() RunImageOption {
	return func(rh *runImageHandler) {
		rh.attachOutput = true
	}
}

// listens for these message (we use strings.Contains)
// and spawns a browser with url in the message
// the messagePrefix must contain a URL for autospawn
// or this is silently ignored
func OptionWithAutoSpawnBrowserOnURLMessages(messages []string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.spawnWebBrowserOnURLMessage = true
		rh.spawnWebBrowserOnURLTriggerMessages = messages
	}
}

func OptionWithExitErrorMessages(messages []string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.exitOnError = true
		rh.exitOnErrorTriggerMessages = messages
	}
}

func OptionWithDebug(isDebug bool) RunImageOption {
	return func(rh *runImageHandler) {
		// currently only enable output in debug mode
		if isDebug {
			rh.attachOutput = true
			rh.args = append(rh.args, "--debug")
		}
	}
}
