# 🔮 scry.quest 🔮

Welcome, brave adventurer, to **scry.quest** - an AI-powered dungeon master that scrys the depths of Dungeons and Dungeons & Dragons ™️ lore!

📖 This grimoire aids adventurers in their most perilous sessions, wielding the complete System Reference Document [SRD CC v5.2.1](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf) as its spellbook. 📚✨

## ⚡ What Magic Awaits ⚡

🎲 **AI Dungeon Master**: Your digital companion draws upon vast repositories of D&D 5e knowledge  
🏰 **Complete SRD Integration**: Access to spells, monsters, classes, and rules at lightning speed  
🗡️ **Session Support**: Real-time assistance for both players and DMs during gameplay  
📊 **JSON-Powered Database**: Structured data for spells, creatures, and game mechanics  
🧙‍♂️ **Go-Powered Backend**: Fast, reliable service worthy of the finest artificer  

*Like a cleric's prepared spells at dawn, everything you need is ready when adventure calls!*

## Contributing

## 📖 The Tome of Contents 📖

### Core Systems 🏛️
- **SRD Data**: Complete D&D 5e reference materials in JSON format
- **Spell Database**: Every cantrip to 9th level spell at your fingertips  
- **Bestiary**: Ancient dragons to humble goblins, all statted and ready
- **Character Creation**: Classes, species, backgrounds, and advancement rules
- **Game Mechanics**: Combat, exploration, and social interaction guidelines

### Project Structure 🗂️
```
scry.quest/
├── srd/                   # The great library of D&D knowledge
│   ├── spells/            # Arcane formulae and divine miracles
│   ├── beastiary/         # Creatures both wondrous and terrifying  
│   ├── classes/           # Paths of power and specialization
│   ├── species/           # The many peoples of the realms
│   └── items/             # Tools, treasures, and magical artifacts
├── log/                   # Chronicles of system events
├── env/                   # Environmental configurations
└── Makefile               # Build incantations and development spells
```

## ⚙️ Configuration ⚙️

### Environment Variables 🌍

All environment variables are prefixed with `SCRY_`. The application comes with sensible defaults, but you can customize behavior by setting these variables:

#### Logging Configuration 📋
- `SCRY_LOG_LEVEL` - Log level (debug, info, warn, error) [default: info]
- `SCRY_LOG_FORMAT` - Log format (json, text) [default: json]

#### Server Configuration 🌐
- `SCRY_PORT` - HTTP server port [default: 8080] 
- `SCRY_HOST` - Host to bind server [default: localhost]

### Environment File Setup 📄

1. Copy the provided `.env` file to `.env.local`:
   ```bash
   cp .env.template .env.local
   ```

2. Customize `.env.local` with your specific settings
3. The application will automatically load these variables with the `SCRY_` prefix

## 🎮 For Players & DMs 🎮

Whether you're a **seasoned adventurer** seeking quick rule clarifications or a **fledgling dungeon master** needing creature stats mid-session, scry.quest serves as your ever-present familiar. 🦉

*"Knowledge is the sharpest blade and the strongest shield."*

---

## 📜 Legal Scroll 📜

> This work includes material from the System Reference Document 5.2.1 ("SRD 5.2.1") by Wizards of the Coast LLC, available at [https://www.dndbeyond.com/srd](https://www.dndbeyond.com/srd). The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International License, available at [https://creativecommons.org/licenses/by/4.0/legalcode](https://creativecommons.org/licenses/by/4.0/legalcode).

---

*May your dice roll high and your adventures be legendary! 🎲⚔️*

