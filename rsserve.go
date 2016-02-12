// GOLang Cross Compiling
// https://medium.com/@rakyll/go-1-5-cross-compilation-488092ba44ec#.z0ilhvual
// linux	amd64

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/redis.v3"
)

type redisConnection struct {
	address string
	port    string
	pass    string
	client  *redis.Client
}

type redisServer struct {
	keyPrefix string
	keySuffix string
	redisKey  string

	address string
	port    string
}

func main() {
	// Set up 'viper' configuration and point it to config.json
	setupConfig()

	// Set defaults
	viper.SetDefault("redis_port", 6379)
	viper.SetDefault("redis_address", "localhost")
	viper.SetDefault("redis_pass", "")
	viper.SetDefault("key_suffix", ":index.html")
	viper.SetDefault("http_address", "localhost")
	viper.SetDefault("http_port", 80)

	// Collect all required configuration directives
	redisDb := redisConnection{
		address: viper.GetString("redis_address"),
		port:    viper.GetString("redis_port"),
		pass:    viper.GetString("redis_pass"),
	}

	// Connect to Redis
	redisDb.mustConnectToRedis()

	// Prepare web server
	htmlServer := redisServer{
		keyPrefix: mustGetConfig("key_prefix"),
		keySuffix: viper.GetString("key_suffix"),
		address:   viper.GetString("http_address"),
		port:      viper.GetString("http_port"),
	}
	htmlServer.redisKey = fmt.Sprintf("%s%s", htmlServer.keyPrefix, htmlServer.keySuffix)

	// Register handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// Redis key prefix
		redisKey := htmlServer.redisKey

		// Has the user requested a specific version?
		queries := r.URL.Query()
		if queries.Get("version") != "" {
			redisKey = fmt.Sprintf("%s%s", queries.Get("version"), htmlServer.keySuffix)
		}

		// Get the current revision to publish
		html, err := redisDb.client.Get(redisKey).Result()
		if err != nil {
			log.Println("Could not find current revision based on key:", redisKey)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Send the result back to the client
		fmt.Fprintf(w, html)
	})

	// Start web server
	host := fmt.Sprintf("%s:%s", htmlServer.address, htmlServer.port)
	fmt.Println("Listening at ...", host)
	log.Fatal(http.ListenAndServe(host, nil))
}

func setupConfig() {
	var configPath = flag.String("c", "", "Configuration file path.")
	flag.Parse()

	if *configPath == "" {
		log.Println("No configuration path found. Will attempt to use the current path.")
		*configPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}

	// Collect environment variables
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(*configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Could not read configuration from path %s. Err: %s\n", *configPath, err)
		log.Fatalln("config.json must be located in the current working directory or the -c argument should be provided.")
	}
}

// mustGetConfig will attempt to retrieve a config option from viper and will
// fatal error if the argument was not found.
func mustGetConfig(key string) (ret string) {
	ret = viper.GetString(key)
	if ret == "" {
		log.Fatalf("Could not parse '%s' from config.\n", key)
	}
	return
}

func (db *redisConnection) mustConnectToRedis() {
	db.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", db.address, db.port),
		Password: db.pass,
		DB:       0, // Use default DB
	})

	// Ensure the connection has been established
	if _, err := db.client.Ping().Result(); err != nil {
		log.Fatalln("Could not ping Redis after connection.", err)
	}
}
