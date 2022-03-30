// Here is the core api for the hilbi shell itself
// Basically, stuff about the shell itself and other functions
// go here.
package main

import (
	"errors"
	"fmt"
	"os"
//	"os/exec"
	"runtime"
	"strings"
//	"syscall"
	"time"

	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
	"github.com/maxlandon/readline"
//	"github.com/blackfireio/osinfo"
//	"mvdan.cc/sh/v3/interp"
)

var exports = map[string]util.LuaExport{
	"alias": util.LuaExport{hlalias, 2, false},
/*
	"appendPath": hlappendPath,
*/
	"complete": util.LuaExport{hlcomplete, 2, false},
	"cwd": util.LuaExport{hlcwd, 0, false},
/*
	"exec": hlexec,
*/
	"runnerMode": util.LuaExport{hlrunnerMode, 1, false},
/*
	"goro": hlgoro,
*/
	"highlighter": util.LuaExport{hlhighlighter, 1, false},
	"hinter": util.LuaExport{hlhinter, 1, false},
/*
	"multiprompt": hlmlprompt,
	"prependPath": hlprependPath,
*/
	"prompt": util.LuaExport{hlprompt, 1, false},
	"inputMode": util.LuaExport{hlinputMode, 1, false},
	"interval": util.LuaExport{hlinterval, 2, false},
	"read": util.LuaExport{hlread, 1, false},
/*
	"run": hlrun,
	"timeout": hltimeout,
	"which": hlwhich,
*/
}

var greeting string
var hshMod *rt.Table
var hilbishLoader = packagelib.Loader{
	Load: hilbishLoad,
	Name: "hilbish",
}

func hilbishLoad(rtm *rt.Runtime) (rt.Value, func()) {
	mod := rt.NewTable()
	util.SetExports(rtm, mod, exports)
	hshMod = mod

//	host, _ := os.Hostname()
	username := curuser.Username

	if runtime.GOOS == "windows" {
		username = strings.Split(username, "\\")[1] // for some reason Username includes the hostname on windows
	}

	greeting = `Welcome to {magenta}Hilbish{reset}, {cyan}` + username + `{reset}.
The nice lil shell for {blue}Lua{reset} fanatics!
Check out the {blue}{bold}guide{reset} command to get started.
`
/*
	util.SetField(L, mod, "ver", lua.LString(version), "Hilbish version")
	util.SetField(L, mod, "user", lua.LString(username), "Username of user")
	util.SetField(L, mod, "host", lua.LString(host), "Host name of the machine")
*/
	util.SetField(rtm, mod, "home", rt.StringValue(curuser.HomeDir), "Home directory of the user")
	util.SetField(rtm, mod, "dataDir", rt.StringValue(dataDir), "Directory for Hilbish's data files")
/*
	util.SetField(L, mod, "interactive", lua.LBool(interactive), "If this is an interactive shell")
	util.SetField(L, mod, "login", lua.LBool(interactive), "Whether this is a login shell")
*/
	util.SetField(rtm, mod, "greeting", rt.StringValue(greeting), "Hilbish's welcome message for interactive shells. It has Lunacolors formatting.")
	/*util.SetField(l, mod, "vimMode", lua.LNil, "Current Vim mode of Hilbish (nil if not in Vim mode)")
	util.SetField(l, hshMod, "exitCode", lua.LNumber(0), "Exit code of last exected command")
	util.Document(L, mod, "Hilbish's core API, containing submodules and functions which relate to the shell itself.")
*/

	// hilbish.userDir table
	hshuser := rt.NewTable()

	util.SetField(rtm, hshuser, "config", rt.StringValue(confDir), "User's config directory")
	util.SetField(rtm, hshuser, "data", rt.StringValue(userDataDir), "XDG data directory")
	//util.Document(rtm, hshuser, "User directories to store configs and/or modules.")
	mod.Set(rt.StringValue("userDir"), rt.TableValue(hshuser))

/*
	// hilbish.os table
	hshos := L.NewTable()
	info, _ := osinfo.GetOSInfo()

	util.SetField(L, hshos, "family", lua.LString(info.Family), "Family name of the current OS")
	util.SetField(L, hshos, "name", lua.LString(info.Name), "Pretty name of the current OS")
	util.SetField(L, hshos, "version", lua.LString(info.Version), "Version of the current OS")
	util.Document(L, hshos, "OS info interface")
	L.SetField(mod, "os", hshos)
*/

	// hilbish.aliases table
	aliases = newAliases()
	aliasesModule := aliases.Loader(rtm)
	//util.Document(L, aliasesModule, "Alias inferface for Hilbish.")
	mod.Set(rt.StringValue("aliases"), rt.TableValue(aliasesModule))

/*
	// hilbish.history table
	historyModule := lr.Loader(L)
	util.Document(L, historyModule, "History interface for Hilbish.")
	L.SetField(mod, "history", historyModule)

	// hilbish.completion table
	hshcomp := L.NewTable()

	util.SetField(L, hshcomp, "files", L.NewFunction(luaFileComplete), "Completer for files")
	util.SetField(L, hshcomp, "bins", L.NewFunction(luaBinaryComplete), "Completer for executables/binaries")
	util.Document(L, hshcomp, "Completions interface for Hilbish.")
	L.SetField(mod, "completion", hshcomp)

*/
	// hilbish.runner table
	runnerModule := runnerModeLoader(rtm)
	//util.Document(L, runnerModule, "Runner/exec interface for Hilbish.")
	mod.Set(rt.StringValue("runner"), rt.TableValue(runnerModule))

	// hilbish.jobs table
	jobs = newJobHandler()
/*
	jobModule := jobs.loader(L)
	util.Document(L, jobModule, "(Background) job interface.")
	L.SetField(mod, "jobs", jobModule)
*/

	return rt.TableValue(mod), nil
}

