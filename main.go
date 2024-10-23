package main

import (
	"log"
	"os"

	"pocketbase-meili/meili"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	// serves static files from the provided public dir (if exists)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))
		return nil
	})

	meili.RegisterMeiliHooks(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
