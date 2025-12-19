cat > README.md << 'EOF'
# 🕷️ Secure Web Crawling Feed Aggregation Service

A production-ready news aggregator built with Go that crawls websites, deduplicates content, and provides a personalized feed API.

##  Features

-  **JWT Authentication** - Secure user authentication with role-based access
-  **Web Crawler** - Asynchronous HTML parsing and article extraction
-  **Content Deduplication** - SHA-256 hashing to prevent duplicate articles
-  **RESTful API** - Clean API with pagination support
-  **MongoDB Integration** - Optimized with indexes for fast queries
-  **Background Jobs** - Non-blocking crawl operations with goroutines
-  **Subscription Management** - Users can subscribe to multiple sources

##  Tech Stack

- **Language:** Go 1.21+
- **Framework:** Gin
- **Database:** MongoDB
- **Authentication:** JWT
- **Web Scraping:** goquery

##  Prerequisites

- Go 1.21 or higher
- MongoDB 4.4 or higher
- Git

##  Installation

1. **Clone the repository:**
```bash
git clone https://github.com/Bavithbabu/Secure-Web-Crawling-Feed-Aggregation-Service.git
cd Secure-Web-Crawling-Feed-Aggregation-Service
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Create `.env` file:**
```bash
PORT=9000
MONGODB_URL=mongodb+srv://username:password@cluster.mongodb.net/
SECRET_KEY=your_secret_key_here
```

4. **Run the application:**
```bash
go run main.go
```

Server starts at: `http://localhost:9000`

##  API Documentation

### Authentication

#### Signup
```http
POST /users/signup
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "password": "Password@123",
  "phone": "1234567890",
  "user_type": "USER"
}
```

#### Login
```http
POST /users/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "Password@123"
}
```

### Subscriptions (Protected)

#### Add Subscription
```http
POST /api/subscriptions
token: <your_jwt_token>
Content-Type: application/json

{
  "url": "https://news.ycombinator.com"
}
```

#### List Subscriptions
```http
GET /api/subscriptions
token: <your_jwt_token>
```

#### Remove Subscription
```http
DELETE /api/subscriptions/:id
token: <your_jwt_token>
```

### Crawling (Protected)

#### Crawl Specific Source
```http
POST /api/crawl/:subscription_id
token: <your_jwt_token>
```

#### Crawl All Sources
```http
POST /api/crawl/all
token: <your_jwt_token>
```

### Feed (Protected)

#### Get Feed
```http
GET /api/feed?page=1&limit=20
token: <your_jwt_token>
```

##  Project Structure

```
go-lang-jwt/
├── controllers/       # HTTP request handlers
├── database/         # MongoDB connection & indexes
├── helpers/          # JWT & hashing utilities
├── middleware/       # Authentication middleware
├── models/          # Data structures
├── routes/          # API route definitions
├── services/        # Business logic
├── main.go          # Application entry point
├── go.mod           # Go module dependencies
└── .env            # Environment variables
```

##  Database Collections

- **users** - User accounts with hashed passwords
- **sources** - Crawled website sources
- **subscriptions** - User-source mappings
- **articles** - Extracted and deduplicated articles

##  Security Features

- Bcrypt password hashing
- JWT token authentication
- Role-based access control (ADMIN/USER)
- Input validation
- MongoDB injection prevention

##  Performance Optimizations

- Database indexes on frequently queried fields
- Pagination for large datasets
- Background crawling with goroutines
- Content deduplication (SHA-256)
- Article limit per source (50 max)

##Supported Sites

Currently optimized for:
- Hacker News
- Lobsters
- Any site with standard HTML structure

##  Contributing

Pull requests are welcome! For major changes, please open an issue first.

##  License

MIT License

## Author

**Bavith Babu**
- GitHub: [@Bavithbabu](https://github.com/Bavithbabu)

##  Future Enhancements

- [ ] RSS feed parser
- [ ] Scheduled crawling with cron
- [ ] Email notifications
- [ ] Full-text search
- [ ] Docker containerization
- [ ] Rate limiting & caching

---

⭐ Star this repo if you found it helpful!
EOF