func getenv(key, fallback string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return fallback
    }
    return value
}

/*
func luaFileComplete(L *lua.LState) int {
	query := L.CheckString(1)
	ctx := L.CheckString(2)
	fields := L.CheckTable(3)

	var fds []string
	fields.ForEach(func(k lua.LValue, v lua.LValue) {
		fds = append(fds, v.String())
	})

	completions := fileComplete(query, ctx, fds)
	luaComps := L.NewTable()

	for _, comp := range completions {
		luaComps.Append(lua.LString(comp))
	}

	L.Push(luaComps)

	return 1
}

func luaBinaryComplete(L *lua.LState) int {
	query := L.CheckString(1)
	ctx := L.CheckString(2)
	fields := L.CheckTable(3)

	var fds []string
	fields.ForEach(func(k lua.LValue, v lua.LValue) {
		fds = append(fds, v.String())
	})

	completions, _ := binaryComplete(query, ctx, fds)
	luaComps := L.NewTable()

	for _, comp := range completions {
		luaComps.Append(lua.LString(comp))
	}

	L.Push(luaComps)

	return 1
}
*/

func setVimMode(mode string) {
	util.SetField(l, hshMod, "vimMode", rt.StringValue(mode), "Current Vim mode of Hilbish (nil if not in Vim mode)")
	hooks.Em.Emit("hilbish.vimMode", mode)
}

func unsetVimMode() {
	util.SetField(l, hshMod, "vimMode", rt.NilValue, "Current Vim mode of Hilbish (nil if not in Vim mode)")
}

/*
// run(cmd)
// Runs `cmd` in Hilbish's sh interpreter.
// --- @param cmd string
func hlrun(L *lua.LState) int {
	var exitcode uint8
	cmd := L.CheckString(1)
	err := execCommand(cmd)

	if code, ok := interp.IsExitStatus(err); ok {
		exitcode = code
	} else if err != nil {
		exitcode = 1
	}

	L.Push(lua.LNumber(exitcode))
	return 1
}
*/

// cwd()
// Returns the current directory of the shell
func hlcwd(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	cwd, _ := os.Getwd()

	return c.PushingNext1(t.Runtime, rt.StringValue(cwd)), nil
}


// read(prompt) -> input?
// Read input from the user, using Hilbish's line editor/input reader.
// This is a separate instance from the one Hilbish actually uses.
// Returns `input`, will be nil if ctrl + d is pressed, or an error occurs (which shouldn't happen)
// --- @param prompt string
func hlread(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	luaprompt, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	lualr := newLineReader("", true)
	lualr.SetPrompt(luaprompt)

	input, err := lualr.Read()
	if err != nil {
		return c.Next(), nil
	}

	return c.PushingNext1(t.Runtime, rt.StringValue(input)), nil
}

