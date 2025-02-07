package util

import (
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const DEFAULT_STACK_BUFFER_SIZE = 1 << 18 // 256k

var nameRegex = regexp.MustCompile(`^(.*) \[(.*)\]:$`)
var fileLineRegex = regexp.MustCompile(`^(.*):(\d+)(\s([\+\-]0x\S+))?$`)

type CallStackRecord struct {
	Call    string `json:"call"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Pointer string `json:"pointer"`
}

type CallStack struct {
	Name    string            `json:"name"`
	Status  string            `json:"status"`
	Records []CallStackRecord `json:"records"`
}

func GetCurrentStackTrace() *CallStack {
	traces := getStackTraces(false)
	if len(traces) != 1 {
		return nil
	}

	return traces[0]
}

func GetAllStackTraces() []*CallStack {
	return getStackTraces(true)
}

func getStackTraces(includeAll bool) []*CallStack {
	length := DEFAULT_STACK_BUFFER_SIZE
	var buffer []byte = nil

	for {
		buffer = make([]byte, length)
		writeLength := runtime.Stack(buffer, includeAll)
		if writeLength == length {
			// The write filled the whole buffer, there may be more.
			length *= 8
		} else {
			// The write was short, we got everything.
			buffer = buffer[0:writeLength]
			break
		}
	}

	rawStacks := strings.Split(string(buffer), "\n\n")
	stacks := make([]*CallStack, 0, len(rawStacks))

	for _, rawStack := range rawStacks {
		stack := parseStack(rawStack)
		if stack != nil {
			stacks = append(stacks, stack)
		}
	}

	return stacks
}

func parseStack(rawStack string) *CallStack {
	lines := strings.Split(strings.TrimSpace(rawStack), "\n")
	if len(lines) == 0 {
		return nil
	}

	match := nameRegex.FindStringSubmatch(lines[0])
	if match == nil {
		return nil
	}

	stack := CallStack{
		Name:    match[1],
		Status:  match[2],
		Records: make([]CallStackRecord, 0, len(lines)/2),
	}

	lines = lines[1:]
	for index := 0; index < (len(lines) / 2); index++ {
		callLine := strings.TrimSpace(lines[(index*2)+0])
		fileLine := strings.TrimSpace(lines[(index*2)+1])

		record := CallStackRecord{
			Call: callLine,
			File: fileLine,
		}

		match := fileLineRegex.FindStringSubmatch(fileLine)
		if match != nil {
			record.File = match[1]
			record.Line, _ = strconv.Atoi(match[2])
			record.Pointer = match[4]
		}

		stack.Records = append(stack.Records, record)
	}

	return &stack
}
