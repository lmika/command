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

package command

import (
	"flag"
	"os"
	"testing"
)

// Tests if global flags default values are set if there are
// no flags provided.
func TestDefaultGlobalFlags(t *testing.T) {
	resetForTesting()

	flagGlobal1 := flag.String("global1", "default-global1", "Description about global1")
	Parse()
	if *flagGlobal1 != "default-global1" {
		t.Error("global flag should be set to default val if the flag is not set")
	}
}

// Tests if global flags are set if they are provided by the user.
func TestGlobalFlags(t *testing.T) {
	resetForTesting("-global1=hello")

	flagGlobal1 := flag.String("global1", "default-global1", "Description about global1")
	Parse()
	if *flagGlobal1 != "hello" {
		t.Errorf("global flag should be set: expected default-global1, found %s", *flagGlobal1)
	}
}

// Tests the total number of globally registered flags.
func TestGlobalFlagsCount(t *testing.T) {
	resetForTesting("-global1=hello", "-global2=hi")

	flag.String("global1", "default-global1", "Description about global1")
	flag.String("global2", "default-global2", "Description about global2")
	Parse()

	total := numOfGlobalFlags()
	if total != 2 {
		t.Error("total number of global flags are expected to be 2, found %v", total)
	}
}

// Tests if subcommand runs if it's provided as a part of arguments.
func TestCommand(t *testing.T) {
	resetForTesting("-global1=hello", "command1")

	flagGlobal1 := flag.String("global1", "default-global1", "Description about global1")
	c1 := &testCmd1{}
	On("command1", "", c1)
	Parse()
	Run()
	if !c1.run {
		t.Error("command 'command1' was expected to run, but it didn't")
	}
	if *c1.flag1 {
		t.Errorf("flag1 should be set to default: expected false, found %v", *c1.flag1)
	}
	if *flagGlobal1 != "hello" {
		t.Errorf("global flag should be set: expected default-global1, found %s", *flagGlobal1)
	}
}

// Tests if subcommand runs and subcommand flags are set.
func TestCommandFlags(t *testing.T) {
	resetForTesting("-global1=hello", "command1", "-flag1=true")

	flag.String("global1", "default-global1", "Description about global1")
	c1 := &testCmd1{}
	On("command1", "", c1)
	Parse()
	Run()
	if !c1.run {
		t.Error("command 'command1' was expected to run, but it didn't")
	}
	if !*c1.flag1 {
		t.Errorf("flag1 should be set: expected true, found %v", *c1.flag1)
	}
}

// Tests if correct subcommand runs if multiple subcommands
// are registered.
func TestMultiCommands(t *testing.T) {
	resetForTesting("command2")

	c1 := &testCmd1{}
	c2 := &testCmd2{}
	On("command1", "", c1)
	On("command2", "", c2)
	Parse()
	Run()
	if c1.run {
		t.Error("command 'command1' was not expected to run, but it did")
	}
	if !c2.run {
		t.Error("command 'command2' was expected to run, but it didn't")
	}
}

// Tests if subcommand runnable has run, if Run is not invoked.
func TestRun(t *testing.T) {
	resetForTesting("command1")

	c1 := &testCmd1{}
	On("command1", "", c1)
	Parse()
	if c1.run {
		t.Error("command 'command1' was not expected to run, but it did")
	}
}

func TestAdditionalCommandArgs(t *testing.T) {
	resetForTesting("command1", "--flag1=true", "somearg")

	c1 := &testCmd1{}
	On("command1", "", c1)
	Parse()
	if len(args) < 1 || args[0] != "somearg" {
		t.Error("additional command 'somearg' is expected, but can't be found")
	}
}

// Test that the presence of a pre-arg will set prearg
func TestPreargs(t *testing.T) {
    resetForTesting("some-prearg", "command1", "--flag1=true", "somearg")

	c1 := &testCmd1{}
	On("command1", "", c1)
    prearg := PreArg("prearg", "Some example prearg")
	Parse()
    if *prearg != "some-prearg" {
        t.Error("prearg expected to be 'some-prearg'")
    }
	if len(args) < 1 || args[0] != "somearg" {
		t.Error("additional command 'somearg' is expected, but can't be found")
	}
}

// Test multiple prearg
func TestMultiplePreargs(t *testing.T) {
    resetForTesting("pa1", "pa2", "pa3", "command1", "--flag1=true", "somearg")

	c1 := &testCmd1{}
	On("command1", "", c1)
    prearg1 := PreArg("prearg1", "Some example prearg")
    prearg2 := PreArg("prearg2", "Some example prearg")
    prearg3 := PreArg("prearg3", "Some example prearg")
	Parse()
    if *prearg1 != "pa1" {
        t.Error("prearg1 expected to be 'pa1'")
    }
    if *prearg2 != "pa2" {
        t.Error("prearg2 expected to be 'pa2'")
    }
    if *prearg3 != "pa3" {
        t.Error("prearg3 expected to be 'pa3'")
    }
	if len(args) < 1 || args[0] != "somearg" {
		t.Error("additional command 'somearg' is expected, but can't be found")
	}
}

// Tests if subcommand runs and subcommand flags are set and preargs are set.
func TestCommandFlagsAndPreargs(t *testing.T) {
	resetForTesting("-global1=hello", "anotherPrearg", "command1", "-flag1=true")

	flag.String("global1", "default-global1", "Description about global1")
	c1 := &testCmd1{}
    prearg := PreArg("pa", "this is a prearg")
	On("command1", "", c1)
	Parse()
	Run()
	if !c1.run {
		t.Error("command 'command1' was expected to run, but it didn't")
	}
    if *prearg != "anotherPrearg" {
		t.Error("prearg expected to be 'anotherPrearg'")
    }
	if !*c1.flag1 {
		t.Errorf("flag1 should be set: expected true, found %v", *c1.flag1)
	}
}

