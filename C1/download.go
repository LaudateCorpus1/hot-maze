package hotmaze

import (
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandlerDownload reads file data from Firestore and writes it to w.
func (s Server) HandlerDownload(w http.ResponseWriter, r *http.Request) {
	fileUUID := strings.TrimPrefix(r.URL.Path, "/get/")
	log.Printf("Reading file %q\n", fileUUID)

	doc, errGet := s.FirestoreClient.Doc("C1/" + fileUUID).Get(r.Context())
	if status.Code(errGet) == codes.NotFound {
		log.Println("C1/" + fileUUID + " not found")
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}
	if errGet != nil {
		log.Println(errGet)
		http.Error(w, "Problem reading from Firestore :(", http.StatusInternalServerError)
		return
	}

	fields := doc.Data()
	if ct, ok := fields["type"].(string); ok {
		w.Header().Set("Content-Type", ct)
	}
	if fileContents, ok := fields["data"].([]byte); ok {
		_, _ = w.Write(fileContents)
	}
}
