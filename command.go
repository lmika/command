// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package command allows you to define subcommands
// for your command line interfaces. It extends the flag package
// to provide flag support for subcommands.
package command

import (
	"flag"
	"fmt"
	"os"
	"strings"
    "sort"
)

// A map of all of the registered sub-commands.
var cmds map[string]*cmdCont = make(map[string]*cmdCont)

// Declaration of pre-args
var preargdefs []*preArgDef = make([]*preArgDef, 0)

// Matching subcommand.
var matchingCmd *cmdCont

// Arguments to call subcommand's runnable.
var args []string

// Flag to determine whether help is
// asked for subcommand or not
var flagHelp *bool = nil

// Indicates whether or not the -h flag is used for command usage.
// If the OnHelpShowUsage() is called, this will be set to false.
var reserveHFlag bool = true

var helpPreargOverride bool = false

// Cmd represents a sub command, allowing to define subcommand
// flags and runnable to run once arguments match the subcommand
// requirements.
type Cmd interface {
	Flags(*flag.FlagSet) *flag.FlagSet
	Run(args []string)
}



type cmdCont struct {
	name          string
	desc          string
	command       Cmd
	requiredFlags []string
    args          cmdArgs
}

type preArgDef struct {
    name    string
    desc    string
    val     string
}

// A parse error.  This is returned from `TryParse()`
type TryParseError struct {
    // The reason why the command line parsing failed.
    Reason      TryParseReason

    // The command the error relates to.  If the error does not relate to a command, this will be
    // set to the empty string.
    Command     string

    // The error message string.
    Message     string
}

func (tp TryParseError) Error() string {
    return tp.Message
}

// Displays an appropriate usage string depending on the error raised.  If the error relates to
// a command, this displays the command usage string.  Otherwise, this will display the program
// usage string.
func (tp TryParseError) Usage() {
    fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], tp.Message)
    if tp.Command != "" {
        subcommandUsageByName(tp.Command)
    } else {
        Usage()
    }
}


// The reason why a call to TryParse failed.  Depending on the type of error encountered, the
// global flags and pre-arguments may or may not be available.
type TryParseReason int

const (
    // No pre-argument was encountered.
    // Global flags were parsed successfully.
    TryParseNoPreArg                =   iota

    // No command was encountered.
    // Global flags and pre-arguments were parsed successfully.
    TryParseNoCommand               =   iota

    // An undefined command name was encountered.
    // Global flags and pre-arguments were parsed successfully.
    TryParseInvalidCommand          =   iota

    // Invalid argument usage.
    // Global flags and pre-arguments were parsed successfully.
    TryParseArgError                =   iota
)


// Provides configuration operations for Cmds.
type CmdBuilder struct {
    cmd         *cmdCont
}

// Adds a set of arguments that is expected of this command.  Each element of the passed in
// slice is an individual argument name, with the format of the name determining whether
// or not the argument is mandatory or not.  If too many or too few arguments are present,
// the command line parsing will fail with a `TryParseArgError` reason.
//
// The valid argument name formats are:
//
//      name    - A mandatory argument
//      [name]  - An optional argument.  This is consumed greedily.
//      '...'   - Indicates that more arguments are possible.
//      
func (cb *CmdBuilder) Arguments(args ...string) *CmdBuilder {
    if (cb.cmd.args == nil) {
        cb.cmd.args = make([]cmdArg, 0, len(args))
    }

    for _, argstr := range args {
        cb.cmd.args = append(cb.cmd.args, cmdArgFromString(argstr))
    }
    return cb
}

// Registers a Cmd for the provided sub-command name. E.g. name is the
// `status` in `git status`.  Returns a CmdBuilder which can be used to further
// configure the specific command.
func On(name, description string, command Cmd) *CmdBuilder {
    var cmd *cmdCont
	cmd = &cmdCont{
		name:          name,
		desc:          description,
		command:       command,
		requiredFlags: nil,
        args:          nil,
	}

    cmds[name] = cmd
    return &CmdBuilder{cmd}
}

// Registers a help command which will display the usage string of other commands.
// When called, this frees up the '-h' flag for commands to use.
func OnHelpShowUsage() {
    reserveHFlag = false
    On("help", "Displays usage string of commands", cmdUsageCmd(subcommandUsageByName))
}

