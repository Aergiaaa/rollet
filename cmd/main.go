package main

import (
	"database/sql"
	"log"

	"github.com/Aergiaaa/rollet/internal/database"
	"github.com/Aergiaaa/rollet/internal/env"
	"github.com/joho/godotenv"
)

type app struct {
	host      string
	port      int
	jwtSecret string
	models    database.Models
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

<<<<<<< HEAD
	db, err := sql.Open("postgres", env.GetEnvString("DATABASE_URL", ""))
	if err != nil {
		log.Fatalf("error opening database: %v", err)
=======
	url := env.GetEnvString("DATABASE_URL", "")
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err = database.RunMigrations(db); err != nil {
			log.Printf("Migration Failed: %v", err)
		}
		return
>>>>>>> 7a6b26f (fix: db close too soon)
	}
	defer db.Close()

	models := database.NewModels(db)
	app := &app{
		host: env.GetEnvString("HOST", "localhost"),
		port: env.GetEnvInt("PORT", 8080),
		jwtSecret: env.GetEnvString("JWT_SECRET",
			"apakah-apakah-bukan-ini-bukan-secret-kamu"),
		models: models,
	}

	if err := app.serve(); err != nil {
		log.Fatalf("error serving app: %v", err)
	}
}
