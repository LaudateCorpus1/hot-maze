package hotmaze

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/storage"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s Server) HandlerForgetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusBadRequest)
		return
	}

	fileUUID := r.FormValue("uuid")
	if fileUUID == "" {
		http.Error(w, "please provide file uuid", http.StatusBadRequest)
		return
	}

	objectName := "transit/" + fileUUID
	log.Println("Forgetting file", objectName)
	err := s.StorageClient.Bucket(s.StorageBucket).Object(objectName).Delete(r.Context())
	switch {
	case err == storage.ErrObjectNotExist:
		log.Println("File", objectName, "did not exist")
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Server) ScheduleForgetFile(ctx context.Context, fileUUID string) (*taskspb.Task, error) {
	// Adapted from https://cloud.google.com/tasks/docs/creating-http-target-tasks#go

	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.NewClient: %v", err)
	}
	defer client.Close()

	req := &taskspb.CreateTaskRequest{
		Parent: s.CloudTasksQueuePath,
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        s.BackendBaseURL + "/forget?uuid=" + fileUUID,
				},
			},
			ScheduleTime: &timestamppb.Timestamp{
				Seconds: time.Now().Add(s.StorageFileTTL).Unix(),
			},
		},
	}

	createdTask, err := client.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CreateTask: %v", err)
	}

	return createdTask, nil
}
