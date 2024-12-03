package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type MetaCommandResult int
type StatementType int
type PrepareResult int

const (
	META_COMMAND_SUCCESS MetaCommandResult = iota
	META_COMMAND_UNRECOGNIZED_COMMAND
)

const (
	STATEMENT_INSERT StatementType = iota
	STATEMENT_SELECT
)

const (
	PREPARE_SUCCESS PrepareResult = iota
	PREPARE_UNRECOGNIZED_STATEMENT
)

type Statement struct {
	statementType StatementType
}

type CLI struct {
	reader *bufio.Reader
	buffer       string
	bufferLength int
	inputLength  int
}

func NewCLI() *CLI {
	return &CLI{
		reader: bufio.NewReader(os.Stdin),
		buffer:       "",
		bufferLength: 0,
		inputLength:  0,
	}
}

func (c *CLI) Start() {
	c.printHelp()
	
	for {
		c.printPrompt()
		c.readInput()

		if len(c.buffer) < 1 {
			return
		}

		if c.buffer[0] == '.' {
			switch c.doMetaCommand() {
			case META_COMMAND_SUCCESS:
				continue
			case META_COMMAND_UNRECOGNIZED_COMMAND:
				fmt.Printf("Unrecognized command '%s'\n", c.buffer)
				continue
			}
		}
		
		statement := Statement{}
		
		switch c.prepareStatement(&statement) {
		case PREPARE_UNRECOGNIZED_STATEMENT:
			fmt.Printf("Unrecognized keyword at start of '%s'.\n", c.buffer)
			continue
		}

		c.executeStatement(&statement)
		fmt.Println("Executed.")
	}
}

func (c *CLI) printHelp() {
	fmt.Println(`
	TinyDB CLI
	Available Commands:
	...
	`)
}

func (c *CLI) printPrompt() {
	fmt.Print("tinydb > ")
}

func (c *CLI) readInput() {
	line, err := c.reader.ReadString('\n')

	if err != nil {
		fmt.Println("Error reading input: ", err)
		os.Exit(0)
	}

	line = strings.TrimRight(line, "\r\n")
	c.buffer = line
	c.bufferLength = len(line)
	c.inputLength = len(line)
}

func (c *CLI) doMetaCommand() MetaCommandResult {
	if strings.Compare(c.buffer, ".exit") == 0 {
		os.Exit(0)
	} else {
		return META_COMMAND_UNRECOGNIZED_COMMAND
	}

	return META_COMMAND_UNRECOGNIZED_COMMAND
}

func (c *CLI) prepareStatement(statement *Statement) PrepareResult {
	if len(c.buffer) >= 6 && strings.Compare(c.buffer[:6], "insert") == 0 {
		statement.statementType = STATEMENT_INSERT
		return PREPARE_SUCCESS
	}

	if strings.HasPrefix(c.buffer, "select") {
		statement.statementType = STATEMENT_SELECT
		return PREPARE_SUCCESS
	}

	return PREPARE_UNRECOGNIZED_STATEMENT
}

func (c *CLI) executeStatement(statement *Statement) {
	switch statement.statementType {
	case STATEMENT_INSERT:
		fmt.Printf("This is where we would do an insert. \n")
	case STATEMENT_SELECT:
		fmt.Printf("This is where we would do an select. \n")
	}
}