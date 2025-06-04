package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/Nilesh2000/conduit/internal/config"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Seeder struct {
	db *sql.DB
}

func main() {
	log.Println("Starting database seeding...")

	// Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Database
	db, err := sql.Open("postgres", cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Ping database to check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	seeder := Seeder{db: db}
	if err := seeder.Seed(); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
}

// Seed seeds the database with sample data.
func (s *Seeder) Seed() error {
	if err := s.clearTables(); err != nil {
		return fmt.Errorf("failed to clear tables: %w", err)
	}

	if err := s.seedUsers(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	if err := s.seedTags(); err != nil {
		return fmt.Errorf("failed to seed tags: %w", err)
	}

	if err := s.seedArticles(); err != nil {
		return fmt.Errorf("failed to seed articles: %w", err)
	}

	if err := s.seedComments(); err != nil {
		return fmt.Errorf("failed to seed comments: %w", err)
	}

	if err := s.seedFollows(); err != nil {
		return fmt.Errorf("failed to seed follows: %w", err)
	}

	if err := s.seedFavorites(); err != nil {
		return fmt.Errorf("failed to seed favorites: %w", err)
	}

	log.Println("Database seeding completed successfully!")
	s.printSummary()

	return nil
}

// clearTables clears all tables in the database.
func (s *Seeder) clearTables() error {
	log.Println("Clearing tables...")

	tables := []string{
		"article_tags",
		"favorites",
		"follows",
		"comments",
		"articles",
		"tags",
		"users",
	}

	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
	}

	sequences := []string{
		"users_id_seq",
		"articles_id_seq",
		"tags_id_seq",
		"comments_id_seq",
	}

	// Reset sequences
	for _, sequence := range sequences {
		_, err := s.db.Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", sequence))
		if err != nil {
			return fmt.Errorf("failed to reset sequence for table %s: %w", sequence, err)
		}
	}

	return nil
}

// seedUsers seeds the users table with sample data.
func (s *Seeder) seedUsers() error {
	log.Println("Seeding users...")
	users := []struct {
		username, email, password, bio, image string
	}{
		{
			username: "johndoe",
			email:    "john@example.com",
			password: "password123",
			bio:      "Full-stack developer passionate about clean code and great user experiences.",
			image:    "https://api.dicebear.com/7.x/avataaars/svg?seed=johndoe",
		},
		{
			username: "janedoe",
			email:    "jane@example.com",
			password: "password123",
			bio:      "Tech writer and blogger. Love sharing knowledge about software development.",
			image:    "https://api.dicebear.com/7.x/avataaars/svg?seed=janedoe",
		},
		{
			username: "bobsmith",
			email:    "bob@example.com",
			password: "password123",
			bio:      "Senior software engineer with 10+ years of experience in Go and distributed systems.",
			image:    "https://api.dicebear.com/7.x/avataaars/svg?seed=bobsmith",
		},
		{
			username: "alicejohnson",
			email:    "alice@example.com",
			password: "password123",
			bio:      "Product manager turned developer. Bridging the gap between business and technology.",
			image:    "https://api.dicebear.com/7.x/avataaars/svg?seed=alicejohnson",
		},
		{
			username: "newuser",
			email:    "newuser@example.com",
			password: "password123",
			bio:      "",
			image:    "",
		},
	}

	for _, user := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(user.password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", user.username, err)
		}

		_, err = s.db.Exec(`
			INSERT INTO users (username, email, password_hash, bio, image)
			VALUES ($1, $2, $3, $4, $5)`,
			user.username, user.email, string(hashedPassword), user.bio, user.image,
		)
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %w", user.username, err)
		}
	}

	log.Printf("✓ Created %d users", len(users))
	return nil
}

// seedTags seeds the tags table with sample data.
func (s *Seeder) seedTags() error {
	log.Println("Seeding tags...")

	tags := []string{
		"golang", "javascript", "react", "nodejs", "python",
		"docker", "kubernetes", "aws", "microservices", "database",
		"testing", "devops", "frontend", "backend", "fullstack",
		"tutorial", "beginners", "advanced", "career", "production",
	}

	for _, tag := range tags {
		_, err := s.db.Exec(`INSERT INTO tags (name) VALUES ($1)`, tag)
		if err != nil {
			return fmt.Errorf("failed to insert tag %s: %w", tag, err)
		}
	}

	log.Printf("✓ Created %d tags", len(tags))
	return nil
}

