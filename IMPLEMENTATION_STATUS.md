# MECCA Implementation Status

This document summarizes the current implementation status compared to the original MECCA language reference.

## ✅ Fully Implemented

### Color & Style Tokens
- ✅ All basic colors (black, red, green, yellow, blue, magenta, cyan, white)
- ✅ Light variants (lightblue, lightgreen, etc.)
- ✅ Hex colors ([#FF0000], [#202])
- ✅ Background colors ([red on white], [on blue])
- ✅ [bright] / [bold]
- ✅ [dim]
- ✅ [blink] / [steady]
- ✅ [underline]
- ✅ [italic]
- ✅ [reverse]
- ✅ [strike]
- ✅ [reset]
- ✅ [save] / [load]
- ✅ [on] directive

### Cursor Control Tokens
- ✅ [cls] - Clear screen
- ✅ [cleos] - Clear to end of screen
- ✅ [cleol] - Clear to end of line
- ✅ [up] / [down] / [left] / [right]
- ✅ [cr] - Carriage return
- ✅ [lf] - Line feed
- ✅ [locate <r> <c>] - Position cursor
- ✅ [savecursor] / [restorecursor]
- ✅ [line <len> <char>] - Draw line

### File Inclusion
- ✅ [include <file>]
- ✅ [ansi <file>]
- ✅ [ansiconvert <file> <charset>] - With CP437 support

### Menu System (Partial)
- ✅ [menu] - Start menu
- ✅ [option <id>] - Mark option text (until [reset])
- ✅ [menuwait] - Wait for input

### Core Features
- ✅ Custom token registration
- ✅ Variable substitution via vars map
- ✅ ASCII codes ([65])
- ✅ UTF-8 codes ([U+00A9])
- ✅ Multiple tokens per bracket
- ✅ Quoted arguments
- ✅ Escaped brackets ([[)

## ❌ Not Implemented (High Priority for BBS Functionality)

### Questionnaire Tokens (Critical for Interactive Menus)
- ❌ [readln] - Read line input from user
- ❌ [choice <c>] - Conditional based on menu choice
- ❌ [store] - Store menu response to file
- ❌ [open <f>] / [sopen <f>] - Open questionnaire file
- ❌ [post] - Write user info to questionnaire
- ❌ [write <l>] - Write line to questionnaire
- ❌ [ansopt] / [ansreq] - Make answers optional/required

### Flow Control & Conditional Logic (Critical)
- ❌ [goto <label>] / [jump <label>] - Jump to label
- ❌ Labels - [/<label>] or [label <label>]
- ❌ [ifentered <s>] - Conditional based on [readln] response
- ❌ [enter] - Wait for Enter key press
- ❌ [quit] - Exit current file
- ❌ [exit] - Exit all files
- ❌ [top] - Jump to top of file

### Additional Cursor/Video Tokens
- ❌ [bell] - Beep (ASCII 07)
- ❌ [bs] - Backspace (ASCII 08)
- ❌ [tab] - Tab character
- ❌ [sysopbell] - Beep on console only

### Color Tokens (Missing)
- ❌ [BG <c>] - Set background only
- ❌ [FG <c>] - Set foreground only

### File Operations
- ❌ [display <f>] - Display file (no return)
- ❌ [link <f>] - Display file (return to caller)
- ❌ [copy <f>] - Copy file to output
- ❌ [delete <f>] - Delete file
- ❌ [on exit <f>] - Set exit file

### Conditional Tokens (User/System State)
- ❌ [color] / [nocolor] / [endcolor] - ANSI color conditional
- ❌ [rip] / [norip] / [endrip] - RIPscrip conditional
- ❌ [iftime <op> <hh>:<mm>] - Time-based conditional
- ❌ [ifexist <file>] - File existence check
- ❌ [ifkey <keys>] / [notkey <keys>] - Key presence check
- ❌ [keyon <keys>] / [keyoff <keys>] - Key manipulation
- ❌ [access <acs-string>] - Access control check
- ❌ [acsfile <acs-string>] - File-level access control

### Informational Tokens (BBS-Specific)
- ❌ All user info tokens ([user], [fname], [city], [date], [time], etc.)
- ❌ All message area tokens ([msg_carea], [msg_cname], etc.)
- ❌ All file area tokens ([file_carea], [file_cname], etc.)
- ❌ System info tokens ([sys_name], [sysop_name], [node_num], etc.)
- Note: These are BBS-specific and should be implemented via custom tokens or variables

### Misc Tokens
- ❌ [more] - "More [Y,n,=]?" prompt
- ❌ [moreon] / [moreoff] - Enable/disable more prompts
- ❌ [pause] - Pause half second
- ❌ [hangup] - Disconnect user
- ❌ [repeat <c>[<n>]] - Repeat character
- ❌ [repeatseq <len>]<s>[<n>] - Repeat string
- ❌ [comment <c>] - Comments
- ❌ [clear_stacked] - Clear input buffer
- ❌ [ckon] / [ckoff] - Enable/disable Ctrl-C checking

### External Program Execution
- ❌ [dos <c>] - Run OS command
- ❌ [xtern_*] - External program tokens
- ❌ [mex <file>] - Run MEX program

## 🔄 Partially Implemented

### Menu System
- ✅ Basic menu functionality exists
- ⚠️ Missing [choice] token for conditional display based on menu selection
- ⚠️ Missing integration with [readln] and questionnaire system
- ⚠️ Missing [store] to save menu responses

## 📋 Implementation Priority

### Priority 1: Core Interactive Features (Essential for Menu/Questionnaire Workflow)
1. **[readln]** - Read user input (single line)
2. **[choice <c>]** - Conditional based on menu/[readln] response
3. **[goto <label>]** + **Labels** - Flow control for menus and questionnaires
4. **[enter]** - Wait for Enter key press
5. **[quit]** / **[exit]** - Exit control

### Priority 2: File Operations
1. **[display <f>]** - Display file without return
2. **[link <f>]** - Display file with return
3. **[on exit <f>]** - Exit file handler

### Priority 3: Additional Interactive Features
1. **[open <f>]** / **[post]** / **[write]** / **[store]** - Questionnaire file system
2. **[ansopt]** / **[ansreq]** - Optional/required answers
3. **[ifentered <s>]** - Conditional on input

### Priority 4: Missing Color/Video Tokens
1. **[BG <c>]** / **[FG <c>]** - Individual color setting
2. **[bell]** / **[bs]** / **[tab]** - Additional control characters
3. **[color]** / **[nocolor]** - Conditional color display

### Priority 5: Conditional Logic (BBS-Specific)
- These depend on having a BBS system context (user state, file system, etc.)
- Should be implemented via custom tokens or integration with BBS backend

### Priority 6: External Program Execution
- Lower priority for standalone library
- May not be applicable depending on use case

## 📝 Notes

### Design Decisions Made
1. **Informational tokens** ([user], [date], etc.) are intentionally not built-in, as they require BBS-specific context. These should be provided via:
   - Custom token registration
   - Variable substitution
   - Integration with application state

2. **Access control tokens** ([access], [acsfile]) require an ACS system which is BBS-specific and should be handled externally.

3. **External program execution** ([dos], [xtern_*]) may not be appropriate for a library and should be handled by the application layer.

### What Makes This Library "BBS-Ready"
To be considered feature-complete for BBS usage, the library should have:
- ✅ Basic color/styling (DONE)
- ✅ Cursor control (DONE)
- ⚠️ Interactive input ([readln]) (MISSING - HIGH PRIORITY)
- ⚠️ Flow control ([goto], labels) (MISSING - HIGH PRIORITY)
- ⚠️ Menu system with [choice] (PARTIALLY DONE)
- ⚠️ File operations ([display], [link]) (MISSING)
- ✅ File inclusion (DONE)

## Summary

**Current Status:** ~40% feature parity
- **Excellent:** Color, styling, cursor control, basic file inclusion
- **Good:** Menu system (basic functionality)
- **Missing:** Interactive input, flow control, conditional logic, advanced file operations

**Next Steps for BBS Parity:**
1. Implement [readln] with WithReader() support
2. Implement [goto] and label system
3. Implement [choice] for menu conditionals
4. Implement [enter] for pauses
5. Implement [display] and [link] for file chaining

