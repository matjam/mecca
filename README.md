# MECCA Language Reference

> **Note:** Parts of this document are based on the original MECCA documentation by Scott J. Dudley, and have been adapted for the Go interpretation of the MECCA language. The original documentation can be found in [max300.txt](./max300.txt) in this repository.

## Introduction

MECCA is a simple, easy-to-learn templating language originally designed for the Maximus BBS software by Scott J. Dudley. It was created to allow non-programmers to create interactive content for BBS systems.

This Go implementation provides a subset of the original MECCA language, adapted for modern terminal-based applications. It's designed to work with the [bubbletea](https://github.com/charmbracelet/bubbletea) framework and uses the [lipgloss](https://github.com/charmbracelet/lipgloss) library for rendering, enabling beautiful terminal user interfaces in modern Go BBS software.

## Usage

### Basic Usage

The simplest way to use the MECCA interpreter is to create a new interpreter and process a string:

```go
package main

import (
	"github.com/matjam/mecca"
)

func main() {
	// Create a new interpreter with default settings
	interpreter := mecca.NewInterpreter()
	
	// Process and display a MECCA template
	interpreter.ExecString("[bold][red]Hello, World![reset]", nil)
}
```

### Using Variables

You can pass variables at execution time to substitute values in templates:

```go
interpreter := mecca.NewInterpreter()

template := "Hello, [name]! You have [count] new messages."

interpreter.ExecString(template, map[string]any{
	"name":  "Alice",
	"count": 42,
})
```

### Registering Custom Tokens

You can register custom tokens that will be called when encountered in templates. This allows you to create dynamic content based on your application's state.

Here's a complete example showing how to register and use custom tokens:

```go
package main

import (
	"fmt"
	"strings"

	"github.com/matjam/mecca"
)

// ServerContext holds application state
type ServerContext struct {
	user     string
	msgCount int
}

// userToken returns the current user's name
func (ctx ServerContext) userToken(_ []string) string {
	return ctx.user
}

// msgCountToken returns the number of new messages
func (ctx ServerContext) msgCountToken(args []string) string {
	return fmt.Sprintf("%d", ctx.msgCount)
}

// repeatToken repeats a string a specified number of times
func (ctx ServerContext) repeatToken(args []string) string {
	if len(args) < 2 {
		return ""
	}
	// Parse count from args[0]
	count := 0
	for _, c := range args[0] {
		count = count*10 + int(c-'0')
	}
	return strings.Repeat(args[1], count)
}

func main() {
	// Create application context
	ctx := ServerContext{
		user:     "Bob",
		msgCount: 3,
	}

	// Create interpreter
	interpreter := mecca.NewInterpreter()

	// Register custom tokens
	interpreter.RegisterToken("user", ctx.userToken, 0)        // Takes 0 arguments
	interpreter.RegisterToken("msgcount", ctx.msgCountToken, 1) // Takes 1 argument (unused in this example)
	interpreter.RegisterToken("repeat", ctx.repeatToken, 2)     // Takes 2 arguments: count and string

	// Use the tokens in a template
	template := `[bold yellow]Welcome, [user]![reset]
You have [lightblue][msgcount][reset] new messages.

[bold][repeat 5 =][reset] Separator [repeat 5 =][reset]
This is a custom token demonstration.
`

	interpreter.ExecString(template, nil)
}
```

### Using Type Methods as Token Functions

As shown in the example above, you can use methods on types as token functions. This allows tokens to access application state naturally:

```go
type MyApp struct {
	currentUser string
}

func (app *MyApp) userToken(_ []string) string {
	return app.currentUser
}

func main() {
	app := &MyApp{currentUser: "Alice"}
	interpreter := mecca.NewInterpreter()
	
	// Register a method as a token function
	interpreter.RegisterToken("user", app.userToken, 0)
	
	interpreter.ExecString("Hello, [user]!", nil)
}
```

### Working with Files

You can also process template files:

```go
// Set a template root directory
interpreter := mecca.NewInterpreter(mecca.WithTemplateRoot("./templates"))

// Execute a template file (returns the rendered string)
output, err := interpreter.ExecTemplate("welcome.mec", map[string]any{
	"user": "Alice",
})
if err != nil {
	log.Fatal(err)
}

// Or render directly to the output writer
err = interpreter.RenderTemplate("welcome.mec", map[string]any{
	"user": "Alice",
})
```

## Language Reference

### Basic Syntax

MECCA templates are plain UTF-8 text files, typically with a `.mec` extension. The interpreter processes tokens enclosed in square brackets and renders everything else as literal text.

**Token Format:** Tokens are delimited by square brackets: `[token]`

- Tokens can contain commands and arguments: `[token arg1 arg2]`
- Multiple tokens can be combined: `[red bold]Hello[reset]`
- Arguments can be quoted if they contain spaces: `[greet "hello world"]`
- To include a literal `[` in your text, use `[[`: `[[Y,n]?` renders as `[Y,n]?`

**Token Rules:**

- **Case-insensitive:** `[user]`, `[USER]`, and `[UsEr]` are equivalent
- **Spaces are ignored:** `[  user  ]`, `[user]`, and `[    user]` are equivalent
- **Multiple tokens per bracket:** `[lightblue blink user]` is equivalent to `[lightblue][blink][user]`