/*
prompt(str)
Changes the shell prompt to `str`
There are a few verbs that can be used in the prompt text.
These will be formatted and replaced with the appropriate values.
`%d` - Current working directory
`%u` - Name of current user
`%h` - Hostname of device
--- @param str string
*/
func hlprompt(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	var prompt string
	err := c.Check1Arg()
	if err == nil {
		prompt, err = c.StringArg(0)
	}
	if err != nil {
		return nil, err
	}
	lr.SetPrompt(fmtPrompt(prompt))

	return c.Next(), nil
}

// multiprompt(str)
// Changes the continued line prompt to `str`
// --- @param str string
/*
func hlmlprompt(L *lua.LState) int {
	multilinePrompt = L.CheckString(1)

	return 0
}
*/

// alias(cmd, orig)
// Sets an alias of `cmd` to `orig`
// --- @param cmd string
// --- @param orig string
func hlalias(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	cmd, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	orig, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	aliases.Add(cmd, orig)

	return c.Next(), nil
}

/*
// appendPath(dir)
// Appends `dir` to $PATH
// --- @param dir string|table
func hlappendPath(L *lua.LState) int {
	// check if dir is a table or a string
	arg := L.Get(1)
	if arg.Type() == lua.LTTable {
		arg.(*lua.LTable).ForEach(func(k lua.LValue, v lua.LValue) {
			appendPath(v.String())
		})
	} else if arg.Type() == lua.LTString {
		appendPath(arg.String())
	} else {
		L.RaiseError("bad argument to appendPath (expected string or table, got %v)", L.Get(1).Type().String())
	}

	return 0
}

func appendPath(dir string) {
	dir = strings.Replace(dir, "~", curuser.HomeDir, 1)
	pathenv := os.Getenv("PATH")

	// if dir isnt already in $PATH, add it
	if !strings.Contains(pathenv, dir) {
		os.Setenv("PATH", pathenv + string(os.PathListSeparator) + dir)
	}
}

// exec(cmd)
// Replaces running hilbish with `cmd`
// --- @param cmd string
func hlexec(L *lua.LState) int {
	cmd := L.CheckString(1)
	cmdArgs, _ := splitInput(cmd)
	if runtime.GOOS != "windows" {
		cmdPath, err := exec.LookPath(cmdArgs[0])
		if err != nil {
			fmt.Println(err)
			// if we get here, cmdPath will be nothing
			// therefore nothing will run
		}

		// syscall.Exec requires an absolute path to a binary
		// path, args, string slice of environments
		syscall.Exec(cmdPath, cmdArgs, os.Environ())
	} else {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
		os.Exit(0)
	}

	return 0
}

// goro(fn)
// Puts `fn` in a goroutine
// --- @param fn function
func hlgoro(L *lua.LState) int {
	fn := L.CheckFunction(1)
	argnum := L.GetTop()
	args := make([]lua.LValue, argnum)
	for i := 1; i <= argnum; i++ {
		args[i - 1] = L.Get(i)
	}

	// call fn
	go func() {
		if err := L.CallByParam(lua.P{
			Fn: fn,
			NRet: 0,
			Protect: true,
		}, args...); err != nil {
			fmt.Fprintln(os.Stderr, "Error in goro function:\n\n", err)
		}
	}()

	return 0
}

// timeout(cb, time)
// Runs the `cb` function after `time` in milliseconds
// --- @param cb function
// --- @param time number
func hltimeout(L *lua.LState) int {
	cb := L.CheckFunction(1)
	ms := L.CheckInt(2)

	timeout := time.Duration(ms) * time.Millisecond
	time.Sleep(timeout)

	if err := L.CallByParam(lua.P{
		Fn: cb,
		NRet: 0,
		Protect: true,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "Error in goro function:\n\n", err)
	}
	return 0
}
*/

