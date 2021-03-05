package hotmaze

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

// HandlerDirectUpload is used from a terminal with cURL.
// It receives file data, produces an ASCII-art QR code in the response body,
// and doesn't require a web browser in the source computer (sender-side).
func (s *Server) HandlerDirectUpload(w http.ResponseWriter, r *http.Request) {
	filedata, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Reading request body:", err)
		http.Error(w, "Could not read request :(", http.StatusInternalServerError)
		return
	}
	if len(bytes.TrimSpace(filedata)) == 0 {
		log.Println("Empty request body")
		http.Error(w, s.usage(), http.StatusBadRequest)
		return
	}
	log.Println("Received resource of size", len(filedata))
	var filename string
	if f := r.FormValue("filename"); f != "" {
		filename = f
	}

	//
	// Store the incoming file data into a new temp GCS object
	//

	fileUUID := uuid.New().String()
	objectName := "transit/" + fileUUID
	objectHandle := s.StorageClient.Bucket(s.StorageBucket).Object(objectName)
	fileWriter := objectHandle.NewWriter(r.Context())
	n, err := fileWriter.Write(filedata)
	if err != nil {
		log.Printf("Writing %d bytes to object %q in bucket %q: %v", len(filedata), objectName, s.StorageBucket, err)
		http.Error(w, "Could not write file to GCS :(", http.StatusInternalServerError)
		return
	}
	if n != len(filedata) {
		log.Printf("Wrote only %d / %d bytes to object %q in bucket %q: %v", n, len(filedata), objectName, s.StorageBucket, err)
		http.Error(w, "Could not write file to GCS :(", http.StatusInternalServerError)
		return
	}
	err = fileWriter.Close()
	if err != nil {
		log.Printf("Writing %d bytes to object %q in bucket %q: %v", len(filedata), objectName, s.StorageBucket, err)
		http.Error(w, "Could not write file to GCS :(", http.StatusInternalServerError)
		return
	}
	if filename != "" {
		_, err := objectHandle.Update(r.Context(), storage.ObjectAttrsToUpdate{
			ContentDisposition: `filename="` + filename + `"`,
		})
		if err != nil {
			log.Printf("Writing filename %q to object %q in bucket %q: %v", filename, objectName, s.StorageBucket, err)
			http.Error(w, "Could not write file to GCS :(", http.StatusInternalServerError)
			return
		}
	}
	log.Printf("Written object %q in bucket %q: %v", objectName, s.StorageBucket)
	getURL := s.BackendBaseURL + "/get/" + fileUUID

	tsched := time.Now()
	_, err = s.ScheduleForgetFile(r.Context(), fileUUID)
	if err != nil {
		log.Printf("Scheduling deletion of object %q in bucket %q: %v", objectName, s.StorageBucket, err)
	}
	log.Printf("Scheduled deletion of object %q in bucket %q", objectName, s.StorageBucket)

	//
	// Provide the access URL to the client in QR code + in clear text
	//

	qr, err := qrcode.New(getURL, qrcode.Medium)
	if err != nil {
		log.Println("Generating QR code:", err)
		http.Error(w, "Could not generate QR code :(", http.StatusInternalServerError)
		return
	}
	qrArt := qr2CharBlocks(qr)

	fmt.Fprintln(w, qrArt)
	fmt.Fprintf(w, "\nPlease find your file at %s\n", getURL)

	tDelete := tsched.Add(s.StorageFileTTL)
	hDelete := tDelete.Format("15:04 MST")
	fmt.Fprintf(w, "\nIt is valid until %s\n", hDelete)
}

func (s *Server) usage() string {
	return fmt.Sprintf(`
    Welcome to Hot Maze in command-line mode

    Usage:
        curl --data-binary @my_file.png ` + s.BackendBaseURL + `/term
	`)
}

// QRCode object -> ASCII-art string image of the QR code.
// Makes sense only with fixed-width fonts.
// The filled block character '█' is technically Unicode, not ASCII.
func qr2CharBlocks(qr *qrcode.QRCode) string {
	bits := qr.Bitmap()
	var buf bytes.Buffer
	for y := range bits {
		buf.WriteString(" ")
		for x := range bits[y] {
			if bits[y][x] {
				// Blank square
				buf.WriteString("  ")
			} else {
				// Filled square
				buf.WriteString("██")
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}
