package cmd

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zricethezav/gitleaks/v8/ucmp"
	"os"
	"path"
	"runtime"
)

const BINARY_INSTALL_PATH_KEY = "install-path"

func init() {
	enableCmd.Flags().String(string(ucmp.AUDIT_CONFIG_KEY_URL), "https://audit.ucmp.uplus.co.kr/gitleaks", "Audit Backend Url (Default : https://audit.ucmp.uplus.co.kr/gitleaks/)")
	enableCmd.Flags().String(BINARY_INSTALL_PATH_KEY, "", "Gitleaks Binary Install Path (Default : /usr/local/bin/ for linux, C:/Windows/ for windows")
	enableCmd.Flags().Bool(string(ucmp.AUDIT_CONFIG_KEY_DEBUG), false, "Enable debug output")
	enableCmd.Flags().Int64(string(ucmp.AUDIT_CONFIG_KEY_TIMEOUT), 0, "Audit Backend Timeout (Default : 5 Seconds")
	// enableCmd.MarkFlagRequired(string(ucmp.AUDIT_CONFIG_KEY_URL))
	rootCmd.AddCommand(enableCmd)
}

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable gitleaks in pre-commit script",
	Run:   runEnable,
}

func runEnable(cmd *cobra.Command, args []string) {
	auditConfig := ucmp.GetAuditConfigInstance()

	// 1. Enable Global Git Hooks (pre-commit, post-commit)
	err := auditConfig.SetGlobalHooksPath()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to set global hooks path")
		os.Exit(-1)
	}

	// 2. Setting Url and other flags (debug, enable)
	url, _ := cmd.Flags().GetString(string(ucmp.AUDIT_CONFIG_KEY_URL))
	auditConfig.SetAuditConfig(ucmp.GIT_SCOPE_GLOBAL, ucmp.AUDIT_CONFIG_KEY_URL, url) // Check Global Git Config

	debug, _ := cmd.Flags().GetBool("debug")
	if debug {
		// If enable command with --debug flag, print the all commands logs
		auditConfig.SetAuditConfig(ucmp.GIT_SCOPE_LOCAL, ucmp.AUDIT_CONFIG_KEY_DEBUG, debug) // Check Local Git Config
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	timeout, _ := cmd.Flags().GetInt64(string(ucmp.AUDIT_CONFIG_KEY_TIMEOUT))
	if timeout > 0 {
		auditConfig.SetAuditConfig(ucmp.GIT_SCOPE_GLOBAL, ucmp.AUDIT_CONFIG_KEY_TIMEOUT, timeout) // Check Global Git Config
	}

	auditConfig.SetAuditConfig(ucmp.GIT_SCOPE_GLOBAL, ucmp.AUDIT_CONFIG_KEY_ENABLE, true) // Check Global Git Config

	// Insert Script Content in $HOME/.githooks pre-commit, post-commit

	// 기 존재하는 스크립트 삭제하여 Desired Script 로 설정 될 수 있도록 함
	ucmp.RemoveGitHookScript(ucmp.PreCommitScriptPath)
	ucmp.RemoveGitHookScript(ucmp.PostCommitScriptPath)

	// If Specified Install Path, Ensure the Path Exists
	binaryInstallPath, _ := cmd.Flags().GetString(BINARY_INSTALL_PATH_KEY)
	if len(binaryInstallPath) != 0 {
		ucmp.EnsurePathDirectory(binaryInstallPath)
	}

	// 3. Install Global Git Hooks
	ucmp.InstallGitHookScript(ucmp.PreCommitScriptPath, ucmp.LocalPreCommitSupportScript)

	if len(binaryInstallPath) != 0 {
		ucmp.InstallGitHookScript(ucmp.PreCommitScriptPath, path.Join(binaryInstallPath, ucmp.PreCommitScript))
		ucmp.InstallGitHookScript(ucmp.PostCommitScriptPath, path.Join(binaryInstallPath, ucmp.PostCommitScript))
	} else {
		switch runtime.GOOS {
		case "windows":
			ucmp.InstallGitHookScript(ucmp.PreCommitScriptPath, path.Join(ucmp.WindowsBinaryInstallPath, ucmp.PreCommitScript))
			ucmp.InstallGitHookScript(ucmp.PostCommitScriptPath, path.Join(ucmp.WindowsBinaryInstallPath, ucmp.PostCommitScript))
		default:
			ucmp.InstallGitHookScript(ucmp.PreCommitScriptPath, path.Join(ucmp.LinuxBinaryInstallPath, ucmp.PreCommitScript))
			ucmp.InstallGitHookScript(ucmp.PostCommitScriptPath, path.Join(ucmp.LinuxBinaryInstallPath, ucmp.PostCommitScript))
		}
	}

	log.Info().Msg("Gitleaks Enabled")
}
