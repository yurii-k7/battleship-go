# 🚢 Battleship Game

A modern, multiplayer battleship game built with React frontend, Go backend, and PostgreSQL database. Features real-time gameplay, chat, leaderboards, and comprehensive deployment options.

[![CI/CD Pipeline](https://github.com/yourusername/battleship-go/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/yourusername/battleship-go/actions)
[![Security Scan](https://github.com/yourusername/battleship-go/workflows/Security%20Scan/badge.svg)](https://github.com/yourusername/battleship-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/battleship-go)](https://goreportcard.com/report/github.com/yourusername/battleship-go)

## ✨ Features

- 🎮 **Multiplayer Gameplay**: Real-time battleship matches between players
- 🔐 **Authentication**: Secure JWT-based user registration and login
- 💬 **Live Chat**: Real-time messaging during games
- 🏆 **Leaderboard**: Competitive scoring system with global rankings
- 🌐 **Real-time Updates**: WebSocket-powered live game state synchronization
- 🐳 **Docker Support**: Complete containerization for easy development
- ⚡ **Serverless Deployment**: AWS Lambda functions for scalable production
- 🧪 **Comprehensive Testing**: Full test coverage for frontend and backend
- 🚀 **CI/CD Pipeline**: Automated testing, building, and deployment
- 📱 **Responsive Design**: Works seamlessly on desktop and mobile devices

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React SPA     │    │   Go Backend    │    │   PostgreSQL    │
│                 │    │                 │    │                 │
│ • Game UI       │◄──►│ • REST API      │◄──►│ • Game State    │
│ • Authentication│    │ • WebSocket     │    │ • User Data     │
│ • Real-time Chat│    │ • Game Logic    │    │ • Leaderboard   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 Project Structure

```
battleship-go/
├── 📁 frontend/              # React TypeScript application
│   ├── 📁 src/
│   │   ├── 📁 components/    # Reusable UI components
│   │   ├── 📁 pages/         # Page components
│   │   ├── 📁 hooks/         # Custom React hooks
│   │   ├── 📁 services/      # API and WebSocket services
│   │   ├── 📁 types/         # TypeScript type definitions
│   │   └── 📁 utils/         # Utility functions
│   ├── 📄 Dockerfile         # Production container
│   ├── 📄 Dockerfile.dev     # Development container
│   └── 📄 package.json       # Dependencies and scripts
├── 📁 backend/               # Go server application
│   ├── 📁 internal/
│   │   ├── 📁 api/           # HTTP handlers and routes
│   │   ├── 📁 auth/          # Authentication logic
│   │   ├── 📁 database/      # Database connection and migrations
│   │   ├── 📁 game/          # Game logic and rules
│   │   ├── 📁 models/        # Data models
│   │   └── 📁 websocket/     # Real-time communication
│   ├── 📄 main.go            # Application entry point
│   ├── 📄 Dockerfile         # Container configuration
│   └── 📄 go.mod             # Go dependencies
├── 📁 database/              # Database schemas and migrations
│   └── 📁 init/              # Initialization scripts
├── 📁 lambda/                # AWS Lambda deployment
│   ├── 📄 serverless.yml     # Serverless configuration
│   ├── 📄 main.go            # Lambda handler
│   └── 📄 deploy.sh          # Deployment script
├── 📁 .github/               # CI/CD workflows
│   └── 📁 workflows/         # GitHub Actions
├── 📄 docker-compose.yml     # Local development environment
├── 📄 Makefile              # Development commands
└── 📄 README.md             # This file
```

## 🚀 Quick Start

### Prerequisites

- **Node.js** 18+
- **Go** 1.21+
- **Docker** & Docker Compose
- **PostgreSQL** (for local development)
- **Make** (optional, for convenience commands)

### 🐳 Docker Development (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/battleship-go.git
   cd battleship-go
   ```

2. **Start all services**
   ```bash
   make up
   # or
   docker-compose up
   ```

3. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Database: localhost:5432

4. **Create your first account and start playing!**

### 🛠️ Manual Development Setup

1. **Setup environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start PostgreSQL**
   ```bash
   docker-compose up postgres -d
   ```

3. **Setup Backend**
   ```bash
   cd backend
   go mod tidy
   go run main.go
   ```

4. **Setup Frontend** (in another terminal)
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## 📖 Game Rules

### Ship Placement
- **Carrier**: 5 cells
- **Battleship**: 4 cells
- **Cruiser**: 3 cells
- **Submarine**: 3 cells
- **Destroyer**: 2 cells

Ships can be placed horizontally or vertically, but cannot overlap or go outside the 10x10 grid.

### Gameplay
1. Both players place their ships on their boards
2. Players take turns firing at opponent's board
3. Hit a ship: Get another turn
4. Miss: Turn passes to opponent
5. First to sink all opponent ships wins!

### Scoring
- **Win**: +100 points
- **Hit**: +10 points
- **Sink ship**: +20 bonus points
- **Perfect game** (no misses): +50 bonus points

## 🔧 Development Commands

```bash
# Start all services
make up

# Stop all services
make down

# View logs
make logs

# Run tests
make test

# Run backend tests only
make backend-test

# Run frontend tests only
make frontend-test

# Clean up Docker resources
make clean

# Reset database
make db-reset

# Setup development environment
make setup
```

## 🧪 Testing

### Backend Tests
```bash
cd backend
go test ./...
go test -race ./...  # Race condition detection
```

### Frontend Tests
```bash
cd frontend
npm test              # Run tests
npm run test:coverage # With coverage report
```

### Integration Tests
```bash
# Start test environment
docker-compose -f docker-compose.test.yml up

# Run integration tests
make test-integration
```

## 🚀 Deployment

### AWS Lambda (Production)

1. **Configure AWS credentials**
   ```bash
   aws configure
   ```

2. **Deploy to AWS**
   ```bash
   cd lambda
   ./deploy.sh prod
   ```

3. **Environment Variables**
   Set these in AWS Lambda:
   - `DATABASE_URL`: PostgreSQL connection string
   - `JWT_SECRET`: Secret key for JWT tokens

### Docker Production

1. **Build production images**
   ```bash
   docker build -t battleship-backend ./backend
   docker build -t battleship-frontend ./frontend
   ```

2. **Deploy with docker-compose**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

## 📚 API Documentation

### Authentication Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login user |

### Game Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/games` | Create new game |
| POST | `/api/games/:id/join` | Join existing game |
| GET | `/api/games` | Get user's games |
| GET | `/api/games/:id` | Get game details |
| POST | `/api/games/:id/ships` | Place ships |
| POST | `/api/games/:id/moves` | Make move |
| GET | `/api/games/:id/moves` | Get game moves |

### Chat Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/games/:id/chat` | Send chat message |
| GET | `/api/games/:id/chat` | Get chat messages |

### WebSocket Events

| Event | Description |
|-------|-------------|
| `connect` | Client connects to game |
| `disconnect` | Client disconnects |
| `chat` | Chat message sent |
| `move` | Game move made |
| `game_update` | Game state changed |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Workflow

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
4. **Add tests** for new functionality
5. **Run the test suite**
   ```bash
   make test
   ```
6. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
7. **Push to your fork**
   ```bash
   git push origin feature/amazing-feature
   ```
8. **Open a Pull Request**

### Code Style

- **Go**: Follow standard Go formatting (`gofmt`)
- **TypeScript/React**: Use ESLint and Prettier
- **Commits**: Use conventional commit messages

## 🐛 Troubleshooting

### Common Issues

**Database connection failed**
```bash
# Reset database
make db-reset
```

**Frontend won't start**
```bash
# Clear node modules and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install
```

**Backend compilation errors**
```bash
# Update Go modules
cd backend
go mod tidy
go mod download
```

**Docker issues**
```bash
# Clean up Docker resources
make clean
docker system prune -a
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [React](https://reactjs.org/) and [Vite](https://vitejs.dev/)
- Backend powered by [Go](https://golang.org/) and [Gin](https://gin-gonic.com/)
- Database: [PostgreSQL](https://www.postgresql.org/)
- Deployment: [AWS Lambda](https://aws.amazon.com/lambda/) and [Docker](https://www.docker.com/)
- Testing: [Vitest](https://vitest.dev/) and Go testing package

## 📞 Support

- 📧 Email: support@battleship-game.com
- 🐛 Issues: [GitHub Issues](https://github.com/yourusername/battleship-go/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/yourusername/battleship-go/discussions)

---

**Happy Gaming! ⚓🎮**
