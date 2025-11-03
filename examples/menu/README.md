# Menu Example

This example demonstrates the interactive menu functionality of the MECCA library.

## Features Demonstrated

- **[menu]** token to start a menu
- **[option]** token to add menu options with styled option IDs
- **[menuwait]** token to wait for user input
- Using `WithReader()` to enable interactive input
- Using `MenuResponse()` to retrieve the selected option
- Color styling and formatting typical of BBS-style menus

## Running the Example

From the project root directory:

```bash
go run examples/menu/main.go
```

The program will:
1. Display a colorful BBS-style menu
2. Wait for you to type a single character (a, b, c, d, e, or q)
3. Display the selected option and exit

## Example Output

```
╔═══════════════════════════════════════════╗
║                                           ║
║    WELCOME TO THE MAIN MENU               ║
║                                           ║
╚═══════════════════════════════════════════╝

Please select an option:

A View Messages
B Read Email
C File Downloads
D User Directory
E Chat Room
Q Quit / Logout

Your choice: 
```

After typing a character (e.g., 'A'), it will display:

```
You selected: A (View Messages)
```

