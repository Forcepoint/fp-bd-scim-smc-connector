package cmd

import (
	"fmt"
	"github.cicd.cloud.fpdev.io/BD/fp-smc-golang/src/smc"
	"github.cicd.cloud.fpdev.io/BD/scim-smc-connector/lib"
	"github.com/fsnotify/fsnotify"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

var (
	cfgFile     string
	SmcInstance smc.Smc
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "smcconnector",
	Short: "SMC Connector",
	Long:  `connect SMC with SCIM`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "the config file)")
	//if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
	//	log.Fatal(err.Error())
	//}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	//set the default values
	viper.SetDefault("issuer", "ForcePoint")
	viper.SetDefault("ROLES_UPDATE_TIME_IN_MINUTES", 3)
	viper.SetDefault("ROLES.PERMISSIONS.VIEWER", true)
	viper.SetDefault("ROLES.PERMISSIONS.LOGS_VIEWER", false)
	viper.SetDefault("ROLES.PERMISSIONS.REPORTS_MANAGER", false)
	viper.SetDefault("ROLES.PERMISSIONS.OWNER", false)
	viper.SetDefault("ROLES.PERMISSIONS.OPERATOR", false)
	viper.SetDefault("ROLES.PERMISSIONS.MONITOR", false)
	viper.SetDefault("ROLES.PERMISSIONS.EDITOR", false)
	viper.SetDefault("ROLES.PERMISSIONS.NSX_ROLE", false)
	viper.SetDefault("ROLES.PERMISSIONS.SUPPERUSER", false)
	viper.SetDefault("ROLES.CAN_USE_API", false)
	viper.SetDefault("ROLES.ALLOW_SUDO", false)
	viper.SetDefault("ROLES.CONSOLE_SUPPER_USER", false)
	viper.SetDefault("ROLES.ALLOW_TO_LOGS_IN_SHARED", true)
	viper.SetDefault("LOG_FORMAT_JSON", false)
	viper.SetDefault("CONNECTOR.HOSTNAME", "localhost")
	viper.SetDefault("CONNECTOR.PORT", 8085)
	viper.SetDefault("SMC.API_VERSION", "6.7")
	viper.SetDefault("SMC.PORT", "8082")
	viper.SetDefault("SMC.NAME", "smc")
	viper.SetDefault("AZURE_ADMIN_LOGIN_NAME", "")
	viper.SetDefault("APP_NAME", "")
	viper.SetDefault("AZURE_ADMIN_LOGIN_PASSWORD", "")

	viper.AutomaticEnv() // read in environment variables that match
	if cfgFile != "" {
		if !lib.FileExists(cfgFile) {
			log.Fatal("the given config file does not exist")
		}
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".smcconnector" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		//viper.SetConfigName("smcconnector")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			if viper.GetBool("LOG_FORMAT_JSON") {
				logrus.SetFormatter(&logrus.JSONFormatter{})
			} else {
				logrus.SetFormatter(&logrus.TextFormatter{})
			}
		})
	}
	if viper.GetBool("LOG_FORMAT_JSON") {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	logrus.SetOutput(os.Stdout)

	SmcInstance = smc.Smc{
		Hostname:   viper.GetString("SMC.IP_ADDRESS"),
		Port:       viper.GetString("SMC.PORT"),
		AccessKey:  viper.GetString("SMC.KEY"),
		APIVersion: viper.GetString("SMC.API_VERSION"),
	}

}
