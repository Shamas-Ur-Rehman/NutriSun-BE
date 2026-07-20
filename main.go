package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"Shamas/nutrisun/config"
	"Shamas/nutrisun/routes"
	"Shamas/nutrisun/utils"
)

var masterDB *gorm.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	masterDB = config.ConnectMasterDB()
	if masterDB == nil {
		log.Fatal("Master DB connection failed")
	}
	log.Println("PostgreSQL master DB connected successfully")

	if err := utils.AutoMigrateMasterDB(masterDB); err != nil {
		log.Fatalf("Master schema migration failed: %v", err)
	}
	if err := utils.AutoMigrateTenantDatabases(masterDB); err != nil {
		log.Fatalf("Tenant schema migration failed: %v", err)
	}
	// if err := utils.SeedRBAC(masterDB); err != nil {
	// 	log.Printf("Warning: Seeder failed: %v", err)
	// }

	r := routes.SetupRouter(masterDB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("starting server on :%s\n", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
