package hotmaze

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s Server) HandlerForgetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusBadRequest)
		return
	}

	if taskName := r.Header.Get("X-Appengine-Taskname"); taskName == "" {
		// Validate the request comes from Cloud Tasks.
		log.Println("Invalid Task: No X-Appengine-Taskname request header found")
		http.Error(w, "Bad Request - Invalid Task", http.StatusBadRequest)
		return
	}

	fileUUID := r.FormValue("uuid")
	if fileUUID == "" {
		http.Error(w, "please provide file uuid", http.StatusBadRequest)
		return
	}

	chunksStr := r.FormValue("chunks")
	if fileUUID == "" {
		http.Error(w, "please provide number of chunks", http.StatusBadRequest)
		return
	}
	chunks, err := strconv.Atoi(chunksStr)
	if err != nil {
		http.Error(w, "please provide proper number of chunks", http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("D1/%s_meta", fileUUID)
	log.Println("Deleting path", path, " from Firestore")
	_, errDelete := s.FirestoreClient.Doc(path).Delete(r.Context())
	if errDelete != nil {
		log.Println(errDelete)
		http.Error(w, "Problem deleting a resource meta from Firestore :(", http.StatusInternalServerError)
		return
	}

	for k := 0; k < chunks; k++ {
		path := fmt.Sprintf("D1/%s_chunk_%d", fileUUID, k)
		log.Println("Deleting path", path, " from Firestore")
		_, errDelete := s.FirestoreClient.Doc(path).Delete(r.Context())
		if errDelete != nil {
			log.Println(errDelete)
			http.Error(w, "Problem deleting a resource chunk from Firestore :(", http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) ScheduleForgetFile(ctx context.Context, fileUUID string, nbChunks int) (*taskspb.Task, error) {
	uri := fmt.Sprintf("/forget?uuid=%s&chunks=%d", fileUUID, nbChunks)

	// Adapted from https://cloud.google.com/tasks/docs/creating-appengine-tasks#go

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	req := &taskspb.CreateTaskRequest{
		Parent: s.CloudTasksQueuePath,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#AppEngineHttpRequest
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod:  taskspb.HttpMethod_POST,
					RelativeUri: uri,
					AppEngineRouting: &taskspb.AppEngineRouting{
						Version: "d1",
					},
				},
			},
			ScheduleTime: &timestamppb.Timestamp{
				Seconds: time.Now().Add(s.StorageFileTTL).Unix(),
			},
		},
	}

	createdTask, err := s.TasksClient.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CreateTask: %v", err)
	}

	return createdTask, nil
}
