package migrations

import (
	"log"

	"github.com/cesarbmathec/bets-backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
	log.Println("🚀 Iniciando migración de base de datos de apuestas...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Transaction{},
		&models.Payment{},
		&models.Event{},
		&models.EventCompetitor{},
		&models.Tournament{},
		&models.TournamentParticipant{},
		&models.UserPick{},
		&models.PickableSelection{},
		&models.Session{},
		&models.UserPaymentMethod{},
	)

	if err != nil {
		log.Fatal("❌ Error migrando tablas:", err)
	}

	// Crear usuario administrador inicial si no existe
	var admin models.User
	if err := db.Where("email = ?", "admin@betsystem.com").First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Crear admin
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Admin123!"), 14)
			admin = models.User{
				Username: "admin",
				Email:    "admin@betsystem.com",
				Password: string(hashedPassword),
				Role:     "admin",
				IsActive: true,
			}

			if err := db.Create(&admin).Error; err != nil {
				log.Printf("⚠️  Error creando usuario admin: %v", err)
			} else {
				log.Println("✅ Usuario administrador creado")
				log.Println("   Email: admin@betsystem.com")
				log.Println("   Password: Admin123!")
				log.Println("   Rol: admin")
			}

			// Crear wallet para el admin
			wallet := models.Wallet{
				UserID:        admin.ID,
				Balance:       0,
				FrozenBalance: 0,
				BonusBalance:  0,
				TokenBalance:  0,
				Currency:      "USD",
			}
			db.Create(&wallet)
		}
	}

	log.Println("✅ Base de datos actualizada y relacionada")
}
