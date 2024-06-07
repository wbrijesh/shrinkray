package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"shrinkray/utils"

	"github.com/redis/go-redis/v9"
)

type ShortLink struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type PageData struct {
	Message string
}

var (
	ctx         = context.Background()
	port        = utils.LoadEnv("PORT")
	opt, _      = redis.ParseURL(utils.LoadEnv("REDIS_CONNECTION_STRING"))
	redisClient = redis.NewClient(opt)
)

func main() {
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/new", createShortLink)
	http.HandleFunc("/", findShortLink)
	http.HandleFunc("/temp", handler)

	fmt.Println("Server is running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": "OK"})
}

func createShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var shortLink ShortLink
	err := json.NewDecoder(r.Body).Decode(&shortLink)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if shortLink.URL == "" || shortLink.Alias == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "URL and Alias are required"})
		return
	}

	_, err = redisClient.Get(ctx, shortLink.Alias).Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Alias already exists"})
		return
	}

	res := redisClient.Set(ctx, shortLink.Alias, shortLink.URL, 0)

	if res.Err() != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": res.Err().Error()})
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"url": shortLink.URL, "alias": shortLink.Alias})
		return
	}
}

// func findShortLink(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	path := r.URL.Path
// 	alias := path[len("/"):]
// 	if alias == "" {
// 		http.Error(w, "Alias is required", http.StatusBadRequest)
// 		return
// 	}
//
// 	res, err := redisClient.Get(ctx, alias).Result()
// 	if err != nil {
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(map[string]string{"error": "Alias not found"})
// 		return
// 	} else {
// 		http.Redirect(w, r, res, http.StatusSeeOther)
// 		return
// 	}
// }

func findShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path
	alias := path[len("/"):]
	if alias == "" {
		http.Error(w, "Alias is required", http.StatusBadRequest)
		return
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("redirect.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := redisClient.Get(ctx, alias).Result()
	if err != nil {
		data := PageData{
			Message: "Alias not found",
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else {
		data := PageData{
			Message: res,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Define your dynamic data
	data := PageData{
		Message: "Hello, World!",
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template with the dynamic data
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
