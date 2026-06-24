package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"ragpack/pkg/auth"
	"ragpack/pkg/config"
	"ragpack/pkg/meta"
)

func bootstrapAPIKey(ctx context.Context, ms meta.MetaStore, cfg config.Config) {
	count, err := ms.CountAPIKeys(ctx)
	if err != nil {
		log.Printf("warning: could not count API keys: %v", err)
		return
	}

	if count == 0 {
		plaintext, _, hint, err := auth.Generate()
		if err != nil {
			log.Fatalf("auth: generate master key: %v", err)
		}
		if _, err := ms.CreateAPIKey(ctx, "master", plaintext); err != nil {
			log.Fatalf("auth: store master key: %v", err)
		}
		keyFile := filepath.Join(cfg.DataPath, "api_key")
		if err := os.MkdirAll(cfg.DataPath, 0700); err != nil {
			log.Printf("warning: could not create data dir %s: %v", cfg.DataPath, err)
		} else if err := os.WriteFile(keyFile, []byte(plaintext+"\n"), 0600); err != nil {
			log.Printf("warning: could not write key file: %v", err)
		}
		log.Println("╔══════════════════════════════════════════════════════╗")
		log.Println("║              RagPack — API Key Created               ║")
		log.Println("╠══════════════════════════════════════════════════════╣")
		log.Printf("║  Key:   %-44s ║", plaintext)
		log.Printf("║  Hint:  rp_••••%-38s ║", hint)
		log.Printf("║  Saved: %-44s ║", keyFile)
		log.Println("╠══════════════════════════════════════════════════════╣")
		log.Println("║  Store this key — it will not be shown again.        ║")
		log.Println("║  Use it to authenticate external API clients.        ║")
		log.Println("╚══════════════════════════════════════════════════════╝")
		return
	}

	keys, err := ms.ListAPIKeys(ctx)
	if err != nil || len(keys) == 0 {
		return
	}
	log.Printf("API key active — hint: rp_••••%s (recover from %s/api_key if lost)", keys[0].KeyHint, cfg.DataPath)
}