// Tests try-parse
func TestTryParse1(t *testing.T) {
	resetForTesting("-global1=hello", "anotherPrearg", "command1", "-flag1=true")

	flag.String("global1", "default-global1", "Description about global1")
	c1 := &testCmd1{}
    PreArg("pa", "this is a prearg")
	On("command1", "", c1)
	res := TryParse()
	if res != TryParseOK {
		t.Error("Try parse must be TryParseOK")
	}
}

// Tests try-parse with missing prearg
func TestTryParse2(t *testing.T) {
	resetForTesting("-global1=hello")

	g1 := flag.String("global1", "default-global1", "Description about global1")
	c1 := &testCmd1{}
    PreArg("pa", "this is a prearg")
	On("command1", "", c1)
	res := TryParse()
	if res != TryParseNoPreArg {
		t.Error("Try parse must be TryParseMissingPreArg")
	}
    if *g1 != "hello" {
		t.Error("Global flag must be defined")
    }
}

// Tests try-parse with missing prearg but using the help override
func TestTryParseWithHelpPreargOverride(t *testing.T) {
	resetForTesting("help")

	c1 := &testCmd1{}
    OnHelpIgnorePreargs()
    OnHelpShowUsage()
    PreArg("pa", "this is a prearg")
	On("command1", "", c1)
	res := TryParse()
	if res != TryParseOK {
		t.Error("Try parse must be OK, was", res)
	}
}

// Tests try-parse with missing command
func TestTryParse3(t *testing.T) {
	resetForTesting("-global1=hello", "prearg")

	g1 := flag.String("global1", "default-global1", "Description about global1")
	c1 := &testCmd1{}
    prearg := PreArg("pa", "this is a prearg")
	On("command1", "", c1)
	res := TryParse()
	if res != TryParseNoCommand {
		t.Error("Try parse must be TryParseMissingPreArg")
	}
    if *g1 != "hello" {
		t.Error("Global flag must be defined")
    }
    if *prearg != "prearg" {
		t.Error("prearg expected to be 'prearg'")
    }
}

// Tests try-parse with bad command
func TestTryParse4(t *testing.T) {
	resetForTesting("-global1=hello", "prearg", "badcommand")

	g1 := flag.String("global1", "default-global1", "Description about global1")
	c1 := &testCmd1{}
    prearg := PreArg("pa", "this is a prearg")
	On("command1", "", c1)
	res := TryParse()
	if res != TryParseInvalidCommand {
		t.Error("Try parse must be TryParseInvalidCommand")
	}
    if *g1 != "hello" {
		t.Error("Global flag must be defined")
    }
    if *prearg != "prearg" {
		t.Error("prearg expected to be 'prearg'")
    }
}

// Tests try-parse with a command missing minimum arguments
func TestTryParseMinArgs1(t *testing.T) {
	resetForTesting("command1")

	c1 := &testCmd1{}
	c2 := &testCmd1{}
	On("command1", "", c1).Arguments("this", "that")
	On("command2", "", c2).Arguments("something")
	res := TryParse()
	if res != TryParseNotEnoughArgs {
		t.Error("Try parse must be TryParseNotEnoughArgs")
	}
}

// Tests try-parse with a command missing minimum arguments
func TestTryParseMinArgs2(t *testing.T) {
	resetForTesting("command1", "fla")

	c1 := &testCmd1{}
	c2 := &testCmd1{}
	On("command1", "", c1).Arguments("this", "that")
	On("command2", "", c2).Arguments("something")
	res := TryParse()
	if res != TryParseNotEnoughArgs {
		t.Error("Try parse must be TryParseNotEnoughArgs")
	}
}

// Tests try-parse with a command missing minimum arguments
func TestTryParseMinArgs3(t *testing.T) {
	resetForTesting("command2", "fla")

	c1 := &testCmd1{}
	c2 := &testCmd1{}
	On("command1", "", c1).Arguments("this", "that")
	On("command2", "", c2).Arguments("something")
	res := TryParse()
	if res != TryParseOK {
		t.Error("Try parse must be TryParseOK")
	}
    Run()
	if !c2.run {
		t.Error("command 'command2' was expected to run, but it didn't")
	}
}

// Tests try-parse with a command missing minimum arguments
func TestTryParseMinArgs4(t *testing.T) {
	resetForTesting("command1", "foo", "bar")

	c1 := &testCmd1{}
	c2 := &testCmd2{}
	On("command1", "", c1).Arguments("this", "that")
	On("command2", "", c2).Arguments("something")
	res := TryParse()
	if res != TryParseOK {
		t.Error("Try parse must be TryParseOK")
	}
    Run()
	if !c1.run {
		t.Error("command 'command1' was expected to run, but it didn't")
	}
}

// Resets os.Args and the default flag set.
func resetForTesting(args ...string) {
	os.Args = append([]string{"cmd"}, args...)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
    clearPreArgs()
    reserveHFlag = true
    helpPreargOverride = false
}

// testCmd1 is a test sub command.
type testCmd1 struct {
	flag1 *bool

	run bool
}

// Defines flags for the sub command.
func (cmd *testCmd1) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.flag1 = fs.Bool("flag1", false, "Description about flag1")
	return fs
}

// Sets the run flag.
func (cmd *testCmd1) Run(args []string) {
	cmd.run = true
}

// testCmd2 is a test sub command.
type testCmd2 struct {
	flag2 *bool

	run bool
}

// Defines flags for the sub command.
func (cmd *testCmd2) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.flag2 = fs.Bool("flag2", false, "Description about flag2")
	return fs
}

// Sets the run flag.
func (cmd *testCmd2) Run(args []string) {
	cmd.run = true
}
