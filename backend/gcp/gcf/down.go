package p

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
)

// Get retrieves the data of a resource and writes it to the response body.
func Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Println("Wrong method", r.Method)
		http.Error(w, "Only GET requests allowed", http.StatusBadRequest)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	log.Println("file id =", id)

	data, filename, err := retrieve(id)
	if err != nil {
		log.Println("Retrieving resource:", err)
		http.Error(w, "Technical problem :(", http.StatusInternalServerError)
		return
	}
	log.Printf("Found something of length %d with filename %q\n", len(data), filename)

	w.Header().Set("Content-Disposition", "filename="+filename)
	_, err = w.Write(data)
	if err != nil {
		log.Println("Error writing resource:", err)
	}
}

func retrieve(id string) (data []byte, filename string, err error) {
	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT")
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, "", err
	}
	docref := client.Doc("Files/" + id)
	docsnap, err := docref.Get(ctx)
	if err != nil {
		return nil, "", err
	}
	data = docsnap.Data()["Data"].([]byte)
	filename = docsnap.Data()["Filename"].(string)
	return data, filename, nil
}
