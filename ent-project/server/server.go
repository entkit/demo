package main

import (
	"context"
	"entgo.io/contrib/entgql"
	"fmt"
	demo "github.com/entkit/entkit-demo"
	example "github.com/entkit/entkit-demo/ent-project"
	"github.com/entkit/entkit-demo/ent-project/core/ent"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"go.uber.org/zap"

	_ "github.com/entkit/entkit-demo/ent-project/core/ent/runtime"
	_ "github.com/mattn/go-sqlite3"
)

var refineFs fs.FS

func init() {
	var err error
	refineFs, err = fs.Sub(demo.Refine, "refine-project/build")
	if err != nil {
		log.Fatal("failed to get ui fs", err)
	}
}
func main() {
	client, err := ent.Open(
		"sqlite3",
		"file:runtime/ent/demo.db?mode=rwc&cache=shared&_fk=1",
	)
	if err != nil {
		log.Fatal("opening ent client", zap.Error(err))
	}
	if err := client.Schema.Create(
		context.Background(),
		//migrate.WithGlobalUniqueID(true),
	); err != nil {
		log.Fatal("running schema migration", zap.Error(err))
	}

	// Up MUX server
	mux := http.NewServeMux()

	srv := handler.NewDefaultServer(example.NewSchema(client))
	srv.Use(entgql.Transactioner{TxOpener: client})
	//srv.Use(&debug.Tracer{})
	mux.HandleFunc("/playground", playground.Handler("Example", "/query"))
	mux.Handle("/query", ent.EntkitAuthMiddleware(srv))

	if os.Getenv("SKIP_EMBED_SERVER") != "true" {
		println("Embed Server...")
		mux.HandleFunc("/", handleStatic)
	}
	log.Println("starting server...")
	if err := http.ListenAndServe(":80", mux); err != nil {
		log.Println("server failed:", err)
	}
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	path := filepath.Clean(r.URL.Path)
	path = strings.TrimPrefix(path, "/")
	if !strings.HasPrefix(path, "static/") &&
		!strings.HasPrefix(path, "images/") &&
		path != "favicon.ico" &&
		path != "asset-manifest.json" &&
		path != "environment.json" {
		path = "index.html"
	}

	file, err := refineFs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("file", path, "not found:", err)
			http.NotFound(w, r)
			return
		}
		log.Println("file", path, "cannot be read:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(path))
	w.Header().Set("Content-Type", contentType)
	if strings.HasPrefix(path, "static/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}
	stat, err := file.Stat()
	if err == nil && stat.Size() > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	}

	n, _ := io.Copy(w, file)
	log.Println("file", path, "copied", n, "bytes")
}
