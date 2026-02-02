package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tpauling/handgover"
)

type Config struct {
	Hello     int           `env:"HELLO"`
	Timeout   time.Duration `env:"TIMEOUT" default:"1h"`
	FromRedis string        `redis:"my_redis_key_1"`
	FromHttp  []byte        `http:"https://filesamples.com/samples/code/json/sample1.json"`
}

func main() {
	os.Setenv("HELLO", "100")
	//os.Setenv("TIMEOUT", "5s")
	redis := RedisMock{}

	var config Config
	if err := ParseFrom(redis, &config); err != nil {
		log.Fatalln(err)
	}

	log.Println("Hello:", config.Hello)          // Hello: WORLD
	log.Println("Timeout:", config.Timeout)      // Timeout: 5s
	log.Println("From Redis:", config.FromRedis) // From Redis: Some value from redis!
	log.Println("From HTTP:", config.FromHttp)   // From HTTP: { "fruit": "Apple", "size": "Large", "color": "Red" }
}

type RedisMock struct{}

func (rm RedisMock) Get(key string) (string, error) {
	return "Some value from redis!", nil
}

func ParseFrom(rm RedisMock, obj interface{}) error {
	sources := []handgover.Source{
		handgover.NewSource(
			"default",
			func(field string) (handgover.Valuer, error) {
				return handgover.Value(field), nil
			},
		),
		handgover.NewSource(
			"env",
			func(field string) (handgover.Valuer, error) {

				val, ok := os.LookupEnv(field)
				if !ok {
					return nil, nil
				}

				return handgover.Value(val), nil
			},
		),
		handgover.NewSource(
			"redit",
			func(field string) (handgover.Valuer, error) {
				value, err := rm.Get(field)
				if err != nil {
					return nil, err
				}
				return handgover.Value(value), nil
			},
		),
		handgover.NewSource(
			"redit",
			func(field string) (handgover.Valuer, error) {
				res, err := http.Get(field)
				if err != nil {
					return nil, err
				}
				defer res.Body.Close()

				data, err := io.ReadAll(res.Body)
				if err != nil {
					return nil, err
				}
				return handgover.Value(string(data)), nil
			},
		),
	}
	return handgover.From(sources).To(obj)
}
