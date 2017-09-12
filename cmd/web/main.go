package main

import (
	"fmt"
	"net/http"

	"beeketing.com/consumer/config"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	config.Load()
}

func main() {
	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintln(w, "Hi")
	})

	restPath := viper.GetString("static.path") + "/rest"

	router.ServeFiles("/rest/*filepath", http.Dir(restPath))

	log.Info("Server listen and serve at http://localhost:8092")

	log.Fatal(http.ListenAndServe(":8088", router))
}
