package hotmaze

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandlerDownload reads file data from Firestore and writes it to w.
// In worflow D1 it is possible to access this URL before the resource has
// actually been uploaded. Thus in case of "not found", we retry every 400ms,
// up to 50s.
func (s Server) HandlerDownload(w http.ResponseWriter, r *http.Request) {
	fileUUID := strings.TrimPrefix(r.URL.Path, "/get/")
	log.Printf("Reading file %q\n", fileUUID)

	deadline := time.Now().Add(50 * time.Second)

	var err error
	var doc *firestore.DocumentSnapshot
	for time.Now().Before(deadline) {
		doc, err = s.FirestoreClient.Doc("D1/" + fileUUID + "_meta").Get(r.Context())
		if status.Code(err) == codes.NotFound {
			log.Println("D1/" + fileUUID + "_meta not found yet...")
			time.Sleep(400 * time.Millisecond)
			continue
		}
		break
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "Problem reading from Firestore :(", http.StatusInternalServerError)
		return
	}
	fields := doc.Data()
	fileSize64 := fields["size"].(int64)
	fileSize := int(fileSize64)

	fullData := make([]byte, fileSize)
	for k := 0; k*chunkSize < fileSize; k++ {
		log.Printf("Reading chunk %d\n", k)
		chunkPath := fmt.Sprintf("D1/%s_chunk_%d", fileUUID, k)
		for time.Now().Before(deadline) {
			doc, err = s.FirestoreClient.Doc(chunkPath).Get(r.Context())
			if status.Code(err) == codes.NotFound {
				log.Println(chunkPath + " not found yet...")
				time.Sleep(400 * time.Millisecond)
				continue
			}
			break
		}
		if err != nil {
			log.Println(err)
			http.Error(w, "Problem reading chunk meta from Firestore :(", http.StatusInternalServerError)
			return
		}
		chunkFields := doc.Data()
		if chunkContents, ok := chunkFields["data"].([]byte); ok {
			copy(fullData[k*chunkSize:], chunkContents)
		} else {
			log.Println("chunk has no []byte field named 'data'")
			http.Error(w, "Problem reading chunk data from Firestore :(", http.StatusInternalServerError)
			return
		}
	}

	if ct, ok := fields["type"].(string); ok {
		w.Header().Set("Content-Type", ct)
	}
	_, err = w.Write(fullData)
	if err != nil {
		log.Println("Writing", len(fullData), "bytes of file contents response:", err)
	}
}
