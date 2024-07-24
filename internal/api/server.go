package api

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"

	// "time"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"

	// "github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	// "github.com/edulinq/autograder/internal/util"
)
var API_REQUEST_CONTENT_KEY = "content"
var ROOT_USER_NUMBER_KEY = "number"

// Run the standard API and Unix server.
func StartServer() error {
	serverErrorChannel := make(chan error, 2)

	go func() {
		serverErrorChannel <- startAPIServer()
	}()

	go func() {
		serverErrorChannel <- startUnixServer()
	}()

	for {
		select {
			case err := <-serverErrorChannel:
				return fmt.Errorf("server error: %v", err)
			case <- waitForShutdownSignal():
				return nil
    	}
	}
}

func waitForShutdownSignal() <-chan os.Signal {
    serverShutdownChannel := make(chan os.Signal, 1)
    signal.Notify(serverShutdownChannel, os.Interrupt, syscall.SIGTERM)

    return serverShutdownChannel
}

func StartUnixServer() {
	err := startUnixServer()
	if (err != nil) {
		log.Fatal("Failed to start the Unix server.", err)
	}
}

func startAPIServer() error {
	var port = config.WEB_PORT.Get()

	log.Info("API Server Started", log.NewAttr("port", port))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), core.GetRouteServer(GetRoutes()));
}

func startUnixServer() error {
	var socketPath = config.UNIX_SOCKET.Get();

	unixListener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Error creating Unix socket. ", err)
	}

	log.Info("Unix Server Started", log.NewAttr("unix_socket", socketPath))
	// go func() {
	// 	<-waitForShutdownSignal()
	// 	unixListener.Close()
	// 	os.Remove(socketPath)
	// }()

	// defer unixListener.Close()
	defer os.Remove(socketPath)
	

	for {
		connection, err := unixListener.Accept()
		if (err != nil) {
			log.Fatal("Failed to accept a connection. ", err)
		}
		go handleConnection(connection)
	}
}

func handleConnection(conn net.Conn) {
	go func() {
		http.ListenAndServe("127.0.0.1:8080", core.GetRouteServer(GetRoutes()))
	}()
	time.Sleep(1 * time.Second)

	defer func() {
		fmt.Println("Closing connection")
	}() 
	
	sizeBuffer := make([]byte, 8)
	_, err := conn.Read(sizeBuffer)
	if err != nil {
		log.Fatal("Failed to read size of the payload. ", err)
	}

	size := binary.BigEndian.Uint64(sizeBuffer)
	
	jsonBuffer := make([]byte, size)
	_, err = conn.Read(jsonBuffer)
	if err != nil {
		log.Error("Failed to read JSON payload", err)
		return
	}


	var request any
	if err := json.Unmarshal(jsonBuffer, &request); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	randomNumber, err := rand.Int(rand.Reader, big.NewInt(10000000))
	if err != nil {
		log.Fatal("Failed to generate random number.", err)
	}

	randomNumberString := strconv.Itoa(int(randomNumber.Int64()))

	core.RandomNumberMap.Store(randomNumberString, randomNumberString)

	// defer core.RandomNumberMap.Delete(randomNumber)

	fmt.Println("type: ", reflect.TypeOf(strconv.Itoa(int(randomNumber.Int64()))))
	form := map[string]string{
		API_REQUEST_CONTENT_KEY: util.MustToJSON(request),
		ROOT_USER_NUMBER_KEY: randomNumberString,
	}

	// fmt.Println("form: ", util.MustToJSONIndent(form))

	url := "http://127.0.0.1:8080" + core.NewEndpoint(`submissions/peek`)

	response, err := common.PostNoCheck(url, form)
	if (err != nil) {
		fmt.Println("Error: ", err)
	}
	fmt.Println("response: ", response)

	// Echo the data back to the connection.
	_, err = conn.Write([]byte("Request received"))
	if err != nil {
		log.Fatal("Failed to echo back data.", err)
	}
}

// func handleGenerate(conn net.Conn) {
// 	randomNumber, err := rand.Int(rand.Reader, big.NewInt(10000000))
// 	if err != nil {
// 		log.Fatal("Failed to generate random number.", err)
// 	}

// 	numberBytes := randomNumber.Bytes()
// 	length := uint64(len(numberBytes))

// 	buffer := new(bytes.Buffer)


// 	// Send the length of the number
// 	err = binary.Write(buffer, binary.BigEndian, length)
// 	if err != nil {
// 		log.Fatal("Failed to write length to connection.", err)
// 	}

// 	buffer.Write(numberBytes)

// 	// Send the number itself
// 	_, err = conn.Write(buffer.Bytes())
// 	if err != nil {
// 		log.Fatal("Failed to write number to connection.", err)
// 	}
// }

// func handleValidate(conn net.Conn, request map[string]string) {
// 	value := request["value"]
// 	fmt.Println("value in validate: ", value)


// 	// value, err := strconv.Atoi(request["value"])
// 	// if err != nil {
// 	// 	log.Error("Invalid value:", err)
// 	// 	return
// 	// }

// 	// expectedValue, exists := RandomNumberMap[key]
// 	// response := map[string]string{"result": "failure"}
// 	// if exists && expectedValue.Cmp(big.NewInt(int64(value))) == 0 {
// 	// 	response["result"] = "success"
// 	// }

// 	// encoder := json.NewEncoder(conn)
// 	// if err := encoder.Encode(response); err != nil {
// 	// 	log.Error("Encode error:", err)
// 	// }
// }
