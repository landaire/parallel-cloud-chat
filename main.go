package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/naoina/toml"
)

const (
	// ConfigFile is the file which is read to get database and other configuration
	ConfigFile = "config.toml"

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	db *sql.DB
)

type config struct {
	Database databaseConfig `toml:"database"`
}

// DatabaseConfig rep
type databaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

func main() {
	var err error
	config := readConfig()

	db, err = sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=verify-full",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
		),
	)

	if err != nil {
		panic(err)
	}

	createTables()

	router := gin.Default()

	router.GET("/", index)
	router.GET("/ws", ws)
	router.GET("/history", history)

	router.Run("localhost:1234")
}

func index(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func history(c *gin.Context) {
}

func ws(c *gin.Context) {
	wsHandler(c.Writer, c.Request)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade websocket", err)
		return
	}

	c := &connection{
		username: r.URL.Query().Get("username"),
		ws:       conn,
		messages: make(chan []byte, 256),
	}

	go c.sendMessages()
	c.readMessages()
}

func readConfig() config {
	var out config
	file, err := os.Open(ConfigFile)
	if err != nil {
		panic("Could not read configuration")
	}

	defer file.Close()

	decoder := toml.NewDecoder(file)

	if err := decoder.Decode(&out); err != nil {
		panic(err)
	}

	return out
}