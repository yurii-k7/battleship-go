# Contributing to Battleship Game

Thank you for your interest in contributing to the Battleship Game! This document provides guidelines and information for contributors.

## 🤝 How to Contribute

### Reporting Bugs

1. **Check existing issues** to avoid duplicates
2. **Use the bug report template** when creating new issues
3. **Provide detailed information**:
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, browser, versions)
   - Screenshots or logs if applicable

### Suggesting Features

1. **Check existing feature requests** first
2. **Use the feature request template**
3. **Explain the use case** and benefits
4. **Consider implementation complexity**

### Code Contributions

1. **Fork the repository**
2. **Create a feature branch** from `develop`
3. **Make your changes**
4. **Add tests** for new functionality
5. **Update documentation** if needed
6. **Submit a pull request**

## 🛠️ Development Setup

### Prerequisites

- Node.js 18+
- Go 1.21+
- Docker & Docker Compose
- Git

### Local Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/battleship-go.git
cd battleship-go

# Add upstream remote
git remote add upstream https://github.com/originalowner/battleship-go.git

# Setup development environment
make setup

# Start development environment
make up
```

## 📝 Coding Standards

### Go Backend

- Follow standard Go formatting (`gofmt`)
- Use `go vet` to check for issues
- Write meaningful variable and function names
- Add comments for exported functions
- Keep functions small and focused
- Handle errors appropriately

```go
// Good
func (g *GameService) CreateGame(playerID int) (*models.Game, error) {
    if playerID <= 0 {
        return nil, errors.New("invalid player ID")
    }
    // ... implementation
}

// Bad
func (g *GameService) cg(p int) (*models.Game, error) {
    // ... implementation without validation
}
```

### React Frontend

- Use TypeScript for type safety
- Follow ESLint and Prettier configurations
- Use functional components with hooks
- Keep components small and focused
- Use meaningful prop and variable names
- Add proper error handling

```tsx
// Good
interface GameBoardProps {
  board: CellState[][];
  onCellClick: (x: number, y: number) => void;
  disabled?: boolean;
}

const GameBoard: React.FC<GameBoardProps> = ({ board, onCellClick, disabled = false }) => {
  // ... implementation
};

// Bad
const GameBoard = (props: any) => {
  // ... implementation without types
};
```

### Database

- Use descriptive table and column names
- Add appropriate indexes
- Include foreign key constraints
- Write migration scripts for schema changes

## 🧪 Testing

### Backend Tests

```bash
# Run all tests
cd backend && go test ./...

# Run tests with coverage
go test -cover ./...

# Run race condition tests
go test -race ./...
```

### Frontend Tests

```bash
# Run all tests
cd frontend && npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

### Test Guidelines

- Write tests for new functionality
- Maintain or improve test coverage
- Use descriptive test names
- Test both success and error cases
- Mock external dependencies

## 📋 Pull Request Process

### Before Submitting

1. **Sync with upstream**
   ```bash
   git fetch upstream
   git checkout develop
   git merge upstream/develop
   ```

2. **Run tests**
   ```bash
   make test
   ```

3. **Check code formatting**
   ```bash
   # Backend
   cd backend && gofmt -s -l .
   
   # Frontend
   cd frontend && npm run lint
   ```

### PR Requirements

- **Clear description** of changes
- **Link to related issues**
- **Tests pass** in CI
- **Code review** by maintainers
- **Documentation updated** if needed

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] New tests added for new functionality

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

## 🏷️ Commit Messages

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

### Examples

```
feat(game): add ship placement validation

fix(auth): resolve JWT token expiration issue

docs(readme): update installation instructions

test(api): add integration tests for game endpoints
```

## 🔄 Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Workflow

1. **Create release branch** from `develop`
2. **Update version numbers**
3. **Update CHANGELOG.md**
4. **Create pull request** to `main`
5. **Tag release** after merge
6. **Deploy to production**

## 🎯 Areas for Contribution

### High Priority

- [ ] Mobile responsiveness improvements
- [ ] Game replay functionality
- [ ] Tournament system
- [ ] AI opponent
- [ ] Performance optimizations

### Medium Priority

- [ ] Additional ship types
- [ ] Custom game rules
- [ ] Player statistics dashboard
- [ ] Social features (friends, invites)
- [ ] Internationalization (i18n)

### Low Priority

- [ ] Themes and customization
- [ ] Sound effects
- [ ] Animations
- [ ] Admin panel
- [ ] Analytics integration

## 📞 Getting Help

- **Discord**: [Join our Discord server](https://discord.gg/battleship)
- **GitHub Discussions**: [Ask questions](https://github.com/yourusername/battleship-go/discussions)
- **Email**: contributors@battleship-game.com

## 📜 Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## 🙏 Recognition

Contributors will be recognized in:

- README.md contributors section
- Release notes
- Annual contributor highlights

Thank you for contributing to Battleship Game! 🚢⚓
