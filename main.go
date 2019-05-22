package main

import (
	"flag"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/connectome-neuprint/neuPrintHTTP/api"
	"github.com/connectome-neuprint/neuPrintHTTP/config"
	"github.com/janelia-flyem/echo-secure"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func customUsage() {
	fmt.Printf("Usage: %s [OPTIONS] CONFIG.json\n", os.Args[0])
	flag.PrintDefaults()
}

type KafkaLog struct {
	Producer *kafka.Producer
	Topic    string
}

func (k *KafkaLog) Write(p []byte) (int, error) {
	kafkaMsg := &kafka.Message{

		TopicPartition: kafka.TopicPartition{Topic: &k.Topic, Partition: kafka.PartitionAny},

		Value: p,

		Timestamp: time.Now(),
	}

	if err := k.Producer.Produce(kafkaMsg, nil); err != nil {
		return 0, err
	}

	return len(p), nil

}

func main() {
	// create command line argument for port
	var port int = 11000
	var publicRead bool = false
	flag.Usage = customUsage
	flag.IntVar(&port, "port", 11000, "port to start server")
	flag.BoolVar(&publicRead, "public_read", false, "allow all users read access")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		return
	}

	// parse options
	options, err := config.LoadConfig(flag.Args()[0])
	if err != nil {
		fmt.Print(err)
		return
	}

	// create datastore based on configuration
	store, err := config.CreateStore(options)
	if err != nil {
		fmt.Println(err)
		return
	}

	// create echo web framework
	e := echo.New()

	// setup logger
	logFile := os.Stdout

	if options.LoggerFile != "" {
		if logFile, err = os.OpenFile(options.LoggerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			fmt.Println(err)
			return
		}
		defer logFile.Close()
	}
	logWriter := io.Writer(logFile)

	// use kafka for logging if available
	if len(options.KafkaServers) > 0 {
		serverstr := strings.Join(options.KafkaServers, ",")
		kp, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": serverstr})

		portstr := strconv.Itoa(port)
		logWriter = &KafkaLog{kp, "neuprint_" + options.Hostname + "_" + portstr}
	}

	e.Use(LoggerWithConfig(LoggerConfig{
		Format: "{\"uri\": \"${uri}\", \"status\": ${status}, \"bytes_in\": ${bytes_in}, \"bytes_out\": ${bytes_out}, \"duration\": ${latency}, \"time\": ${time_unix}, \"user\": \"${custom:email}\", \"category\": \"${category}\", \"debug\": \"${custom:debug}\"}\n",
		Output: logWriter,
	}))

	e.Use(middleware.Recover())
	e.Pre(middleware.NonWWWRedirect())

	var authorizer secure.Authorizer
	// call new secure API and set authorization method
	if options.AuthDatastore != "" {
		authorizer, err = secure.NewDatastoreAuthorizer(options.AuthDatastore, options.AuthToken)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		authorizer, err = secure.NewFileAuthorizer(options.AuthFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	sconfig := secure.SecureConfig{
		SSLCert:          options.CertPEM,
		SSLKey:           options.KeyPEM,
		ClientID:         options.ClientID,
		ClientSecret:     options.ClientSecret,
		AuthorizeChecker: authorizer,
		Hostname:         options.Hostname,
	}
	secureAPI, err := secure.InitializeEchoSecure(e, sconfig, []byte(options.Secret))
	if err != nil {
		fmt.Println(err)
		return
	}

	// create read only group
	readGrp := e.Group("/api")
	if publicRead {
		readGrp.Use(secureAPI.AuthMiddleware(secure.NOAUTH))
	} else {
		readGrp.Use(secureAPI.AuthMiddleware(secure.READ))
	}
	// setup server status message to show if it is public
	e.GET("/api/serverinfo", secureAPI.AuthMiddleware(secure.NOAUTH)(func(c echo.Context) error {
		info := struct {
			IsPublic bool
		}{publicRead}
		return c.JSON(http.StatusOK, info)
	}))

	// setup default page
	// TODO: point to swagger documentation
	if options.StaticDir != "" {
		e.Static("/", options.StaticDir)
		customHTTPErrorHandler := func(err error, c echo.Context) {
			if he, ok := err.(*echo.HTTPError); ok {
				req := c.Request()
				if !strings.HasPrefix(req.RequestURI, "/api") && (he.Code == http.StatusNotFound) {
					c.File(options.StaticDir)
				}
			}
			e.DefaultHTTPErrorHandler(err, c)
		}

		e.HTTPErrorHandler = customHTTPErrorHandler

	} else {
		e.GET("/", secureAPI.AuthMiddleware(secure.NOAUTH)(func(c echo.Context) error {
			return c.HTML(http.StatusOK, "<html><title>neuprint http</title><body><a href='/token'><button>Download API Token</button></a><p><b>Example query using neo4j cypher:</b><br>curl -X GET -H \"Content-Type: application/json\" -H \"Authorization: Bearer YOURTOKEN\" https://SERVERADDR/api/custom/custom -d '{\"cypher\": \"MATCH (m :Meta) RETURN m.dataset AS dataset, m.lastDatabaseEdit AS lastmod\"}'</p><a href='/api/help'>Documentation</a><form action='/logout' method='post'><input type='submit' value='Logout' /></form></body></html>")
		}))
	}

	if options.SwaggerDir != "" {
		e.Static("/api/help", options.SwaggerDir)
	}

	// load connectomic READ-ONLY API
	if err = api.SetupRoutes(e, readGrp, store); err != nil {
		fmt.Print(err)
		return
	}

	// start server
	secureAPI.StartEchoSecure(port)
}
