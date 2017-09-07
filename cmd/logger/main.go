package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"beeketing.com/consumer/config"
	"beeketing.com/consumer/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	config.Load()

	redis.Init()
}

const batchSize int64 = 100
const listName string = "logger"
const delayTime time.Duration = time.Millisecond * 500

func main() {
	start()
}

func start() {
	filePath := viper.GetString("storage.path") + "/logs/main.log"

	for {
		// read
		messages, err := redis.Client.LRange(listName, 0, batchSize-1).Result()
		failOnError(err, "lrange logger failed")

		for i := 0; i < len(messages); i++ {
			err = redis.Client.LPop("logger").Err()
			if err != nil {
				log.Error("Lpop failed", err)
			}
		}

		log.Info("Messages:", len(messages))

		if len(messages) > 0 {
			data := fmt.Sprintf("%s\n", strings.Join(messages, "\n"))

			f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				failOnError(err, "Open log file failed")
			}

			if _, err = f.WriteString(data); err != nil {
				failOnError(err, "Write data to file failed")
			}

			if err := f.Close(); err != nil {
				failOnError(err, "Close file failed")
			}
		}

		time.Sleep(delayTime)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
