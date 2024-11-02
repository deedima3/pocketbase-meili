package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

var indexableCollections = os.Getenv("INDEXED_COLLECTION")
var splitIndexableCollections = strings.Split(indexableCollections, ",")
var host = os.Getenv("MEILI_HOST")
var master_key = os.Getenv("MEILI_MASTER_KEY")

var client = meilisearch.New(host, meilisearch.WithAPIKey(master_key))

func RegisterMeiliHooks(app *pocketbase.PocketBase) {
	app.OnRecordAfterCreateRequest(splitIndexableCollections...).Add(func(e *core.RecordCreateEvent) error {
		go createOrUpdateMeiliDocument(e.Record, app)
		return nil
	})
	app.OnRecordAfterUpdateRequest(splitIndexableCollections...).Add(func(e *core.RecordUpdateEvent) error {
		go createOrUpdateMeiliDocument(e.Record, app)
		return nil
	})
	app.OnRecordAfterDeleteRequest(splitIndexableCollections...).Add(func(e *core.RecordDeleteEvent) error {
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
		app.Logger().Error("Failed updating data in meilisearch. Task UID: " + strconv.FormatInt(task.TaskUID, 10))
		return
	}
}

func main() {
	fmt.Println("Host : ", host)
	fmt.Println("Master Key : ", master_key)
	app := pocketbase.New()

	// serves static files from the provided public dir (if exists)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))
		RegisterMeiliHooks(app)
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
