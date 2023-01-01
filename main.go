package main

import (
	"L5/env"
	"L5/interruption"
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"sync"
)

const arraySize = 30

var indices = make([]int, arraySize)

func init() {
	for i := range indices {
		indices[i] = i
	}
}

type UserData struct {
	Array [arraySize]int
	Mu    sync.Mutex
}

func (data *UserData) Get(i int) int {
	data.Mu.Lock()
	defer data.Mu.Unlock()
	return data.Array[i]
}

func (data *UserData) Set(i int, val int) {
	data.Mu.Lock()
	defer data.Mu.Unlock()
	data.Array[i] = val
}

func main() {
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.LoadHTMLGlob("static/*")

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Panicln("[main] error setting trusted proxies:", err)
	}

	router.GET("/", func(c *gin.Context) {
		userData := updateCookie(c)
		userData.Mu.Lock()
		defer userData.Mu.Unlock()
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"numbers": userData.Array,
			"indices": indices,
		})
	})

	api := router.Group("/api")

	api.GET("/get", func(c *gin.Context) {
		userData := updateCookie(c)

		indexString := c.Query("i")
		index, err := strconv.Atoi(indexString)
		if err != nil || index < 0 || index >= arraySize {
			log.Println("index =", indexString, "& error:", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"value": userData.Get(index),
		})
	})

	api.PUT("/set", func(c *gin.Context) {
		userData := updateCookie(c)

		indexString := c.Query("i")
		index, err := strconv.Atoi(indexString)
		if err != nil || index < 0 || index >= arraySize {
			log.Println("index =", indexString, "& error:", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		valueString := c.Query("v")
		value, err := strconv.Atoi(valueString)
		if err != nil {
			log.Println("value =", valueString, "& error:", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		userData.Set(index, value)
		c.Status(http.StatusOK)
	})

	go func() {
		if err := router.Run(":" + env.Port()); err != nil {
			log.Println("[main] error running server:", err)
		}
	}()

	interruption.Wait()
}

var sessions sync.Map

func updateCookie(c *gin.Context) *UserData {
	cookie, err := c.Cookie("sessionId")
	if err == nil {
		if userData, ok := sessions.Load(cookie); ok {
			return userData.(*UserData)
		}
	} else if err != http.ErrNoCookie {
		log.Panicln("[main.updateCookie] unexpected cookie error:", err)
	}
	token, userData := newToken()
	c.SetCookie("sessionId", token, 0, "/", env.Domain(), true, true)
	return userData
}

func newToken() (string, *UserData) {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		log.Panicln("[main.updateCookie] unexpected rand gen error:", err)
	}

	tokenString := base64.RawURLEncoding.EncodeToString(token[:])

	userData := new(UserData)
	for i := range userData.Array {
		userData.Array[i] = i + 1
	}

	if _, loaded := sessions.LoadOrStore(tokenString, userData); loaded {
		return newToken()
	}

	return tokenString, userData
}
