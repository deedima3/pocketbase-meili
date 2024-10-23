package meili

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

var indexableCollections = []string{"users", "posts", "jobs"}
var client = meilisearch.New(os.Getenv("MEILI_HOST"), meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")))

func RegisterMeiliHooks(app *pocketbase.PocketBase) {
	app.OnRecordAfterCreateRequest(indexableCollections...).Add(func(e *core.RecordCreateEvent) error {
		go createOrUpdateMeiliDocument(e.Record, app)
		return nil
	})
	app.OnRecordAfterUpdateRequest(indexableCollections...).Add(func(e *core.RecordUpdateEvent) error {
		go createOrUpdateMeiliDocument(e.Record, app)
		return nil
	})
	app.OnRecordAfterDeleteRequest(indexableCollections...).Add(func(e *core.RecordDeleteEvent) error {
		go deleteMeiliDocument(e.Record, app)
		return nil
	})
}

func createOrUpdateMeiliDocument(record *models.Record, app *pocketbase.PocketBase) {
	data, err := record.MarshalJSON()
	if err != nil {
		app.Logger().Error(err.Error())
		return
	}
	index := client.Index(record.Collection().Name)
	var meiliData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &meiliData); err != nil {
		app.Logger().Error(err.Error())
		return
	}
	task, err := index.AddDocuments(meiliData, "id")
	if err != nil {
		app.Logger().Error(err.Error())
		app.Logger().Error("Failed updating data in meilisearch. Task UID: " + strconv.FormatInt(task.TaskUID, 10))
		return
	}
}
func deleteMeiliDocument(record *models.Record, app *pocketbase.PocketBase) {
	index := client.Index(record.Collection().Name)
	task, err := index.DeleteDocument(record.Id)
	if err != nil {
		app.Logger().Error(err.Error())
		app.Logger().Error("Failed deleting data from meilisearch. Task UID: " + strconv.FormatInt(task.TaskUID, 10))
		return
	}
}
