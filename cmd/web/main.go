package main

import (
	"fmt"
	"net/http"
	"os"

	"beeketing.com/consumer/config"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	config.Load()

	initLog()
}

func main() {
	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintln(w, "Hi")
	})

	restPath := viper.GetString("static.path") + "/rest"

	router.ServeFiles("/rest/*filepath", http.Dir(restPath))

	log.Info("Server listen and serve at http://localhost:8088")

	log.Fatal(http.ListenAndServe(":8088", router))
}

func initLog() {
	// log.SetFormatter(&log.JSONFormatter{})

	logOutput := viper.GetString("log.output")

	if logOutput == "file" {
		logFile, err := os.OpenFile("ccart.log", os.O_CREATE|os.O_WRONLY, 0666)

		if err == nil {
			log.SetOutput(logFile)
		} else {
			log.Fatal(err)
			log.Info("Failed to log to file, using default stderr")
		}
	}
}
