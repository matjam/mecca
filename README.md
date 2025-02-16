# MECCA Language Reference

Note: Parts of this document is based on the original MECCA documentation by Scott J. Dudley, and has been adapted for the Go interpretation of the MECCA language. You can find the original documentation at http://software.bbsdocumentary.com/IBM/DOS/MAXIMUS/max300.txt

MECCA is based on the MECCA language originally designed for the Maximus BBS software by Scott J. Dudley. MECCA is a simple, easy to learn language that is designed to be used by non-programmers to create interactive content for BBS systems.

This implementation provides a subset of the MECCA language primarily for creating terminal based user interfaces that would be driven via the https://github.com/charmbracelet/bubbletea framework in modern Go BBS software. As we use the lipgloss library for rendering, the MECCA language has been adapted to work with lipgloss and has been extended to support some additional features unique to lipgloss.

## Syntax

The input file for the MECCA interpreter normally has a .mec extension and consists of plain UTF-8 text. The file can contain specia MECCA tokens which are parsed and rendered by the MECCA interpreter.

In MECCA, a token is delimited by a set of square brackets (e.g. `[token]`). The token can contain a command and arguments. The command is separated from the arguments by a space. Additional arguments are also separated by spaces, and can be quoted if they contain spaces. Anything outside of a token is considered to be text that is rendered as-is. Different types of tokens can be inserted in a MECCA file to achieve different effects. For example, the following line of text:

    This is your [usercall] call to the system.

might be displayed as:

    This is your 14th call to the system.

Other tokens can be used to display the user's name, show information about the system, or display a menu of options.

