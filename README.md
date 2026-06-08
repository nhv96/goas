# GOAS
The AI Agen that actually *works* with Ollama local setup.

### Project structure recommended by Gemini (for myself to read)
```
goas/
├── .github/
│   └── workflows/
│       └── test.yml       # Automates testing on every GitHub push
├── cmd/
│   └── goas/
│       └── main.go        # The entry point (keeps main clean)
├── internal/
│   └── config/            # Application logic hidden from external packages
│       └── config.go
├── pkg/                   # Reusable code others could import (optional)
│   └── calculator/
│       └── calc.go
├── .gitignore
├── GNUmakefile            # Task runner for quick build/test/install
├── LICENSE                # Essential for open-source (e.g., MIT)
├── README.md              # Documentation on how to install and use
├── go.mod
└── go.sum
```

### What I am doing right now
This project initially for me to learn how AI assistants work, how models work and try to implement
for myself an assistant.

So the plan is just code and make it work first, have simple features (chat, history, tool call,...)
then improve the project structure later.