// When called, will ignore all preargs if the first argument is "help".  Useful for avoiding
// the need for a prearg to show the subcommand usage.  When used, all prearguments will be
// set to the empty string.
func OnHelpIgnorePreargs() {
    helpPreargOverride = true
}

// Registers a PreArg.  This is an argument which is read before the command.
// Returns a string pointer which will be set after calling Parse.
func PreArg(name, description string) *string {
    newPreArgDef := &preArgDef{name, description, ""}
    preargdefs = append(preargdefs, newPreArgDef)
    return &(newPreArgDef.val)
}

// Prints the usage.
func Usage() {
	program := os.Args[0]
	if len(cmds) == 0 {
		// no subcommands
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", program)
		flag.PrintDefaults()
		return
	}

    names := make([]string, 0, len(cmds))
    for _, cmd := range cmds {
        names = append(names, cmd.name)
    }
    sort.Strings(names)

	//fmt.Fprintf(os.Stderr, "Usage: %s <command>\n\n", program)
	fmt.Fprintf(os.Stderr, "Usage: %s", program)
    for _, preargdef := range preargdefs {
        fmt.Fprintf(os.Stderr, " <%s>", preargdef.name)
    }
	fmt.Fprintf(os.Stderr, " <command>\n\n")

	fmt.Fprintf(os.Stderr, "where <command> is one of:\n")
	for _, name := range names {
        cont := cmds[name]
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", name, cont.desc)
	}

	if numOfGlobalFlags() > 0 {
		fmt.Fprintf(os.Stderr, "\navailable flags:\n")
		flag.PrintDefaults()
	}
    if (reserveHFlag) {
        fmt.Fprintf(os.Stderr, "\n%s <command> -h for subcommand help\n", program)
    }
}

func subcommandUsageByName(cmdName string) {
    cont, hasCont := cmds[cmdName]
    if hasCont {
        subcommandUsage(cont)
    } else {
        fmt.Fprintf(os.Stderr, "unreognised command: %s\n", cmdName)
        Usage()
        os.Exit(1)
    }
}

func subcommandUsage(cont *cmdCont) {
	fmt.Fprintf(os.Stderr, "%s\n\n", cont.desc)

	fs := cont.command.Flags(flag.NewFlagSet(cont.name, flag.ContinueOnError))

	fmt.Fprintf(os.Stderr, "Usage: %s %s", os.Args[0], cont.name)
    if (cont.args != nil) {
        for _, arg := range cont.args {
            fmt.Fprintf(os.Stderr, " %s", arg.name)
        }
    }
	fmt.Fprintf(os.Stderr, "\n\n")

    flagCount := 0
    fs.VisitAll(func(_ *flag.Flag) { flagCount++ })

    if (flagCount > 0) {
        fmt.Fprintf(os.Stderr, "Available flags:\n")
        fs.PrintDefaults()
	    if len(cont.requiredFlags) > 0 {
		    fmt.Fprintf(os.Stderr, "\nRequired flags:\n")
            fmt.Fprintf(os.Stderr, "  %s\n\n", strings.Join(cont.requiredFlags, ", "))
	    }
    }
}

// Clear pre-args
func clearPreArgs() {
    preargdefs = make([]*preArgDef, 0)
}

// Parses the flags and leftover arguments to match them with a
// sub-command. Evaluate all of the global flags and register
// sub-command handlers before calling it. Sub-command handler's
// `Run` will be called if there is a match.
// A usage with flag defaults will be printed if provided arguments
// don't match the configuration.
// Global flags are accessible once Parse executes.
func Parse() {
    res := TryParse()

    if (res != nil) {
        res.(TryParseError).Usage()
        os.Exit(1)
    }
}

