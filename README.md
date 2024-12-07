# CLIo
**CLIo** is your terminal's ultimate companion, 
user-friendly TUI (Text User Interface) application designed to help you explore, use, 
and understand the terminal commands effortlessly.

Whether you're a beginner trying to learn the command-line interface (CLI) or a seasoned pro looking for a handy way of storing your library, 
CLIo offers an intuitive and powerful way to master your shell.

![CLIo demo](./assets/demo.gif)

## Installation
### Install latest binary
The recommended approach is to use the installation script, which automatically handles the installation of **CLIo** including the requirements for your environment.
```sh
curl --proto '=https' --tlsv1.2 -LsSf https://github.com/lian-rr/clio/releases/latest/download/clio-installer.sh | sh
```
### From source

#### Prerequisites
- go `v.1.20+`

#### Install with Go
```sh
go install --tags "fts5" github.com/lian-rr/clio@latest
```


### Build from source
```sh
git clone https://github.com/lian-rr/clio
cd clio
go build --tags "fts5" .
```

## Features

- 📚 **Command Library**: Browse a comprehensive list of terminal commands with detailed descriptions.
- 📖 **Command Explanations**: Get beginner-friendly explanations of commands, powered by OpenAI.
- 🔍 **Search and Filter**: Quickly find commands by name, keyword, or functionality.

### Roadmap
- Export/Import command library.
- History of previous uses.
- Configure custom Explanation engines (e.g Ollama)
- Custom Themes
- Custom Keymaps

## Configuration
In case you want to customize some of **CLIo** options 
you can provide the necessary configuration in the config file. 

**CLIo** checks for the configuration in the `$HOME/.config/clio/clio.toml` file.

### Example config
```toml
# write debug logs.
debug = false
# base path for storing the application data, e.g. SQLite db.
pathOverride = ""

# explanation feature configuration.
[professor]
# used for enabling the explanation feature.
enabled = false
# type of processor. Supported values [openai]. required if professor is enable.
type = "openai"

# openAI config for the openai professor.
[professor.openai]
# OpenAI key. required
key = "key"
# Used if you want to customize the explanation prompt.
customPrompt = ""
# Url for the API.
url = ""
# OpenAI model.
model = ""
```

## Discloure
Until the version `v.1.0.0`, bugs are expected and backwards compatibility not promised.
