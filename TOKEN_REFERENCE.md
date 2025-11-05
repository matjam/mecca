# MECCA Token Reference

This document provides a comprehensive reference for all MECCA language tokens, organized by category. Each token includes a description and example usage.

## Table of Contents

- [Color & Style Tokens](#color--style-tokens)
- [Cursor Control Tokens](#cursor-control-tokens)
- [File Operations](#file-operations)
- [Menu System](#menu-system)
- [Interactive Input & Flow Control](#interactive-input--flow-control)
- [Advanced Interactive](#advanced-interactive)
- [Questionnaire System](#questionnaire-system)
- [Text Utilities](#text-utilities)
- [Core Features](#core-features)

---

## Color & Style Tokens

### Basic Colors

#### `[black]`, `[red]`, `[green]`, `[yellow]`, `[blue]`, `[magenta]`, `[cyan]`, `[white]`
Sets the foreground color to the specified color.

**Example:**
```
[red]This text is red![reset]
[blue]This text is blue![reset]
```

#### `[lightblack]`, `[lightred]`, `[lightgreen]`, `[lightyellow]`, `[lightblue]`, `[lightmagenta]`, `[lightcyan]`, `[lightwhite]`
Sets the foreground color to the light/bright variant.

**Example:**
```
[lightred]This is bright red![reset]
[lightblue]This is bright blue![reset]
```

### Hex Colors

#### `[#RRGGBB]` or `[#RGB]`
Sets the foreground color using a hex color code. Supports both 6-digit (`#FF0000`) and 3-digit (`#F00`) formats.

**Example:**
```
[#FF0000]Pure red text[reset]
[#00FF00]Pure green text[reset]
[#0000FF]Pure blue text[reset]
[#F0F]Short form magenta[reset]
```

### Background Colors

#### `[on <color>]`
Sets the background color. The color can be a named color or hex color.

**Example:**
```
[red on white]Red text on white background[reset]
[white on blue]White text on blue background[reset]
[#FF0000 on #0000FF]Red on blue[reset]
```

#### `[BG <color>]` / `[bg <color>]`
Sets only the background color, leaving the foreground unchanged.

**Example:**
```
[red][BG white]Red text on white background[reset]
[BG #FF0000]Background only[reset]
```

#### `[FG <color>]` / `[fg <color>]`
Sets only the foreground color, leaving the background unchanged.

**Example:**
```
[BG white][FG red]Red text on white background[reset]
[FG #00FF00]Foreground only[reset]
```

### Text Styles

#### `[bold]` / `[bright]`
Makes text bold/bright. `[bright]` is a synonym for `[bold]`.

**Example:**
```
[bold]This text is bold![reset]
[bright]This is also bold![reset]
```

#### `[dim]`
Makes text dim/faint.

**Example:**
```
[dim]This text is dim[reset]
```

#### `[underline]`
Underlines text.

**Example:**
```
[underline]This text is underlined[reset]
```

#### `[italic]`
Italicizes text.

**Example:**
```
[italic]This text is italicized[reset]
```

#### `[reverse]`
Reverses foreground and background colors.

**Example:**
```
[red on white][reverse]This reverses the colors[reset]
```

#### `[strike]`
Strikes through text.

**Example:**
```
[strike]This text is struck through[reset]
```

#### `[blink]`
Makes text blink.

**Example:**
```
[blink]This text blinks[reset]
```

#### `[steady]`
Cancels the blink attribute.

**Example:**
```
[blink]Blinking text[steady]Not blinking anymore[reset]
```

### Style Management

#### `[reset]`
Removes all styling and resets to default colors and styles. Also ends menu option capture.

**Example:**
```
[bold][red]Styled text[reset]Normal text
```

#### `[save]`
Saves the current color and style state.

**Example:**
```
[red][bold]Red bold text[save]
[blue]Blue text[load]Red bold text again[reset]
```

#### `[load]`
Restores the previously saved color and style state.

**Example:**
```
[red][bold][save]Red bold[blue]Blue[load]Red bold again[reset]
```

### Conditional Display

#### `[color]` / `[colour]`
Displays following text only if ANSI color support is enabled. Canadian spelling `[colour]` is also supported.

**Example:**
```
[color][red]This only shows if color is supported[endcolor]
```

#### `[nocolor]` / `[nocolour]`
Displays following text only if ANSI color support is disabled. Canadian spelling `[nocolour]` is also supported.

**Example:**
```
[nocolor]This only shows if color is NOT supported[endcolor]
```

#### `[endcolor]` / `[endcolour]`
Ends a color-conditional block started by `[color]` or `[nocolor]`.

**Example:**
```
[color][red]Colored text[endcolor]
[nocolor]Plain text[endcolor]
```

---

## Cursor Control Tokens

### Screen Clearing

#### `[cls]`
Clears the entire screen and moves the cursor to the top-left corner.

**Example:**
```
[cls]Screen is now clear
```

#### `[cleos]`
Clears from the current cursor position to the end of the screen.

**Example:**
```
Some text[cleos]
```

#### `[cleol]`
Clears from the current cursor position to the end of the current line.

**Example:**
```
Some text[cleol]
```

### Cursor Movement

#### `[up]`
Moves the cursor up one line.

**Example:**
```
Line 1
Line 2[up]Back to line 1
```

#### `[down]`
Moves the cursor down one line.

**Example:**
```
Line 1[down]Line 2
```

#### `[left]`
Moves the cursor left one column.

**Example:**
```
Text[left]Moved left
```

#### `[right]`
Moves the cursor right one column.

**Example:**
```
Text[right]Moved right
```

#### `[locate <row> <column>]`
Positions the cursor at the specified row and column (0-indexed).

**Example:**
```
[locate 5 10]Cursor is at row 5, column 10
[locate 0 0]Top-left corner
```

#### `[cr]`
Moves the cursor to the beginning of the current line (carriage return).

**Example:**
```
Line with text[cr]New text (overwrites previous)
```

#### `[lf]`
Moves the cursor to the next line (line feed).

**Example:**
```
Line 1[lf]Line 2
```

### Cursor Position Management

#### `[savecursor]`
Saves the current cursor position.

**Example:**
```
[savecursor]Cursor position saved
[locate 10 20]Moved away
[restorecursor]Back to saved position
```

#### `[restorecursor]`
Restores the previously saved cursor position.

**Example:**
```
[savecursor][locate 5 5][restorecursor]Back to saved position
```

### Line Drawing

#### `[line <length> <character>]`
Draws a line of the specified length using the specified character.

**Example:**
```
[line 40 =]
[line 20 -]
[line 10 *]
```

---

## File Operations

### File Inclusion

#### `[include <file>]`
Includes and parses a MECCA template file. The file is processed as MECCA code and can contain any tokens.

**Example:**
```
[include header.mec]
Main content
[include footer.mec]
```

#### `[ansi <file>]`
Includes a raw ANSI file without parsing. The file contents are inserted directly into the output.

**Example:**
```
[ansi logo.ans]
[ansi welcome.ans]
```

#### `[ansiconvert <file> <charset>]`
Converts a file from the specified character encoding to UTF-8 and includes it. Currently supports `cp437` (Code Page 437).

**Example:**
```
[ansiconvert old_ansi.ans cp437]
```

#### `[copy <file>]`
Copies file contents directly to output without any parsing. Unlike `[include]`, this does not process MECCA tokens in the file.

**Example:**
```
[copy readme.txt]
```

### File Chaining

#### `[display <file>]`
Displays the specified file and stops processing the current file. Control does not return to the calling file.

**Example:**
```
Some content
[display another.mec]
This line is never reached
```

#### `[link <file>]`
Displays the specified file and returns control to the calling file. Supports nesting up to 8 levels.

**Example:**
```
Before link
[link menu.mec]
After link (execution continues here)
```

#### `[on exit <file>]` / `[onexit <file>]`
Sets a file to be executed when the current file finishes processing (either normally or via `[quit]` or `[exit]`).

**Example:**
```
[on exit cleanup.mec]
Main content
[quit]cleanup.mec will be executed
```

---

## Menu System

### Menu Definition

#### `[menu]`
Starts a new menu, clearing any existing menu options.

**Example:**
```
[menu]
[option a]Option A
[option b]Option B
[menuwait]
```

#### `[option <id>]`
Marks the beginning of option text. The option ID must be a single alphanumeric character. Option text continues until `[reset]` is encountered.

**Example:**
```
[menu]
[option a]Add new user[reset]
[option b]Delete user[reset]
[option c]List users[reset]
[menuwait]
```

#### `[menuwait]`
Waits for user input and reads the selected menu option. The selected option ID is stored and can be used with `[choice]` or `[store]`.

**Example:**
```
[menu]
[option 1]First option[reset]
[option 2]Second option[reset]
[menuwait]
[choice 1]You selected option 1[choice 2]You selected option 2
```

---

## Interactive Input & Flow Control

### User Input

#### `[readln]` / `[readln <description>]`
Reads a line of input from the user. If a description is provided, it's stored in the questionnaire data with the response.

**Example:**
```
[readln]Enter your name:
[readln Name]What is your name?
```

#### `[enter]`
Waits for the user to press Enter before continuing.

**Example:**
```
Press ENTER to continue[enter]
Continuing...
```

### Flow Control

#### `[goto <label>]` / `[jump <label>]`
Jumps to the specified label. `[jump]` is a synonym for `[goto]`.

**Example:**
```
[goto start]
Some text
[/start]Start of loop
```

#### `[/<label>]` / `[label <label>]`
Defines a label that can be jumped to with `[goto]`. Labels can be defined using either format.

**Example:**
```
[/start]Beginning of content
[goto start]
[/menu]Menu section
[goto menu]
```

#### `[top]`
Jumps to the top of the current file and re-parses labels. Useful for creating loops.

**Example:**
```
[/loop]Content here
[top]Jumps back to top
```

#### `[quit]`
Exits the current file and stops processing. If an `[on exit]` file is set, it will be executed.

**Example:**
```
Some content
[quit]Stops here
More content (never reached)
```

#### `[exit]`
Exits all files (including any linked files) and stops processing entirely.

**Example:**
```
[link menu.mec]
[exit]Exits everything immediately
```

### Conditionals

#### `[choice <character>]`
Displays the rest of the line only if the menu response or first character of `[readln]` response matches the specified character (case-insensitive).

**Example:**
```
[readln]Enter y or n:
[choice y]You entered yes
[choice n]You entered no
```

#### `[ifentered <string>]`
Displays the rest of the line only if the `[readln]` response matches the specified string exactly (case-insensitive).

**Example:**
```
[readln]Enter your choice:
[ifentered yes]You entered 'yes'
[ifentered no]You entered 'no'
```

---

## Advanced Interactive

### More Prompts

#### `[more]`
Displays the "More [Y,n,=]?" prompt and waits for user input:
- **Y**: Clears screen, resets line counter, continues
- **n**: Quits/stops processing
- **=**: Shows one more line

**Example:**
```
Long content here
[more]
More content after more prompt
```

#### `[moreon]`
Enables automatic more prompts. When enabled, the system will automatically prompt when the output nears the bottom of the terminal.

**Example:**
```
[moreon]
Long content that will trigger automatic prompts
```

#### `[moreoff]`
Disables automatic more prompts.

**Example:**
```
[moreon]Automatic prompts enabled
[moreoff]Automatic prompts disabled
```

---

## Questionnaire System

The questionnaire system collects data in memory. The application layer can retrieve the collected data using `QuestionnaireData()` and handle persistence.

### Data Collection

#### `[readln <description>]`
Collects user input and stores it in questionnaire data with the optional description.

**Example:**
```
[readln Name]What is your name?
[readln Email]What is your email?
```

#### `[store]` / `[store <description>]`
Stores the current menu response in questionnaire data.

**Example:**
```
[menu]
[option a]Option A[reset]
[menuwait]
[store Menu Choice]Stores the selected option
```

#### `[write <text>]`
Writes a line of text directly to questionnaire data.

**Example:**
```
[write Header: User Registration]
[write Date: 2024-01-01]
```

### Answer Requirements

#### `[ansopt]`
Makes subsequent answers optional. Empty answers are allowed.

**Example:**
```
[ansopt]
[readln Optional Field]This field is optional
```

#### `[ansreq]`
Makes subsequent answers required. This is the default behavior.

**Example:**
```
[ansopt]
[readln Optional]Optional field
[ansreq]
[readln Required]Required field
```

---

## Text Utilities

### Control Characters

#### `[bell]`
Outputs a bell/beep character (ASCII 07).

**Example:**
```
[bell]You should hear a beep
```

#### `[bs]`
Outputs a backspace character (ASCII 08).

**Example:**
```
Text[bs]Moves cursor back
```

#### `[tab]`
Outputs a tab character (ASCII 09).

**Example:**
```
Column1[tab]Column2[tab]Column3
```

### Text Manipulation

#### `[pause]`
Pauses execution for half a second (500 milliseconds).

**Example:**
```
Processing...[pause]Done!
```

#### `[repeat <character> <count>]`
Repeats the specified character the specified number of times.

**Example:**
```
[repeat = 40]
[repeat - 20]
[repeat * 10]
```

#### `[comment <text>]`
Comment token. Everything after `comment` is ignored and not processed.

**Example:**
```
[comment This is a comment and will be ignored]
[comment This is useful for documentation]
```

---

## Core Features

### Variable Substitution

Variables can be passed to templates via the `vars` map parameter. Variables are referenced using `[variablename]`.

**Example:**
```go
interpreter.ExecString("Hello, [name]! You have [count] messages.", map[string]any{
    "name":  "Alice",
    "count": 42,
})
```

**Output:**
```
Hello, Alice! You have 42 messages.
```

### Custom Tokens

You can register custom tokens using `RegisterToken()`.

**Example:**
```go
interpreter.RegisterToken("greet", func(args []string) string {
    if len(args) > 0 {
        return "Hello, " + args[0] + "!"
    }
    return "Hello!"
}, 1)

interpreter.ExecString("[greet World]", nil)
```

**Output:**
```
Hello, World!
```

### Character Codes

#### `[<number>]`
Outputs the character with the specified ASCII code.

**Example:**
```
[65]Outputs 'A'
[66]Outputs 'B'
```

#### `[U+<hex>]`
Outputs the character with the specified UTF-8 code point.

**Example:**
```
[U+00A9]Copyright symbol (©)
[U+263A]Smiley face (☺)
```

### Multiple Tokens

Multiple tokens can be combined in a single bracket.

**Example:**
```
[bold red]Bold red text[reset]
[underline blue on white]Underlined blue text on white[reset]
```

### Quoted Arguments

Arguments containing spaces can be quoted.

**Example:**
```
[readln "Full Name"]What is your full name?
[write "Header: User Registration"]
```

### Escaped Brackets

Use `[[` to output a literal `[` character.

**Example:**
```
[[This outputs a literal bracket]
```

---

## Token Usage Notes

### Case Sensitivity
Most tokens are case-insensitive. For example, `[RED]`, `[Red]`, and `[red]` are equivalent.

### Token Processing Order
Tokens are processed left-to-right when multiple tokens appear in the same bracket.

### Error Handling
Invalid tokens or missing files will output error messages inline, such as `[ERROR: file not found]`.

### Reader Configuration
Interactive tokens like `[readln]`, `[menuwait]`, `[enter]`, and `[more]` require a reader to be configured using the `WithReader()` option.

### File Paths
All file operations resolve paths relative to the template root directory, which can be set using `WithTemplateRoot()`.

---

## Quick Reference

### Most Common Tokens

| Token | Purpose |
|-------|---------|
| `[red]`, `[blue]`, etc. | Colors |
| `[bold]`, `[underline]` | Styles |
| `[reset]` | Reset styles |
| `[cls]` | Clear screen |
| `[readln]` | Read input |
| `[menu]` / `[menuwait]` | Menus |
| `[goto <label>]` | Jump to label |
| `[include <file>]` | Include file |

---

*This reference covers all implemented tokens in the MECCA interpreter. For API usage and examples, see the main README.md file.*