The MECCA interpreter only processes tokens that are contained inside square brackets. To include a left square bracket in the text, use two left square brackets. Only the left square bracket needs to be doubled. For example, use the following

    Want to check for your mail [[Y,n]?

which will render as:

    Want to check for your mail [Y,n]?

When using MECCA tokens, also keep these points in mind:

Tokens are not case-sensitive. That means that the following tokens are equivalent in all respects:

    [user]
    [USER]
    [UsEr]

Spaces are ignored. Inside MECCA tokens, any spaces, tabs or newlines will have no effect. This means that the following tokens are equivalent in all respects:

    [  user  ]
    [    user]
    [user    ]

More than one token can be inserted inside a set of square brackets, as long as tokens are separated from each other using spaces. For example, this line:

    [lightblue][blink][user]

can also be written as follows:

    [lightblue blink user]

MECCA also allows you to use ASCII and UTF-8 codes directly in the .mec file. To insert a specific ASCII code, simply enclose the ASCII code number inside a pair of square brackets. For example, the token [123] will be compiled to ASCII code 123 in the output file. To insert a UTF-8 code, use the format [U+xxxx] where xxxx is the hexadecimal UTF-8 code. For example, the token [U+00A9] will be compiled to the copyright symbol ©.

Note that the [user] token mentioned above is not a token that is recognized by the MECCA interpreter. It is used here as an example of a token that could be used in a MECCA file, once the MECCA interpreter has been extended to support it. It is intended that the application using the MECCA interpreter will define the tokens that are recognized by the interpreter.

### Color Tokens

MECCA supports a number of color tokens that can be used to change the color of text. TTo display text in a specific color, simply enclose the name of the color in square brackets. For example, to display the text "Hello, world!" in red, use the following token:

    [red]Hello, world!

To display text on a colored background, use a token of the form "foreground on background". For example, to display the text "Hello, world!" in red on a white background, use the following token:

    [red on white]Hello, world!

MECCA supports the ANSI 4-bit color palette, which includes the following colors which can be used as foreground and background colors:

    black
    red
    green
    yellow
    blue
    magenta
    cyan
    white

as well as the following additional colors which can only be used as foreground colors:

    lightblack
    lightred
    lightgreen
    lightyellow
    lightblue
    lightmagenta
    lightcyan
    lightwhite

MECCA also supports ANSI 256-color mode, which allows you to specify colors using a 256-color palette. To use a 256-color, prefix the color number with a `#`. For example, to display text in color 202 (orange), use the following token:

    [#202]Hello, world!

Lastly, MECCA supports true color mode, which allows you to specify colors using RGB values. To use true color, prefix the RGB values with a `#`. For example, to display text in a light blue color, use the following token:

    [#ADD8E6]Hello, world!

These color modes can be combined to create more complex color schemes. For example, to display text in light blue on a dark blue background, use the following token:

    [lightblue on #00008B]Hello, world!

### Text Style Tokens

MECCA supports a number of text style tokens that can be used to change the style of text. To display text in a specific style, simply enclose the name of the style in square brackets. For example, to display the text "Hello, world!" in bold, use the following token:

    [bold]Hello, world!

The supported text styles are:

    [blink]                 Any text that follows this token will blink. The blink attribute is reset when the MECCA interpreter encounters a color token.
    [bright] / [bold]       makes the text bright or bold.
    [dim]                   makes the text dim.
    [ON <color>]            tells the interpreter that the next token is a background color. By itself, it will not change the foreground color.
    [load]                  restores the color and style that was in effect before the last [save] token.
    [save]                  saves the current color and style so that it can be restored later using the [load] token.
    [steady]                cancels the previous [blink] token.
    [cleol]                 clears to the end of the line.
    [cleos]                 clears to the end of the screen.
    [cls]                   clears the screen.
    [cr]                    moves the cursor to the beginning of the line.
    [lf]                    moves the cursor down one line.
    [up]                    moves the cursor up one line.
    [down]                  moves the cursor down one line.
    [right]                 moves the cursor right one column.
    [left]                  moves the cursor left one column.
    [locate <row> <column>] moves the cursor to the specified row and column. The first row is 0 and the first column is 0.
    [savecursor]            saves the current cursor position.
    [restorecursor]         restores the cursor position saved by the [savecursor] token.
    [box <width> <height>]  draws a box of the specified width and height from the current cursor position.
    [line <length> <char>]  draws a line of the specified length using the specified character from the current cursor position.
    [underline]             underlines the text.
    [reverse]               reverses the foreground and background colors.
    [italic]                makes the text italic.
    [strike]                strikes through the text.    

### Implementing a new Token

Users of this library can implement their own tokens by implementing a function that takes an array of `string` arguments and returns a `string`. The function should be registered with the `mecca` package using the `RegisterToken` function. For example, to implement a token that displays the user's name, you could write a function like this:

```go
func userToken(args []string) string {
    return "John Doe"
}
```

You would then register this function with the `mecca` package like this:

```go
mecca.RegisterToken("user", userToken, 0)
```

The third argument to `RegisterToken` is the number of arguments that the token takes. In this case, the `user` token takes no arguments, so we pass 0.

Once you have registered the token, you can use it in a MECCA file like this:

```
Hello, [user]!
```

When the MECCA interpreter encounters the `[user]` token, it will call the `userToken` function and replace the token with the return value of the function.

Tokens registered in this way are intended to be purely a simple string as the return value.

## Example

Here is an example of a simple MECCA file that displays a welcome message to the user:

```
[lightblue][bold]Welcome to the MECCA interpreter![load]
```

When this file is passed to the MECCA interpreter, it will display the text "Welcome to the MECCA interpreter!" in light blue and bold text.

## License

This implementation of the MECCA language is released under the MIT license. See the LICENSE file for more information.

# Acknowledgements

The original MECCA language was designed by Scott J. Dudley for the Maximus BBS software. This implementation is based on the original MECCA language and has been adapted for the Go programming language. 

I thank Scott J. Dudley for creating the original MECCA language and for making it available to the public. When I was a teenager, I spent many hours creating interactive content for my BBS using the MECCA language, and I have fond memories of those days. I hope that this implementation will allow others to create similar content for modern BBS systems.
