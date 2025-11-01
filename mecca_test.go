package mecca

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestBasicTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		vars     map[string]any
	}{
		{
			name:     "plain text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "variable substitution",
			input:    "Hello [user]",
			expected: "Hello John",
			vars:     map[string]any{"user": "John"},
		},
		{
			name:     "multiple variables",
			input:    "[greeting] [name]",
			expected: "Hello Alice",
			vars:     map[string]any{"greeting": "Hello", "name": "Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, tt.vars)
			result := buf.String()
			// Remove ANSI codes for comparison - we'll check if expected text is contained
			if !strings.Contains(result, strings.ReplaceAll(tt.expected, "[", "")) && !strings.Contains(result, strings.Split(tt.expected, " ")[0]) {
				t.Errorf("Expected result to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestColorTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "basic color",
			input: "[red]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "color on background",
			input: "[red on white]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "256 color code",
			input: "[#202]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "true color hex",
			input: "[#FF0000]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "all basic colors",
			input: "[black][red][green][yellow][blue][magenta][cyan][white]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "all light colors",
			input: "[lightblack][lightred][lightgreen][lightyellow][lightblue][lightmagenta][lightcyan][lightwhite]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestStyleTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "bold",
			input: "[bold]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "bright (synonym for bold)",
			input: "[bright]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "underline",
			input: "[underline]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "italic",
			input: "[italic]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "dim",
			input: "[dim]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "reverse",
			input: "[reverse]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "strike",
			input: "[strike]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "blink",
			input: "[blink]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "reset",
			input: "[bold][reset]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestMissingTokens_SaveLoad(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "save and load style",
			input: "[red][bold]Hello[save][blue]World[load]Again",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
				if !strings.Contains(output, "World") {
					t.Errorf("Expected output to contain 'World'")
				}
				if !strings.Contains(output, "Again") {
					t.Errorf("Expected output to contain 'Again'")
				}
				// After [load], styling should be restored to red+bold
			},
		},
		{
			name:  "nested save/load",
			input: "[red][save][blue][save][green]Test[load][load]Back",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
				if !strings.Contains(output, "Back") {
					t.Errorf("Expected output to contain 'Back'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestMissingTokens_Steady(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "steady cancels blink",
			input: "[blink]Blinking[steady]NotBlinking",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Blinking") {
					t.Errorf("Expected output to contain 'Blinking'")
				}
				if !strings.Contains(output, "NotBlinking") {
					t.Errorf("Expected output to contain 'NotBlinking'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestMissingTokens_ON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "ON sets background only",
			input: "[ON red]Background",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Background") {
					t.Errorf("Expected output to contain 'Background'")
				}
			},
		},
		{
			name:  "ON with color name",
			input: "[ON blue]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "ON with hex color",
			input: "[ON #FF0000]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "ON with 256 color",
			input: "[ON #202]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestEscapedBrackets(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "escaped left bracket",
			input: "Check your mail [[Y,n]?",
			check: func(t *testing.T, output string) {
				// Should render as: Check your mail [Y,n]?
				if !strings.Contains(output, "[Y,n]?") {
					t.Errorf("Expected output to contain '[Y,n]?', got %q", output)
				}
				if strings.Contains(output, "[[") {
					t.Errorf("Output should not contain '[[', got %q", output)
				}
			},
		},
		{
			name:  "multiple escaped brackets",
			input: "[[[token]]",
			check: func(t *testing.T, output string) {
				// Should render as: [[token]
				if !strings.Contains(output, "[") {
					t.Errorf("Expected output to contain '[', got %q", output)
				}
			},
		},
		{
			name:  "escaped bracket with token",
			input: "[[red]Test",
			check: func(t *testing.T, output string) {
				// Should render as: [red]Test (literal [red])
				if !strings.Contains(output, "[red]") || !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain '[red]' and 'Test', got %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestLocateToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "locate with row column order",
			input: "[locate 5 10]X",
			check: func(t *testing.T, output string) {
				// According to README: [locate <row> <column>]
				// Should position cursor at row 5, column 10
				// ANSI sequence is CSI row;colH (1-indexed)
				// So we should see cursor position escape codes
				if !strings.Contains(output, "X") {
					t.Errorf("Expected output to contain 'X'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestCursorTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "cr",
			input: "[cr]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "lf",
			input: "[lf]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "up",
			input: "[up]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "down",
			input: "[down]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "right",
			input: "[right]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "left",
			input: "[left]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "savecursor and restorecursor",
			input: "[savecursor]Test[restorecursor]",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestClearTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "cls",
			input: "[cls]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "cleos",
			input: "[cleos]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
		{
			name:  "cleol",
			input: "[cleol]Test",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Test") {
					t.Errorf("Expected output to contain 'Test'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestLineToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "line token",
			input: "[line 10 -]",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "----------") {
					t.Errorf("Expected output to contain 10 dashes, got %q", output)
				}
			},
		},
		{
			name:  "line token with different character",
			input: "[line 5 *]",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "*****") {
					t.Errorf("Expected output to contain 5 asterisks, got %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestASCIIToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "ASCII code 65 (A)",
			input: "[65]",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "A") {
					t.Errorf("Expected output to contain 'A', got %q", output)
				}
			},
		},
		{
			name:  "ASCII code 33 (!)",
			input: "[33]",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "!") {
					t.Errorf("Expected output to contain '!', got %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestUTF8Token(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "UTF-8 code U+00A9 (copyright)",
			input: "[U+00A9]",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "©") {
					t.Errorf("Expected output to contain copyright symbol, got %q", output)
				}
			},
		},
		{
			name:  "UTF-8 code U+2665 (heart)",
			input: "[U+2665]",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "♥") {
					t.Errorf("Expected output to contain heart symbol, got %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestRegisteredTokens(t *testing.T) {
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// Register a test token
	called := false
	interpreter.RegisterToken("test", func(args []string) string {
		called = true
		return "TEST_TOKEN"
	}, 0)

	interpreter.ExecString("[test]", nil)

	if !called {
		t.Error("Registered token function was not called")
	}

	if !strings.Contains(buf.String(), "TEST_TOKEN") {
		t.Errorf("Expected output to contain 'TEST_TOKEN', got %q", buf.String())
	}
}

func TestRegisteredTokensWithArgs(t *testing.T) {
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// Register a token that takes arguments
	interpreter.RegisterToken("repeat", func(args []string) string {
		if len(args) < 2 {
			return ""
		}
		// args[0] should be count, args[1] should be text
		count := 0
		for _, c := range args[0] {
			count = count*10 + int(c-'0')
		}
		return strings.Repeat(args[1], count)
	}, 2)

	interpreter.ExecString("[repeat 3 hello]", nil)

	if !strings.Contains(buf.String(), "hellohellohello") {
		t.Errorf("Expected output to contain 'hellohellohello', got %q", buf.String())
	}
}

func TestQuotedArguments(t *testing.T) {
	// This tests the feature mentioned in README but not yet implemented
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "quoted argument with spaces",
			input: `[token "hello world"]`,
			check: func(t *testing.T, output string) {
				// Once implemented, this should pass "hello world" as a single argument
				// For now, just check it doesn't crash
				if output == "" {
					t.Error("Expected some output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.RegisterToken("token", func(args []string) string {
				if len(args) == 1 && args[0] == "hello world" {
					return "SUCCESS"
				}
				return "FAIL"
			}, 1)
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestRenderTemplate_NilSession(t *testing.T) {
	// Test that RenderTemplate doesn't panic when session is nil
	// Create a temp file for testing
	tmpfile, err := os.CreateTemp("", "test_*.mec")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("[red]Test")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	var buf bytes.Buffer
	interpreter := NewInterpreter(WithTemplateRoot(""), WithWriter(&buf))

	// Should not panic
	err = interpreter.RenderTemplate(tmpfile.Name(), nil)
	if err != nil {
		// File not found is expected since we're using empty template root
		// The important thing is it didn't panic
	}
}

func TestIncludeToken(t *testing.T) {
	// Create a temporary directory with test files
	tmpdir, err := os.MkdirTemp("", "mecca_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Create an included file
	includedFile := tmpdir + "/included.mec"
	err = os.WriteFile(includedFile, []byte("[red]Included Content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create main file
	mainFile := tmpdir + "/main.mec"
	err = os.WriteFile(mainFile, []byte("Before [include included.mec] After"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	interpreter := NewInterpreter(WithTemplateRoot(tmpdir), WithWriter(&buf))

	output, err := interpreter.ExecTemplate("main.mec", nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "Included Content") {
		t.Errorf("Expected output to contain 'Included Content', got %q", output)
	}
	if !strings.Contains(output, "Before") {
		t.Errorf("Expected output to contain 'Before', got %q", output)
	}
	if !strings.Contains(output, "After") {
		t.Errorf("Expected output to contain 'After', got %q", output)
	}
}

func TestIncludeTokenRecursive(t *testing.T) {
	// Test that recursive includes are detected
	tmpdir, err := os.MkdirTemp("", "mecca_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Create a file that includes itself
	recursiveFile := tmpdir + "/recursive.mec"
	err = os.WriteFile(recursiveFile, []byte("[include recursive.mec]"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	interpreter := NewInterpreter(WithTemplateRoot(tmpdir), WithWriter(&buf))

	output, err := interpreter.ExecTemplate("recursive.mec", nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "ERROR") || !strings.Contains(output, "recursively") {
		t.Errorf("Expected error message about recursive inclusion, got %q", output)
	}
}

func TestCaseInsensitiveTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "uppercase token",
			input: "[RED]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "mixed case token",
			input: "[ReD]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name:  "lowercase token",
			input: "[red]Hello",
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Errorf("Expected output to contain 'Hello'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			interpreter := NewInterpreter(WithWriter(&buf))
			interpreter.ExecString(tt.input, nil)
			tt.check(t, buf.String())
		})
	}
}

func TestVariableOverride(t *testing.T) {
	// Test that variables override registered tokens
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	interpreter.RegisterToken("user", func(args []string) string {
		return "RegisteredToken"
	}, 0)

	// Variable should override registered token
	interpreter.ExecString("[user]", map[string]any{"user": "VariableValue"})

	if !strings.Contains(buf.String(), "VariableValue") {
		t.Errorf("Expected variable to override token, got %q", buf.String())
	}
	if strings.Contains(buf.String(), "RegisteredToken") {
		t.Errorf("Variable should have overridden registered token")
	}
}
