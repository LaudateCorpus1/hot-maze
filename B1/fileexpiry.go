package hotmaze

import "net/http"

func (s Server) HandlerForgetFile(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("uuid")
	if uuid == "" {
		http.Error(w, "please provide file uuid", http.StatusBadRequest)
		return
	}

	objectName := "transit/" + uuid
	err := s.StorageClient.Bucket(s.StorageBucket).Object(objectName).Delete(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
