Wisata API - Run Documentation

Prerequisites
- Golang (v1.18 or newer)
- MariaDB / MySQL server

Step 1: Database Setup
Ensure your MariaDB/MySQL service is running:
sudo systemctl start mariadb

Log in to the database as the root user:
sudo mysql -u root

Execute the following SQL commands to create the database:
CREATE DATABASE wisata_db;
GRANT ALL PRIVILEGES ON wisata_db.* TO 'yoda'@'localhost';
FLUSH PRIVILEGES;
exit;

Step 2: Project Configuration
Open main.go and ensure the DSN string matches your credentials:
dsn := "yoda:@tcp(127.0.0.1:3306)/wisata_db?charset=utf8mb4&parseTime=True&loc=Local"

Step 3: Install Dependencies
Navigate to the root directory of the project and run:
go mod tidy

Step 4: Run the Application
Start the server (GORM will auto-migrate tables):
go run .

Step 5: Testing the API
Example - Register a new user:
curl -X POST http://localhost:8080/auth/register \
-H "Content-Type: application/json" \
-d '{
  "name": "Admin Test",
  "email": "admin@example.com",
  "password": "password123",
  "confirmPassword": "password123"
}'