// Like Parse() but will return an error if there was a problem parsing the flag without
// displaying the usage and exiting.
func TryParse() error {
    var expectedArgCount int = 1
    var commandNameArgN int = 0

	flag.Parse()
	// if there are no subcommands registered,
	// return immediately
	if len(cmds) < 1 {
		return nil
	}


    // Read and set the preargs
    consumePreargs := (helpPreargOverride && !((flag.NArg() > 0) && (flag.Arg(0) == "help"))) || !helpPreargOverride

    if consumePreargs {
        commandNameArgN = len(preargdefs)
        expectedArgCount = commandNameArgN + 1
        if flag.NArg() < expectedArgCount - 1 {
            return TryParseError{TryParseNoPreArg, "", fmt.Sprintf("expected %d argument(s) before command", expectedArgCount - 1)}
        }

        for i, preargdef := range preargdefs {
            preargdef.val = flag.Arg(i)
        }
    }

    // Read and set the commands
	if flag.NArg() < expectedArgCount {
        return TryParseError{TryParseNoCommand, "", "missing command"}
    }

	name := flag.Arg(commandNameArgN)
	if cont, ok := cmds[name]; ok {
		fs := cont.command.Flags(flag.NewFlagSet(name, flag.ExitOnError))
        if (reserveHFlag) {
            flagHelp = fs.Bool("h", false, "")
        }
		fs.Parse(flag.Args()[commandNameArgN + 1:])
		args = fs.Args()
		matchingCmd = cont

		// Check for required flags.
		flagMap := make(map[string]bool)
		for _, flagName := range cont.requiredFlags {
			flagMap[flagName] = true
		}
		fs.Visit(func(f *flag.Flag) {
			delete(flagMap, f.Name)
		})
		if len(flagMap) > 0 {
			return TryParseError{TryParseInvalidCommand, name, name + ": missing required flags"}
		}

        // Validate the arguments
        if (cont.args != nil) {
            err := cont.args.Validate(args)
            if err != nil {
                return TryParseError{TryParseArgError, name, name + ": " + err.Error()}
            }
        }

		return nil
	} else {
        return TryParseError{TryParseInvalidCommand, "", "invalid command: " + name}
	}
}

// Runs the subcommand's runnable. If there is no subcommand
// registered, it silently returns.
func Run() {
	if matchingCmd != nil {
		if (flagHelp != nil) && (*flagHelp) {
			subcommandUsage(matchingCmd)
			return
		}
		matchingCmd.command.Run(args)
	}
}

// Parses flags and run's matching subcommand's runnable.
func ParseAndRun() {
	Parse()
	Run()
}

// Returns the total number of globally registered flags.
func numOfGlobalFlags() (count int) {
	flag.VisitAll(func(flag *flag.Flag) {
		count++
	})
	return
}

// Builtin command for displaying the usage of other commands.
//

type cmdUsageCmd    func(cmd string)

func (cmd cmdUsageCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
    return fs
}

func (cmd cmdUsageCmd) Run(args []string) {
    if (len(args) == 0) {
        Usage()
    } else if (len(args) == 1) {
        cmd(args[0])
    } else {
        cmd("help")
    }
}

// -----------------------------------------------------------------
// Command arguments

type cmdArgType     int
const (
    atMandatory     cmdArgType  =   iota
    atOptional      cmdArgType  =   iota
    atEllipse       cmdArgType  =   iota
)

type cmdArg struct {
    name          string
    argType       cmdArgType
}

// Returns a cmdArgs structure from a string
func cmdArgFromString(argPattern string) cmdArg {
    if (argPattern == "...") {
        return cmdArg{argPattern, atEllipse}
    } else if (argPattern[0] == '[') && (argPattern[len(argPattern)-1] == ']') {
        return cmdArg{argPattern, atOptional}
    } else {
        return cmdArg{"<" + argPattern + ">", atMandatory}
    }
}

// A collection of cmd arguments
type cmdArgs    []cmdArg

// Validates the parsed command line arguments.
func (ca cmdArgs) Validate(args []string) error {
    for _, a := range ca {
        switch a.argType {
        case atMandatory:
            if len(args) == 0 {
                return fmt.Errorf("too few arguments")
            }
            // 'consume' the argument
            args = args[1:]
        case atOptional:
            // Only 'consume' the argument if there are some arguments remaining
            if len(args) > 0 {
                args = args[1:]
            }
        case atEllipse:
            // Consume the remaining arguments
            args = args[0:0]
        }
    }

    if (len(args) != 0) {
        return fmt.Errorf("too many arguments")
    } else {
        return nil
    }
}
