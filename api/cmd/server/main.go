package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Arthur-Bamberg/nacao-assistidos/api/internal/httpapi"
	"github.com/Arthur-Bamberg/nacao-assistidos/api/internal/store"
	"github.com/Arthur-Bamberg/nacao-assistidos/api/migrations"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	email := getenv("SEED_OPERADOR_EMAIL", "operador@nacao.local")
	senha := getenv("SEED_OPERADOR_SENHA", "nacao-dev")

	var handler http.Handler
	st, err := store.Open(ctx, store.DatabaseURL())
	if err != nil {
		log.Printf("postgres indisponível (%v); usando armazenamento em memória", err)
		mem := store.NewMemory()
		if err := mem.SeedOperador(email, senha); err != nil {
			log.Fatal(err)
		}
		handler = httpapi.New(mem).Handler()
	} else {
		defer st.Close()
		if err := st.Migrate(ctx, migrations.Operadores()); err != nil {
			log.Fatal(err)
		}
		if err := st.Migrate(ctx, migrations.Cadastro()); err != nil {
			log.Fatal(err)
		}
		if err := st.Migrate(ctx, migrations.UnidadeOficinas()); err != nil {
			log.Fatal(err)
		}
		if err := st.Migrate(ctx, migrations.Nucleo()); err != nil {
			log.Fatal(err)
		}
		if err := st.Migrate(ctx, migrations.CatalogoCrud()); err != nil {
			log.Fatal(err)
		}
		if err := st.SeedOperador(ctx, email, senha); err != nil {
			log.Fatal(err)
		}
		handler = httpapi.New(st).Handler()
	}

	addr := getenv("API_ADDR", ":8080")
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("api em %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
