package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/hoisie/redis"
)

var client redis.Client
var templates = template.Must(template.ParseGlob("templates/*"))
var validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)$")

func main() {
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8889", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		postHandler(w, r)
	} else {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			handleDefault(w, r)
			return
		}
		getHandler(w, r, m[1])
	}
}

func handleDefault(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "indexPage",
		map[string]interface{}{"Host": r.Host})
	checkError(err, w)
}

func generateKey(size int64, w http.ResponseWriter) (key string) {
	var initData []byte

	i := 0
	for ; i < 10; i++ {
		key = randStr(size)
		success, err := client.Setnx(key, initData)
		checkError(err, w)
		if success {
			break
		}
	}

	if i >= 10 {
		http.Error(w, "Could not find a free key",
			http.StatusInternalServerError)
	}
	return
}

func randStr(str_size int64) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, str_size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	checkError(err, w)

	values := r.URL.Query()
	timeout := values.Get("timeout")

	if timeout == "" {
		timeout = "1800"
	}

	timeoutI, err := strconv.ParseInt(timeout, 10, 64)
	checkError(err, w)
	if timeoutI > 1800 {
		timeoutI = 1800
	}

	key := generateKey(10, w)

	size := len(body)
	if size > 2048 {
		size = 2048
	}

	client.Setex(key, timeoutI, []byte(body)[:size])

	fmt.Fprintf(w, "http://%s/%s", r.Host, key)
}

func getHandler(w http.ResponseWriter, r *http.Request, key string) {
	exists, err := client.Exists(key)
	if err != nil || !exists {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	result, err := client.Get(key)

	binary.Write(w, binary.LittleEndian, result)
}

func checkError(err error, w http.ResponseWriter) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
