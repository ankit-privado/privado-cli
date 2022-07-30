package cmd

import (
	"fmt"
	"time"

	"github.com/Privado-Inc/privado/pkg/config"
	"github.com/Privado-Inc/privado/pkg/docker"
	"github.com/Privado-Inc/privado/pkg/fileutils"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan <repository>",
	Short: "Scan a codebase or repository to identify privacy issues and generate compliance reports",
	Args:  cobra.ExactArgs(1),
	Run:   scan,
}

func defineScanFlags(cmd *cobra.Command) {
	scanCmd.Flags().StringP("rules", "r", "", "Specifies the rule directory to be passed to privado-core for scanning. These external rules are merged with the default set of rules that Privado defines")

	scanCmd.Flags().BoolP("ignore-default-rules", "i", false, "If specified, the default rules are ignored and only the specified rules (-r) are considered")

	scanCmd.Flags().BoolP("overwrite", "o", false, "If specified, the warning prompt for existing scan results is disabled and any existing results are overwritten")
	scanCmd.Flags().Bool("debug", false, "Enables privado-core image output in debug mode")
	scanCmd.Flags().MarkHidden("debug")
}

func scan(cmd *cobra.Command, args []string) {
	repository := args[0]
	debug, _ := cmd.Flags().GetBool("debug")
	// overwriteResults, _ := cmd.Flags().GetBool("overwrite")

	externalRules, _ := cmd.Flags().GetString("rules")
	if externalRules != "" {
		externalRules = fileutils.GetAbsolutePath(externalRules)
		externalRulesExists, _ := fileutils.DoesFileExists(externalRules)
		if !externalRulesExists {
			exit(fmt.Sprintf("Could not validate the rules directory: %s", externalRules), true)
		}
	}

	ignoreDefaultRules, _ := cmd.Flags().GetBool("ignore-default-rules")
	if ignoreDefaultRules && externalRules == "" {
		exit(fmt.Sprint(
			"Default rules cannot be ignored without any external rules.\n",
			"You can specify your own rules using the `-r` option.\n\n",
			"For more info, run: 'privado help'\n",
		), true)
	}

	hasUpdate, updateMessage, err := checkForUpdate()
	if err == nil && hasUpdate {
		fmt.Println(updateMessage)
		time.Sleep(config.AppConfig.SlowdownTime)
		fmt.Println("To use the latest version of Privado CLI, run `privado update`")
		time.Sleep(config.AppConfig.SlowdownTime)
		fmt.Println()
	}

	// if overwrite flag is not specified, check for existing results
	// if !overwriteResults {
	// 	resultsPath := filepath.Join(utils.GetAbsolutePath(repository), config.AppConfig.PrivacyResultsPathSuffix)
	// 	if exists, _ := utils.DoesFileExists(resultsPath); exists {
	// 		fmt.Printf("> Scan report already exists (%s)\n", config.AppConfig.PrivacyResultsPathSuffix)
	// 		// fmt.Println("> If you want to view or edit existing results, run 'privado load' instead")

	// 		fmt.Println("\n> Rescan will overwrite existing results and progress")
	// 		confirm, _ := utils.ShowConfirmationPrompt("Continue?")
	// 		if !confirm {
	// 			exit("Terminating..", false)
	// 		}
	// 		fmt.Println()
	// 	}
	// }

	// "pass -ir even when internal rules are ignored (-i)"
	commandArgs := []string{config.AppConfig.Container.SourceCodeVolumeDir, "-ir", config.AppConfig.Container.InternalRulesVolumeDir}

	fmt.Println("> Scanning directory:", fileutils.GetAbsolutePath(repository))
	// run image with options
	err = docker.RunImage(
		docker.OptionWithArgs(commandArgs),
		docker.OptionWithAttachedOutput(),
		docker.OptionWithSourceVolume(fileutils.GetAbsolutePath(repository)),
		docker.OptionWithUserConfigVolume(config.AppConfig.UserConfigurationFilePath),
		docker.OptionWithUserKeyVolume(config.AppConfig.UserKeyPath),
		docker.OptionWithIgnoreDefaultRules(ignoreDefaultRules),
		docker.OptionWithExternalRulesVolume(externalRules),
		docker.OptionWithDebug(debug),
	)
	if err != nil {
		exit(fmt.Sprintf("Received error: %s", err), true)
	}
}

func init() {
	defineScanFlags(scanCmd)
	rootCmd.AddCommand(scanCmd)
}
