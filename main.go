package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/naoina/toml"
)

const (
	// ConfigFile is the file which is read to get database and other configuration
	ConfigFile = "config.toml"
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
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
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

	go pool.run()

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.GET("/", index)
	router.GET("/ws", ws)
	router.GET("/history", history)

	router.Run("localhost:1234")
}

func index(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func history(c *gin.Context) {
	messages := recentMessages()
	c.JSON(200, messages)
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
