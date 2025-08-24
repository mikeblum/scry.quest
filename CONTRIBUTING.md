# Contributing to Scry.Quest 🔮

## 🚀 Getting Started 🚀

### Prerequisites 📋
- **Go** 1.25+
- **Git**

### Project Structure 🗂️
```
scry.quest/
├── srd/                   # The great library of D&D knowledge
│   ├── spells/            # Arcane formulae and divine miracles
│   ├── bestiary/          # Creatures both wondrous and terrifying  
│   ├── classes/           # Paths of power and specialization
│   ├── species/           # The many peoples of the realms
│   └── items/             # Tools, treasures, and magical artifacts
├── log/                   # Chronicles of system events
├── env/                   # Environmental configurations
└── Makefile               # Build incantations and development spells
```

## ⚙️ Configuration ⚙️

### Environment Variables 🌍

All environment variables are prefixed with `SCRY_` and sensible defaults, but you can customize behavior by setting these variables:

#### Logging Configuration 📋
- `SCRY_LOG_LEVEL` - Log level (debug, info, warn, error) [default: info]
- `SCRY_LOG_FORMAT` - Log format (json, text) [default: json]

#### Server Configuration 🌐
- `SCRY_PORT` - HTTP server port [default: 8080] 
- `SCRY_HOST` - Host to bind server [default: localhost]

### Environment File Setup 📄

1. Copy the provided `.env.template` file to `.env`:
   ```bash
   cp .env.template .env
   ```

2. Customize `.env` with your specific settings ‼️ note the `SCRY_` prefix ‼️

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
