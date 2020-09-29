package hotmaze

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	validity  = 300 * time.Second
	KB        = 1024
	chunkSize = 512 * KB
)

func (s Server) HandlerUpload(w http.ResponseWriter, r *http.Request) {
	if errCORS := s.accessControlAllowHotMaze(w, r); errCORS != nil {
		log.Println(errCORS)
		http.Error(w, errCORS.Error(), http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusBadRequest)
		return
	}

	fileData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read file data :(", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	if len(fileData) == 0 {
		http.Error(w, "Empty file :(", http.StatusBadRequest)
		return
	}
	contentType := r.Header.Get("Content-Type")
	fileUUID := r.FormValue("uuid")
	if fileUUID == "" {
		http.Error(w, "Please provide uuid", http.StatusBadRequest)
		return
	}
	log.Println("Received resource", fileUUID, "of type", contentType, "and size", len(fileData))

	_, errCreate := s.FirestoreClient.Doc("D1/"+fileUUID+"_meta").Create(r.Context(), map[string]interface{}{
		"type":    contentType,
		"size":    len(fileData),
		"created": time.Now(),
	})
	if errCreate != nil {
		log.Println("Writing resource meta:", errCreate)
		http.Error(w, "Problem writing to Firestore :(", http.StatusInternalServerError)
		return
	}

	// Write file contents in 512KB chunks.
	// (a Firestore document can never hold more than 1MB)
	nbChunks := (len(fileData) + chunkSize - 1) / chunkSize

	_, err = s.ScheduleForgetFile(r.Context(), fileUUID, nbChunks)
	if err != nil {
		log.Println("scheduling file expiry:", err)
		// Better fail now, than keeping a user file forever in Firestore
		http.Error(w, "Problem with file allocation :(", http.StatusInternalServerError)
		return
	}
	log.Println("File", fileUUID, "is scheduled for deletion")

	for k := 0; k*chunkSize < len(fileData); k++ {
		var chunk []byte
		if len(fileData) < (k+1)*chunkSize {
			chunk = fileData[k*chunkSize:]
		} else {
			chunk = fileData[k*chunkSize : (k+1)*chunkSize]
		}
		log.Println("For UUID", fileUUID, "writing chunk", k)
		chunkPath := fmt.Sprintf("D1/%s_chunk_%d", fileUUID, k)
		_, errCreate := s.FirestoreClient.Doc(chunkPath).Create(r.Context(), map[string]interface{}{
			"data": chunk,
		})
		if errCreate != nil {
			log.Println("Writing chunk:", errCreate)
			http.Error(w, "Problem writing to Firestore :(", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"downloadURL": s.BackendBaseURL + "/get/" + fileUUID,
	})
}
