package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aip2pnews.local/internal/latestapp"
)

var version = "v0.2.47-demo"

func main() {
	runtimePaths, err := latestapp.DefaultRuntimePaths()
	if err != nil {
		log.Fatal(err)
	}
	addr := flag.String("listen", "0.0.0.0:51818", "http listen address")
	store := flag.String("store", runtimePaths.StoreRoot, "AiP2P store root")
	project := flag.String("project", "aip2p.news", "project name to index")
	archive := flag.String("archive", runtimePaths.ArchiveRoot, "UTC+0 markdown mirror root")
	rules := flag.String("subscriptions", runtimePaths.RulesPath, "local subscription rules JSON")
	writerPolicy := flag.String("writer-policy", runtimePaths.WriterPolicyPath, "local writer policy JSON")
	netFile := flag.String("net", runtimePaths.NetPath, "network bootstrap config")
	syncModeFlag := flag.String("sync-mode", string(latestapp.SyncModeManaged), "sync mode: managed, external, or off")
	syncBinary := flag.String("sync-binary", runtimePaths.SyncBinPath, "managed sync binary path")
	syncStaleAfter := flag.Duration("sync-stale-after", 2*time.Minute, "restart managed sync worker after this stale interval")
	flag.Parse()

	app, err := latestapp.New(*store, *project, version, *archive, *rules, *writerPolicy, *netFile)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	syncMode, err := latestapp.ParseSyncMode(*syncModeFlag)
	if err != nil {
		log.Fatal(err)
	}
	var supervisor *latestapp.ManagedSyncSupervisor
	if syncMode == latestapp.SyncModeManaged {
		supervisor, err = latestapp.StartManagedSyncSupervisor(ctx, latestapp.ManagedSyncConfig{
			Runtime:    runtimePaths,
			BinaryPath: *syncBinary,
			StoreRoot:  *store,
			NetPath:    *netFile,
			RulesPath:  *rules,
			WriterPolicyPath: *writerPolicy,
			Trackers:   runtimePaths.TrackerPath,
			StaleAfter: *syncStaleAfter,
			Logf:       log.Printf,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer supervisor.Stop()
		log.Printf("managed sync mode enabled; aip2p-newsd will supervise %s", latestapp.ProjectSyncBinaryNameForLogs())
	}
	log.Printf("AiP2P News Public listening on http://%s", *addr)
	log.Printf("markdown archive mirror: %s", *archive)
	log.Printf("subscription rules: %s", *rules)
	log.Printf("writer policy: %s", *writerPolicy)
	log.Printf("network bootstrap config: %s", *netFile)
	if err := app.ListenAndServe(*addr); err != nil {
		log.Fatal(err)
	}
}
