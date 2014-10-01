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
    minArgs       []string
}

type preArgDef struct {
    name    string
    desc    string
    val     string
}

// TryParse result
type TryParseResult int
const (
    // The command was parsed successfully
    TryParseOK      TryParseResult  =   iota

    // No pre-argument was encountered.
    // Global flags were parsed successfully.
    TryParseNoPreArg                =   iota

    // No command was encountered.
    // Global flags and pre-arguments were parsed successfully.
    TryParseNoCommand               =   iota

    // An undefined command name was encountered.
    // Global flags and pre-arguments were parsed successfully.
    TryParseInvalidCommand          =   iota

    // Not enough required arguments entered.
    // Global flags and pre-arguments were parsed successfully.
    TryParseNotEnoughArgs           =   iota
)


// Provides configuration operations for Cmds.
type CmdBuilder struct {
    cmd         *cmdCont
}

// Adds a set of arguments that the command expects.
func (cb *CmdBuilder) Arguments(args ...string) *CmdBuilder {
    if (cb.cmd.minArgs == nil) {
        cb.cmd.minArgs = make([]string, 0, len(args))
    }
    cb.cmd.minArgs = append(cb.cmd.minArgs, args...)
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
        minArgs:       nil,
	}

    cmds[name] = cmd
    return &CmdBuilder{cmd}
}

// Registers a help command which will display the usage string of other commands.
// When called, this frees up the '-h' flag for commands to use.
func OnHelpShowUsage() {
    reserveHFlag = false
    On("help", "Displays usage string of commands", CmdUsageCmd(subcommandUsageByName))
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

	//fmt.Fprintf(os.Stderr, "Usage: %s <command>\n\n", program)
	fmt.Fprintf(os.Stderr, "Usage: %s", program)
    for _, preargdef := range preargdefs {
        fmt.Fprintf(os.Stderr, " <%s>", preargdef.name)
    }
	fmt.Fprintf(os.Stderr, " <command>\n\n")

	fmt.Fprintf(os.Stderr, "where <command> is one of:\n")
	for name, cont := range cmds {
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
    if (cont.minArgs != nil) {
        for _, arg := range cont.minArgs {
            fmt.Fprintf(os.Stderr, " %s", arg)
        }
    }

	fmt.Fprintf(os.Stderr, "\n\n")
	fmt.Fprintf(os.Stderr, "Available flags:\n")
	// should only output sub command flags, ignore h flag.
	fs.PrintDefaults()
	if len(cont.requiredFlags) > 0 {
		fmt.Fprintf(os.Stderr, "\nRequired flags:\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", strings.Join(cont.requiredFlags, ", "))
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

    if (res != TryParseOK) {
        if (res == TryParseNotEnoughArgs) {
            fmt.Fprintf(os.Stderr, "%s: not enough args to %s\n", os.Args[0], matchingCmd.name)
            subcommandUsage(matchingCmd)
        } else {
            flag.Usage = Usage
            flag.Usage()
        }
        os.Exit(1)
    }
}

// Like Parse() but returns a TryParseResult.
func TryParse() TryParseResult {
    var expectedArgCount int = 1
    var commandNameArgN int = 0

	flag.Parse()
	// if there are no subcommands registered,
	// return immediately
	if len(cmds) < 1 {
		return TryParseOK
	}


    // Read and set the preargs
    consumePreargs := (helpPreargOverride && !((flag.NArg() > 0) && (flag.Arg(0) == "help"))) || !helpPreargOverride

    if consumePreargs {
        commandNameArgN = len(preargdefs)
        expectedArgCount = commandNameArgN + 1
        if flag.NArg() < expectedArgCount - 1 {
            return TryParseNoPreArg
        }

        for i, preargdef := range preargdefs {
            preargdef.val = flag.Arg(i)
        }
    }

    // Read and set the commands
	if flag.NArg() < expectedArgCount {
        return TryParseNoCommand
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
			return TryParseInvalidCommand
		}

        // If the command declares a minimum number of args, check that the argument count
        // matches.
        if (cont.minArgs != nil) && (len(args) < len(cont.minArgs)) {
            return TryParseNotEnoughArgs
        }

		return TryParseOK
	} else {
        return TryParseInvalidCommand
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

type CmdUsageCmd    func(cmd string)

func (cmd CmdUsageCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
    return fs
}

func (cmd CmdUsageCmd) Run(args []string) {
    if (len(args) == 0) {
        Usage()
    } else if (len(args) == 1) {
        cmd(args[0])
    } else {
        cmd("help")
    }
}
