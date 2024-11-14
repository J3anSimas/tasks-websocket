package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"tasks-websocket/internal/config"
	"tasks-websocket/internal/handlers"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/sijms/go-ora/v2"
)

func init() {
	config.Instantiate()
}

type Client struct {
	conn  *websocket.Conn // Conexão WebSocket
	board string          // Sala em que o cliente está
}
type Hub struct {
	boards     map[string]map[*websocket.Conn]bool // Mapa de salas e suas conexões
	register   chan *Client                        // Canal para registrar novos clientes
	unregister chan *Client                        // Canal para remover clientes
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if h.boards[client.board] == nil {
				h.boards[client.board] = make(map[*websocket.Conn]bool)
			}
			h.boards[client.board][client.conn] = true

		case client := <-h.unregister:
			if clients, ok := h.boards[client.board]; ok {
				if _, ok := clients[client.conn]; ok {
					delete(clients, client.conn)
					if len(clients) == 0 {
						delete(h.boards, client.board)
					}
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	db, err := sql.Open("oracle", config.Cfg.DBConnectionString)
	hub := &Hub{
		boards:     make(map[string]map[*websocket.Conn]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	// Iniciando o Hub em uma goroutine
	go hub.run()
	if err != nil {
		panic(fmt.Errorf("error in sql.Open: %w", err))
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()
	r := gin.Default()
	r.LoadHTMLGlob("internal/views/*")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/authenticate", func(ctx *gin.Context) {
		handlers.Authenticate(db, ctx)
	})

	r.GET("/login", handlers.RenderLoginPage)
	authorized := r.Group("/")
	authorized.Use(handlers.EnsureAuthenticated())
	authorized.GET("/", func(ctx *gin.Context) {
		handlers.RenderIndexPage(ctx, db)
	})

	r.GET("/ws/:board", func(c *gin.Context) {
		board := c.Param("board")
		if board == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Board parameter is required"})
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			// panic(err)
			log.Printf("%s, error while Upgrading websocket connection\n", err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		client := &Client{conn: conn, board: board}
		hub.register <- client
		defer func() {
			hub.unregister <- client
			conn.Close()
		}()

		for {
			// Read message from client
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				// panic(err)
				log.Printf("%s, error while reading message\n", err.Error())
				c.AbortWithError(http.StatusInternalServerError, err)
				break
			}

			type Message struct {
				CardId string `json:"cardId"`
				Status string `json:"status"`
			}
			msg := Message{}
			json.Unmarshal(p, &msg)
			fmt.Println("Messagem recebida", msg.Status, msg.CardId)
			if clients, ok := hub.boards[board]; ok {
				for client := range clients {
					if client != conn {
						err = client.WriteMessage(messageType, p)
						if err != nil {
							// panic(err)
							log.Printf("%s, error while writing message\n", err.Error())
							c.AbortWithError(http.StatusInternalServerError, err)
							break
						}
					}
				}
			}
			// Echo message back to client
			err = conn.WriteMessage(messageType, p)
			if err != nil {
				// panic(err)
				log.Printf("%s, error while writing message\n", err.Error())
				c.AbortWithError(http.StatusInternalServerError, err)
				break
			}
		}
	})

	authorized.GET("/boards/:id", func(ctx *gin.Context) {
		handlers.RenderBoardPage(ctx, db)
	})
	authorized.PUT("/cards/:id", func(ctx *gin.Context) {
		handlers.UpdateCardStatus(ctx, db)
	})
	r.Run()
}

// connectionString := "oracle://" + dbParams["username"] + ":" + dbParams["password"] + "@" + dbParams["server"] + ":" + dbParams["port"] + "/" + dbParams["service"]
// if val, ok := dbParams["walletLocation"]; ok && val != "" {
// 	connectionString += "?TRACE FILE=trace.log&SSL=enable&SSL Verify=false&WALLET=" + url.QueryEscape(dbParams["walletLocation"])
// }
// fmt.Println(connectionString)
// db, err := sql.Open("oracle", connectionString)
// if err != nil {
// 	panic(fmt.Errorf("error in sql.Open: %w", err))
// }

// err = db.Ping()
// if err != nil {
// 	panic(fmt.Errorf("error pinging db: %w", err))
// }
// rows, err := db.Query("SELECT user_id, login FROM usuario where login = 'jean.simas'")
// if err != nil {
// 	panic(fmt.Errorf("error in db.Query: %w", err))
// }
// for rows.Next() {
// 	var user User
// 	err = rows.Scan(&user.ID, &user.Name)
// 	if err != nil {
// 		panic(fmt.Errorf("error in rows.Scan: %w", err))
// 	}
// 	fmt.Printf("User: %v\n", user)
// }
