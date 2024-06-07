package main

import (
	"context"
	"encoding/json"
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

func init() {
	envs := []string{
		"PORT",
		"REDIS_CONNECTION_STRING",
		"RATE_LIMIT_REQ_PER_SEC",
	}

	for _, env := range envs {
		if !utils.VerifyEnv(env) {
			log.Fatalf("Environment variable %s is required", env)
		}
	}
}

func main() {
	reqMux := http.NewServeMux()

	reqMux.HandleFunc("/health", healthCheck)
	reqMux.HandleFunc("/new", createShortLink)
	reqMux.HandleFunc("/", findShortLink)

	log.Println("Server is running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, utils.Limit(reqMux)))
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

func findShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path
	alias := path[len("/"):]
	if alias == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Alias is required"})
		return
	}

	tmpl, err := template.ParseFiles("redirect.html")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Error parsing template" + err.Error()})
		return
	}

	res, err := redisClient.Get(ctx, alias).Result()
	if err != nil {
		data := PageData{
			Message: "Alias not found",
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": "Error executing template: " + err.Error()})
		}
		return
	} else {
		data := PageData{
			Message: res,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": "Error executing template: " + err.Error()})
		}
		return
	}
}