**Special Character Codes:**

- ASCII codes: `[65]` renders as the character 'A'
- UTF-8 codes: `[U+00A9]` renders as the copyright symbol ©

If you need to include CP437 characters, convert them to UTF-8 first:

```bash
iconv -f CP437 -t UTF-8 input.mec > output.mec
```

### Variables

Variables can be passed at execution time to substitute values in templates. Variables are passed as a `map[string]any` where keys match token names.

```go
template := "Hello, [name]! You have [count] messages."
interpreter.ExecString(template, map[string]any{
	"name":  "Alice",
	"count": 42,
})
```

**Variable Priority:** Variables passed at execution time override registered tokens with the same name. This allows you to provide runtime values that take precedence over static token implementations.

### Color Tokens

MECCA supports multiple color modes for text coloring.

**Basic Colors (4-bit ANSI):**

Set foreground color: `[red]text[reset]`  
Set background color: `[red on white]text[reset]`

Available colors for both foreground and background:
- `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`

Additional foreground-only colors (light variants):
- `lightblack`, `lightred`, `lightgreen`, `lightyellow`, `lightblue`, `lightmagenta`, `lightcyan`, `lightwhite`

**256-Color Mode:**

Use a color number prefixed with `#`: `[#202]text[reset]` (orange)

**True Color (RGB):**

Use a 6-digit hex color code: `[#FF0000]text[reset]` (red)

**Background Colors:**

Set background using `on`: `[red on white]`, `[lightblue on #00008B]`

You can also use the `[ON <color>]` token to set background only: `[ON blue]text[reset]`

### Text Style Tokens

MECCA provides various tokens to control text appearance:

**Text Styles:**
- `[bold]` or `[bright]` - Bold text
- `[dim]` - Dim/faint text
- `[underline]` - Underlined text
- `[italic]` - Italic text
- `[strike]` - Strikethrough text
- `[reverse]` - Reverse foreground and background colors
- `[blink]` - Blinking text (use `[steady]` to cancel)
- `[reset]` - Reset all styling

**Style Management:**
- `[save]` - Save current color and style state
- `[load]` - Restore previously saved color and style state

### Cursor Control Tokens

MECCA provides tokens for cursor movement and screen control:

**Screen Clearing:**
- `[cls]` - Clear the entire screen
- `[cleos]` - Clear to end of screen
- `[cleol]` - Clear to end of line

**Cursor Movement:**
- `[up]` - Move cursor up one line
- `[down]` - Move cursor down one line
- `[left]` - Move cursor left one column
- `[right]` - Move cursor right one column
- `[cr]` - Carriage return (move to beginning of line)
- `[lf]` - Line feed (move to next line)
- `[locate <row> <column>]` - Move cursor to specific position (0-indexed)

**Cursor Position:**
- `[savecursor]` - Save current cursor position
- `[restorecursor]` - Restore saved cursor position

**Drawing:**
- `[line <length> <char>]` - Draw a line using the specified character (e.g., `[line 40 -]`)

### File Inclusion

MECCA supports including other files in templates:

```
[include filename.mec]
```

The included file path is resolved relative to the template root directory. Included files can contain any valid MECCA tokens, including other `[include]` directives (circular includes are detected and prevented).

**Other File Tokens:**

- `[ansi <filename>]` - Include an ANSI art file directly
- `[ansiconvert <filename> <charset>]` - Convert and include a file from a specific character encoding (e.g., `cp437`)

### Custom Tokens

You can extend MECCA by registering custom tokens that execute Go functions when encountered in templates.

**Token Function Signature:**

```go
type TokenFunc func(args []string) string
```

The function receives a slice of string arguments and returns the substitution string. For example, `[repeat 3 hello]` with `argCount: 2` would call the function with `args = ["3", "hello"]`.

**Registering Tokens:**

```go
interpreter.RegisterToken("name", function, argCount)
```

- **name**: Token name (case-insensitive)
- **function**: The function to call
- **argCount**: Number of expected arguments

**Example:**

```go
// Token with no arguments
interpreter.RegisterToken("user", func(args []string) string {
	return "John Doe"
}, 0)

// Token with one argument
interpreter.RegisterToken("greet", func(args []string) string {
	if len(args) > 0 {
		return "Hello, " + args[0] + "!"
	}
	return "Hello!"
}, 1)
```

**Important Notes:**

- Token names are case-insensitive
- Registered tokens can be overridden by variables passed at execution time
- Registering a token with an existing name will cause a panic
- Token functions should return plain strings; styling is applied by the interpreter
- If a token is called with fewer arguments than specified, the function receives an empty slice

## License

This implementation of the MECCA language is released under the MIT license. See the LICENSE file for more information.

## Acknowledgements

The original MECCA language was designed by Scott J. Dudley for the Maximus BBS software. This implementation is based on the original MECCA language and has been adapted for the Go programming language.

I thank Scott J. Dudley for creating the original MECCA language and for making it available to the public. When I was a teenager, I spent many hours creating interactive content for my BBS using the MECCA language, and I have fond memories of those days. I hope that this implementation will allow others to create similar content for modern BBS systems.
