package coolmaze

import (
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// For security reasons, listening webpage must register:
// - backend generates a new pair (qrKey, chanID),
// - backend stores the pair in Datastore + Memcache,
// - backend gives the pair to the webpage,
// - webpage displays a 2D matrix for qrKey,
// - webpage listens to channel chanID,
// - mobile reads qrKey,
// - mobile sends (qrKey, message) to backend,
// - backend reads mapping, and sends message to channel chanID,
// - webpage receives message.

func init() {
	// This is important, to generate always different pairs
	rndOnce.Do(randomize)

	http.HandleFunc("/register", register)
}

func register(w http.ResponseWriter, r *http.Request) {
	accessControlAllowCoolMaze(w, r)
	c := appengine.NewContext(r)

	if r.Method != "POST" {
		log.Warningf(c, "Only POST method is accepted")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Only POST method is accepted")
		return
	}

	qrKey, chanID, err := generatePair(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof(c, "Registered new pair [%q, %q]", qrKey, chanID)
	response := Response{
		"qrKey":  qrKey,
		"chanID": chanID,
	}
	fmt.Fprintln(w, response)
}

// CORS non-sense.
// See http://stackoverflow.com/a/1850482/871134 .
func accessControlAllowCoolMaze(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	origin = strings.TrimSpace(origin)

	// Specific hosts.
	whiteList := []string{
		"https://coolmaze.net",
		"https://www.coolmaze.net",
		"https://coolmaze.io",
		"https://www.coolmaze.io",
		// For debug.
		"http://localhost:8080",
	}
	for _, whiteItem := range whiteList {
		if origin == whiteItem {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			return
		}
	}

	// WebExtension prefixes.
	whitePrefixes := []string{
		"moz-extension://",
		"chrome-extension://",
	}
	for _, whitePrefix := range whitePrefixes {
		if strings.HasPrefix(origin, whitePrefix) {
			// TODO set a more specific WebExtension test
			// TODO make sure this also works in IE
			w.Header().Set("Access-Control-Allow-Origin", origin)
			return
		}
	}

	c := appengine.NewContext(r)
	log.Warningf(c, "Origin [%s] not in whitelist", origin)

	//w.Header().Add("Access-Control-Allow-Origin", "*")
}
