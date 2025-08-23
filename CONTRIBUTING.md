# Contributing to Scry.Quest 🔮

## 🚀 Getting Started 🚀

### Prerequisites 📋
- **Go** 1.25+
- **Git**

### Installation Ritual 🔧

1. **Clone the Repository** 
   ```bash
   git clone https://github.com/mikeblum/scry.quest
   cd scry.quest
   ```

2. **Configure Environment Variables**
   ```bash
   cp .env .env.local  # Copy and customize your environment settings
   ```

### Development Commands 🛠️

| Command | Effect | Description |
|---------|--------|-------------|
| `make test` | 🧪 | Run all tests (like checking your spell components) |
| `make build` | ⚒️ | Build the application (forge your magical item) |
| `make lint` | 👁️ | Run linter checks (peer review from the archmages) |
| `make fmt` | ✨ | Format code (organize your spellbook) |
| `make tidy` | 📚 | Tidy modules (organize your dependencies) |
| `make clean` | 🧹 | Remove build artifacts (clear the workspace) |
| `make test-perf` | ⚡ | Run benchmark tests (measure your spells' power) |
| `make vuln` | 🛡️ | Scan for vulnerabilities (ward against dark magic) |
| `make pre-commit` | ✅ | Run all checks (complete ritual preparation) |

## 🎯 Contributing 🎯

### 📝 Commit Messages 📝
Use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary):
- `feat ✨:`: New feature
- `fix 🐛:`: Bug fix
- `docs 📚`: Documentation
- `test 🧪:`: Tests
- `chore 🔧:`: Maintenance

## 📋 PR Checklist 📋

- [ ] Run `make pre-commit` to verify tests, formatting, and deps are ✅
