# dotenv

`dotenv` is a Go package for parsing `.env` files. It supports environment variable loading, variable substitution, and handling of quoted and multi-line values.

## Features

- **Read**: Load environment variables from `.env` files.
- **Parse**: Parse environment variables from a reader.
- **Load**: Load environment variables into the current environment.
- **Variable Substitution**: Supports variable substitution and default values.
- **Quoted and Multi-line Values**: Handles single and double-quoted strings, including multi-line strings.

## Installation

To include `dotenv` in your project, use Go modules:

```bash
go get github.com/codescalersinternships/dotenv-eyadhussein
```

## Usage

### Reading from a File/s

```go
package main

import (
    "fmt"
    "log"
    "github.com/codescalersinternships/dotenv-eyadhussein/dotenv"
)

func main() {
    vars, err := dotenv.Read(".env")
    if err != nil {
        log.Fatal(err)
    }

    for key, value := range vars {
        fmt.Printf("%s=%s\n", key, value)
    }
}
```

### Parsing from an `io.Reader`

```go
package main

import (
    "fmt"
    "os"
    "log"
    "github.com/codescalersinternships/dotenv-eyadhussein/dotenv"
)

func main() {
    file, err := os.Open(".env")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    vars, err := dotenv.Parse(file)
    if err != nil {
        log.Fatal(err)
    }

    for key, value := range vars {
        fmt.Printf("%s=%s\n", key, value)
    }
}
```

### Loading into the Environment

```go
package main

import (
    "log"
    "github.com/codescalersinternships/dotenv-eyadhussein/dotenv"
)

func main() {
    err := dotenv.Load(".env")
    if err != nil {
        log.Fatal(err)
    }

    // Environment variables are now available via os.Getenv
    // Example:
    // fmt.Println(os.Getenv("MY_VARIABLE"))
}
```

## Supported Syntax

- Basic Key-Value Pairs: KEY=value
- Quoted Values: Single (') or double (") quotes
- Multi-line Values: Triple single (''') or double (""") quotes
- Variable Substitution: ${VARIABLE}

## Linting

```bash
make lint
```

## Testing

```bash
make test
```
