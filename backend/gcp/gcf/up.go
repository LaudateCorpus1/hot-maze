package p

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	qrcode "github.com/skip2/go-qrcode"
)

const squareChar = "*"

// CurlUp handles a command-line user uploading a file using cURL.
// Let's offer a URL for this file and put it in a QR-code in
// the response headers.
func CurlUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Wrong method", r.Method)
		http.Error(w, "Only POST requests allowed", http.StatusBadRequest)
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/")
	log.Println("Filename =", filename)

	// A read file data in request body
	var b bytes.Buffer
	_, err := io.Copy(&b, r.Body)
	if err != nil {
		log.Println("Reading request body:", err)
		http.Error(w, "Technical problem :(", http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	hasher := md5.New()
	_, err = hasher.Write(b.Bytes())
	if err == nil {
		log.Println("Data md5 =", hex.EncodeToString(hasher.Sum(nil)))
	} else {
		log.Println("Error computing md5:", err)
	}

	// B store in somewhere: Memcache? Firestore? GCS?
	resourceID, err := store(b.Bytes(), filename)
	if err != nil {
		log.Println("Storing data:", err)
		http.Error(w, "Technical problem :(", http.StatusInternalServerError)
		return
	}
	log.Printf("resourceID=%q\n", resourceID)

	// C build URL for the new resource
	url := "https://us-central1-hot-maze.cloudfunctions.net/down/" + resourceID

	// D create QR-code
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		log.Println("Generating QR-code:", err)
		http.Error(w, "Technical problem :(", http.StatusInternalServerError)
		return
	}

	// E write it in response headers
	w.Header().Set("Resource", resourceID+" available at "+url)
	err = qrHeaders(qr, w)
	if err != nil {
		log.Println("writing QR headers:", err)
		http.Error(w, "Technical problem :(", http.StatusInternalServerError)
		return
	}

	// F Make sure the URL is valid no more than 10mn
	// TODO

	// Everything OK :)
	w.WriteHeader(http.StatusCreated)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func qrHeaders(qr *qrcode.QRCode, w http.ResponseWriter) error {
	w.Header().Set("Usage", "Please scan this QR-code with a mobile QR reader app")
	str := qr2String(qr)
	rows := strings.Split(str, "\n")
	for i, row := range rows {
		hFieldName := fmt.Sprintf("qr%02d", i)
		hFieldValue := row
		w.Header().Set(hFieldName, hFieldValue)
		// log.Println(hFieldName, hFieldValue)
	}
	return nil
}

func qr2String(qr *qrcode.QRCode) string {
	grid := qr.Bitmap()
	buf := bytes.NewBuffer(make([]byte, 4000))
	for i := range grid {
		for j := range grid[i] {
			if grid[i][j] {
				buf.WriteString("  ")
			} else {
				buf.WriteString(squareChar + squareChar)
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func store(data []byte, filename string) (id string, err error) {
	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT")
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return "", err
	}
	collection := client.Collection("Files")
	doc, _, err := collection.Add(ctx, map[string]interface{}{
		"Data":     data,
		"Filename": filename,
	})
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}