// seedArticles seeds the articles table with sample data.
func (s *Seeder) seedArticles() error {
	log.Println("Seeding articles...")

	articles := []struct {
		slug, title, description, body string
		authorID                       int
		tags                           []string
	}{
		{
			slug:        "getting-started-with-golang",
			title:       "Getting Started with Golang",
			description: "A comprehensive guide to start your journey with Go programming language",
			body: `# Getting Started with Golang

Go, also known as Golang, is a programming language developed by Google. It's designed to be simple, efficient, and reliable.

## Why Choose Go?

- **Simple syntax**: Easy to learn and read
- **Fast compilation**: Quick build times
- **Built-in concurrency**: Goroutines and channels
- **Strong standard library**: Batteries included
- **Cross-platform**: Write once, run anywhere

## Your First Go Program

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

This simple program demonstrates the basic structure of a Go application.

## Next Steps

1. Install Go from the official website
2. Set up your development environment
3. Work through the Go Tour
4. Build your first project

Happy coding!`,
			authorID: 1,
			tags:     []string{"golang", "tutorial", "beginners"},
		},
		{
			slug:        "advanced-golang-patterns",
			title:       "Advanced Golang Design Patterns",
			description: "Exploring advanced design patterns and best practices in Go",
			body: `# Advanced Golang Design Patterns

As you grow more experienced with Go, understanding design patterns becomes crucial for writing maintainable code.

## The Repository Pattern

The repository pattern helps abstract data access logic:

` + "```go" + `
type UserRepository interface {
    GetByID(id int) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id int) error
}
` + "```" + `

## Dependency Injection

Dependency injection makes your code more testable and flexible.

## Error Handling Patterns

Go's explicit error handling is one of its strengths when used correctly.

These patterns will help you write more professional Go applications.`,
			authorID: 3,
			tags:     []string{"golang", "advanced"},
		},
		{
			slug:        "building-rest-apis-with-go",
			title:       "Building REST APIs with Go",
			description: "Learn how to build robust REST APIs using Go's standard library",
			body: `# Building REST APIs with Go

Go's standard library provides excellent support for building HTTP APIs without external frameworks.

## Creating Your First Handler

` + "```go" + `
func handleUsers(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // Handle GET
    case http.MethodPost:
        // Handle POST
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
` + "```" + `

## Middleware

Middleware functions help you add cross-cutting concerns like logging, authentication, and CORS.

## Testing Your API

Don't forget to test your API endpoints thoroughly.

Building APIs with Go is both powerful and enjoyable!`,
			authorID: 1,
			tags:     []string{"golang", "backend", "tutorial"},
		},
		{
			slug:        "react-hooks-deep-dive",
			title:       "React Hooks: A Deep Dive",
			description: "Understanding React Hooks and how they revolutionize component logic",
			body: `# React Hooks: A Deep Dive

React Hooks have transformed how we write React components, making functional components as powerful as class components.

## useState Hook

The most basic hook for managing component state:

` + "```javascript" + `
const [count, setCount] = useState(0);
` + "```" + `

## useEffect Hook

Handling side effects in functional components:

` + "```javascript" + `
useEffect(() => {
    document.title = ` + "`Count: ${count}`" + `;
}, [count]);
` + "```" + `

## Custom Hooks

Creating reusable stateful logic:

` + "```javascript" + `
function useCounter(initialValue = 0) {
    const [count, setCount] = useState(initialValue);

    const increment = () => setCount(count + 1);
    const decrement = () => setCount(count - 1);

    return { count, increment, decrement };
}
` + "```" + `

Hooks make React development more intuitive and powerful.`,
			authorID: 2,
			tags:     []string{"react", "javascript", "frontend"},
		},
		{
			slug:        "docker-best-practices",
			title:       "Docker Best Practices for Production",
			description: "Essential Docker practices for deploying applications in production",
			body: `# Docker Best Practices for Production

Docker has revolutionized how we deploy applications, but production deployments require careful consideration.

## Multi-stage Builds

Use multi-stage builds to reduce image size:

` + "```dockerfile" + `
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
` + "```" + `

## Security Considerations

- Don't run as root user
- Use specific image tags, not 'latest'
- Scan images for vulnerabilities
- Use secrets management

## Performance Tips

- Optimize layer caching
- Use .dockerignore files
- Keep images small
- Monitor resource usage

Following these practices will help ensure your containerized applications run smoothly in production.`,
			authorID: 4,
			tags:     []string{"docker", "devops", "production"},
		},
		{
			slug:        "my-coding-journey",
			title:       "My Coding Journey: From Beginner to Senior Developer",
			description: "Reflecting on my path from coding newbie to experienced developer",
			body: `# My Coding Journey: From Beginner to Senior Developer

Looking back at my coding journey, I'm amazed at how much I've learned and grown over the years.

## The Beginning

I started with HTML and CSS, building simple websites. Everything seemed magical back then.

## Learning Programming Logic

JavaScript was my first real programming language. The concepts of variables, functions, and loops were mind-bending at first.

## Building Real Projects

The turning point was when I started building actual projects instead of just following tutorials.

## Key Lessons Learned

1. **Consistency beats intensity**: Daily practice is better than weekend marathons
2. **Build projects**: Nothing teaches like real-world application
3. **Learn from others**: Code reviews and mentorship are invaluable
4. **Stay curious**: Technology evolves rapidly
5. **Focus on fundamentals**: Algorithms and data structures matter

## Advice for Beginners

- Don't get overwhelmed by the vast amount to learn
- Pick one technology and get good at it first
- Join developer communities
- Contribute to open source
- Never stop learning

The journey is challenging but incredibly rewarding. Keep coding!`,
			authorID: 2,
			tags:     []string{"career", "beginners"},
		},
	}

	for _, article := range articles {
		var articleID int
		err := s.db.QueryRow(`
			INSERT INTO articles (slug, title, description, body, author_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			article.slug, article.title, article.description, article.body, article.authorID,
		).Scan(&articleID)
		if err != nil {
			return fmt.Errorf("failed to insert article %s: %w", article.title, err)
		}

		// Add tags to article
		for _, tagName := range article.tags {
			var tagID int
			err := s.db.QueryRow(`SELECT id FROM tags WHERE name = $1`, tagName).Scan(&tagID)
			if err != nil {
				return fmt.Errorf("failed to find tag %s: %w", tagName, err)
			}

			_, err = s.db.Exec(`
				INSERT INTO article_tags (article_id, tag_id)
				VALUES ($1, $2)`,
				articleID, tagID)
			if err != nil {
				return fmt.Errorf(
					"failed to link article %d with tag %d: %w",
					articleID,
					tagID,
					err,
				)
			}
		}
	}

	log.Printf("✓ Created %d articles with tags", len(articles))
	return nil
}

// seedComments seeds the comments table with sample data.
func (s *Seeder) seedComments() error {
	log.Println("Seeding comments...")

	comments := []struct {
		body      string
		articleID int
		authorID  int
	}{
		{
			body:      "Great introduction to Go! This really helped me understand the basics.",
			articleID: 1,
			authorID:  2,
		},
		{
			body:      "Thanks for sharing this. The code examples are very clear.",
			articleID: 1,
			authorID:  3,
		},
		{
			body:      "I've been looking for exactly this kind of tutorial. Bookmarked!",
			articleID: 1,
			authorID:  4,
		},
		{
			body:      "The repository pattern section is excellent. Could you elaborate more on testing strategies?",
			articleID: 2,
			authorID:  1,
		},
		{
			body:      "Advanced patterns can be tricky, but your explanations make them accessible.",
			articleID: 2,
			authorID:  2,
		},
		{
			body:      "Very practical approach to building APIs. I love that you focus on the standard library.",
			articleID: 3,
			authorID:  4,
		},
		{
			body:      "Hooks definitely changed how I think about React components. Great deep dive!",
			articleID: 4,
			authorID:  3,
		},
		{
			body:      "The custom hooks example is particularly useful. Thanks for sharing!",
			articleID: 4,
			authorID:  1,
		},
		{
			body:      "Multi-stage builds are a game changer for production deployments.",
			articleID: 5,
			authorID:  1,
		},
		{
			body:      "Your journey is inspiring! It's great to see how persistence pays off.",
			articleID: 6,
			authorID:  3,
		},
	}

	for _, comment := range comments {
		_, err := s.db.Exec(`
			INSERT INTO comments (body, article_id, user_id)
			VALUES ($1, $2, $3)`,
			comment.body, comment.articleID, comment.authorID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert comment: %w", err)
		}
	}

	log.Printf("✓ Created %d comments", len(comments))
	return nil
}

// seedFollows seeds the follows table with sample data.
func (s *Seeder) seedFollows() error {
	log.Println("Seeding follows...")

	follows := []struct {
		followerID  int
		followingID int
	}{
		{1, 2}, // johndoe follows janedoe
		{1, 3}, // johndoe follows bobsmith
		{2, 1}, // janedoe follows johndoe
		{2, 4}, // janedoe follows alicejohnson
		{3, 1}, // bobsmith follows johndoe
		{3, 2}, // bobsmith follows janedoe
		{4, 2}, // alicejohnson follows janedoe
		{4, 3}, // alicejohnson follows bobsmith
		{5, 1}, // newuser follows johndoe
	}

	for _, follow := range follows {
		_, err := s.db.Exec(`
			INSERT INTO follows (follower_id, following_id)
			VALUES ($1, $2)`,
			follow.followerID, follow.followingID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert follow relationship: %w", err)
		}
	}

	log.Printf("✓ Created %d follow relationships", len(follows))
	return nil
}

// seedFavorites seeds the favorites table with sample data.
func (s *Seeder) seedFavorites() error {
	log.Println("Seeding favorites...")

	favorites := []struct {
		userID    int
		articleID int
	}{
		{1, 4}, // johndoe favorites React article
		{1, 5}, // johndoe favorites Docker article
		{2, 1}, // janedoe favorites her own Go tutorial
		{2, 2}, // janedoe favorites advanced Go patterns
		{3, 1}, // bobsmith favorites Go tutorial
		{3, 3}, // bobsmith favorites REST API article
		{4, 1}, // alicejohnson favorites Go tutorial
		{4, 6}, // alicejohnson favorites coding journey
		{5, 1}, // newuser favorites Go tutorial
		{5, 6}, // newuser favorites coding journey
	}

	for _, favorite := range favorites {
		_, err := s.db.Exec(`
			INSERT INTO favorites (user_id, article_id)
			VALUES ($1, $2)`,
			favorite.userID, favorite.articleID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert favorite: %w", err)
		}
	}

	log.Printf("✓ Created %d favorites", len(favorites))
	return nil
}

// printSummary prints a summary of the seeded data.
func (s *Seeder) printSummary() {
	log.Println("\n" + strings.Repeat("=", 50))
	log.Println("SEED DATA SUMMARY")
	log.Println(strings.Repeat("=", 50))

	// Count users
	var userCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	log.Printf("Users: %d", userCount)

	// Count articles
	var articleCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&articleCount)
	if err != nil {
		log.Fatalf("Failed to count articles: %v", err)
	}
	log.Printf("Articles: %d", articleCount)

	// Count comments
	var commentCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&commentCount)
	if err != nil {
		log.Fatalf("Failed to count comments: %v", err)
	}
	log.Printf("Comments: %d", commentCount)

	// Count tags
	var tagCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM tags").Scan(&tagCount)
	if err != nil {
		log.Fatalf("Failed to count tags: %v", err)
	}
	log.Printf("Tags: %d", tagCount)

	// Count follows
	var followCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM follows").Scan(&followCount)
	if err != nil {
		log.Fatalf("Failed to count follows: %v", err)
	}
	log.Printf("Follow relationships: %d", followCount)

	// Count favorites
	var favoriteCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM favorites").Scan(&favoriteCount)
	if err != nil {
		log.Fatalf("Failed to count favorites: %v", err)
	}
	log.Printf("Favorites: %d", favoriteCount)

	log.Println(strings.Repeat("=", 50))
	log.Println("LOGIN CREDENTIALS (all passwords: 'password123'):")
	log.Println("- john@example.com (johndoe)")
	log.Println("- jane@example.com (janedoe)")
	log.Println("- bob@example.com (bobsmith)")
	log.Println("- alice@example.com (alicejohnson)")
	log.Println("- newuser@example.com (newuser)")
	log.Println(strings.Repeat("=", 50))
}
