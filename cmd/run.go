package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the connector service",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {
		//check if the defined roles in config fine are allowed
		if viper.GetBool("ROLES.CONSOLE_SUPPER_USER") && !viper.GetBool("ROLES.PERMISSIONS.SUPPERUSER") {
			logrus.Error("in config file the consoleSuperuser can be true if and only if roles.permissions.Superuser is true. Please address this issue and rerun the service")
			fmt.Print("service is terminated: ")
			os.Exit(1)

		}

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Printf("\nCTRL-C: ")
			os.Exit(1)
		}()
		go func() {
			var AzureCLIInstance AzureCLI
			// get app assigned users
			if !AzureCLIInstance.IsLogin {
				if err := AzureCLIInstance.Login(); err != nil {
					logrus.Fatal(err)
				}
				logrus.Info("login to azure.... Done")
			}
			for {

				time.Sleep(time.Duration(viper.GetInt("ROLES_UPDATE_TIME_IN_MINUTES")) * time.Minute)
				ApplyRoles()
				if err := DetectDeletedUsers(); err != nil {
					logrus.Error(err)
				}
			}
		}()
		muxRouter := mux.NewRouter().StrictSlash(true)
		router := AddRoutes(muxRouter)
		err := http.ListenAndServe(viper.GetString("CONNECTOR.HOSTNAME")+":"+viper.GetString("CONNECTOR.PORT"),
			router)
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