// interval(cb, time)
// Runs the `cb` function every `time` milliseconds
// --- @param cb function
// --- @param time number
func hlinterval(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	cb, err := c.ClosureArg(0)
	if err != nil {
		return nil, err
	}
	ms, err := c.IntArg(1)
	if err != nil {
		return nil, err
	}
	interval := time.Duration(ms) * time.Millisecond

	ticker := time.NewTicker(interval)
	stop := make(chan rt.Value)

	go func() {
		for {
			select {
			case <-ticker.C:
				_, err := rt.Call1(l.MainThread(), rt.FunctionValue(cb)) 
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error in interval function:\n\n", err)
					stop <- rt.BoolValue(true) // stop the interval
				}
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()

	// TODO: return channel
	return c.Next(), nil
}

// complete(scope, cb)
// Registers a completion handler for `scope`.
// A `scope` is currently only expected to be `command.<cmd>`,
// replacing <cmd> with the name of the command (for example `command.git`).
// `cb` must be a function that returns a table of "completion groups."
// A completion group is a table with the keys `items` and `type`.
// `items` being a table of items and `type` being the display type of
// `grid` (the normal file completion display) or `list` (with a description)
// --- @param scope string
// --- @param cb function
func hlcomplete(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	scope, cb, err := util.HandleStrCallback(t, c)
	if err != nil {
		return nil, err
	}
	luaCompletions[scope] = cb

	return c.Next(), nil
}

/*
// prependPath(dir)
// Prepends `dir` to $PATH
// --- @param dir string
func hlprependPath(L *lua.LState) int {
	dir := L.CheckString(1)
	dir = strings.Replace(dir, "~", curuser.HomeDir, 1)
	pathenv := os.Getenv("PATH")

	// if dir isnt already in $PATH, add in
	if !strings.Contains(pathenv, dir) {
		os.Setenv("PATH", dir + string(os.PathListSeparator) + pathenv)
	}

	return 0
}

// which(binName)
// Searches for an executable called `binName` in the directories of $PATH
// --- @param binName string
func hlwhich(L *lua.LState) int {
	binName := L.CheckString(1)
	path, err := exec.LookPath(binName)
	if err != nil {
		l.Push(lua.LNil)
		return 1
	}

	l.Push(lua.LString(path))
	return 1
}
*/

// inputMode(mode)
// Sets the input mode for Hilbish's line reader. Accepts either emacs for vim
// --- @param mode string
func hlinputMode(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	mode, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	switch mode {
		case "emacs":
			unsetVimMode()
			lr.rl.InputMode = readline.Emacs
		case "vim":
			setVimMode("insert")
			lr.rl.InputMode = readline.Vim
		default:
			return nil, errors.New("inputMode: expected vim or emacs, received " + mode)
	}

	return c.Next(), nil
}

// runnerMode(mode)
// Sets the execution/runner mode for interactive Hilbish. This determines whether
// Hilbish wll try to run input as Lua and/or sh or only do one of either.
// Accepted values for mode are hybrid (the default), hybridRev (sh first then Lua),
// sh, and lua. It also accepts a function, to which if it is passed one
// will call it to execute user input instead.
// --- @param mode string|function
func hlrunnerMode(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	mode := c.Arg(0)

	switch mode.Type() {
		case rt.StringType:
			switch mode.AsString() {
				// no fallthrough doesnt work so eh
				case "hybrid", "hybridRev", "lua", "sh": runnerMode = mode
				default: return nil, errors.New("execMode: expected either a function or hybrid, hybridRev, lua, sh. Received " + mode.AsString())
			}
		case rt.FunctionType: runnerMode = mode
		default: return nil, errors.New("execMode: expected either a function or hybrid, hybridRev, lua, sh. Received " + mode.TypeName())
	}

	return c.Next(), nil
}

// hinter(cb)
// Sets the hinter function. This will be called on every key insert to determine
// what text to use as an inline hint. The callback is passed 2 arguments:
// the current line and the position. It is expected to return a string
// which will be used for the hint.
// --- @param cb function
func hlhinter(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	hinterCb, err := c.ClosureArg(0)
	if err != nil {
		return nil, err
	}
	hinter = hinterCb
	
	return c.Next(), err
}

// highlighter(cb)
// Sets the highlighter function. This is mainly for syntax hightlighting, but in
// reality could set the input of the prompt to display anything. The callback
// is passed the current line as typed and is expected to return a line that will
// be used to display in the line.
// --- @param cb function
func hlhighlighter(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	highlighterCb, err := c.ClosureArg(0)
	if err != nil {
		return nil, err
	}
	highlighter = highlighterCb

	return c.Next(), err
}
