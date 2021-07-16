package readline

import (
	"bytes"
	"strings"
)

// Caller type for dynamic completion
type DynamicCompleteFunc func(string) []string

type PrefixCompleterInterface interface {
	Print(prefix string, level int, buf *bytes.Buffer)
	Do(line []rune, pos int) (newLine [][]rune, length int)
	GetName() []rune
	GetChildren() []PrefixCompleterInterface
	SetChildren(children []PrefixCompleterInterface)
}

type DynamicPrefixCompleterInterface interface {
	PrefixCompleterInterface
	IsDynamic() bool
	GetDynamicNames(line []rune) [][]rune
}

type PrefixCompleter struct {
	Name     []rune
	Dynamic  bool
	Callback DynamicCompleteFunc
	Children []PrefixCompleterInterface
}

func MyTrimPrecedingSlash(x string) string {
	idx := strings.LastIndex(x, "/")
	if idx > -1 && len(x) > 1 {
		return x[idx+1:]
	}
	return x
}

func (p *PrefixCompleter) Tree(prefix string) string {
	buf := bytes.NewBuffer(nil)
	p.Print(prefix, 0, buf)
	return buf.String()
}

func Print(p PrefixCompleterInterface, prefix string, level int, buf *bytes.Buffer) {
	if strings.TrimSpace(string(p.GetName())) != "" {
		buf.WriteString(prefix)
		if level > 0 {
			buf.WriteString("├")
			buf.WriteString(strings.Repeat("─", (level*4)-2))
			buf.WriteString(" ")
		}
		buf.WriteString(string(p.GetName()) + "\n")
		level++
	}
	for _, ch := range p.GetChildren() {
		ch.Print(prefix, level, buf)
	}
}

func (p *PrefixCompleter) Print(prefix string, level int, buf *bytes.Buffer) {
	Print(p, prefix, level, buf)
}

func (p *PrefixCompleter) IsDynamic() bool {
	return p.Dynamic
}

func (p *PrefixCompleter) GetName() []rune {
	return p.Name
}

func (p *PrefixCompleter) GetDynamicNames(line []rune) [][]rune {
	var names = [][]rune{}
	for _, name := range p.Callback(string(line)) {
		names = append(names, []rune(name+" "))
	}
	return names
}

func (p *PrefixCompleter) GetChildren() []PrefixCompleterInterface {
	return p.Children
}

func (p *PrefixCompleter) SetChildren(children []PrefixCompleterInterface) {
	p.Children = children
}

func NewPrefixCompleter(pc ...PrefixCompleterInterface) *PrefixCompleter {
	return PcItem("", pc...)
}

func PcItem(name string, pc ...PrefixCompleterInterface) *PrefixCompleter {
	name += " "
	return &PrefixCompleter{
		Name:     []rune(name),
		Dynamic:  false,
		Children: pc,
	}
}

func PcItemDynamic(callback DynamicCompleteFunc, pc ...PrefixCompleterInterface) *PrefixCompleter {
	//println("2nd PLACE")
	return &PrefixCompleter{
		Callback: callback,
		Dynamic:  true,
		Children: pc,
	}
}

//This gets invoked when TAB is pressed
func (p *PrefixCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	//println("AUTOCOMPLETE WILL DO")
	return doInternal(p, line, pos, line)
}

func Do(p PrefixCompleterInterface, line []rune, pos int) (newLine [][]rune, offset int) {
	return doInternal(p, line, pos, line)
}

func doInternal(p PrefixCompleterInterface, line []rune, pos int, origLine []rune) (newLine [][]rune, offset int) {
	//println("LINE ARG:", string(line))
	//println("POS ARG:", pos)
	//println("ORIGLINE ARG:", string(origLine))
	line = runes.TrimSpaceLeft(line[:pos])
	//line = []rune(MyTrimPrecedingSlash(string(line)))
	goNext := false
	var lineCompleter PrefixCompleterInterface
	//println("JUST B4 THE LOOP")
	if p != nil {
		for _, child := range p.GetChildren() {
			childNames := make([][]rune, 1)

			childDynamic, ok := child.(DynamicPrefixCompleterInterface)
			if ok && childDynamic.IsDynamic() {
				//println("WE SHOULD BE HERE")
				//println("CHILDGETNAME: ", child.GetName())
				//println("Calling the Dynamic Names Func")
				childNames = childDynamic.GetDynamicNames(origLine)
				line = []rune(MyTrimPrecedingSlash(string(line)))
				//println("LENGTH OF childNames: ", len(childNames))
			} else {
				//println("MYSTERY BLOCK RIGHT HERE!")
				childNames[0] = child.GetName()
			}

			//println("LINE IS: ", string(line))
			for _, childName := range childNames {
				if len(line) >= len(childName) {
					if runes.HasPrefix(line, childName) {
						if len(line) == len(childName) {
							newLine = append(newLine, []rune{' '})
						} else {
							newLine = append(newLine, childName)
						}
						offset = len(childName)
						lineCompleter = child
						goNext = true
					}
				} else {
					if runes.HasPrefix(childName, line) {
						//println("SHOULD BE HERE EACH TIME")
						newLine = append(newLine, childName[len(line):])
						offset = len(line)
						lineCompleter = child
					}
				}
			}
		}

		if len(newLine) == 0 && len(origLine) > 0 && origLine[len(origLine)-1] == '/' {
			//println("GOTTA FILL NEWLINE")
			childNames := make([][]rune, 1)
			//println("LENGTH OF GETCHILDREN: ", len(p.GetChildren()))
			for _, child := range p.GetChildren() {
				childDynamic, ok := child.(DynamicPrefixCompleterInterface)
				if ok && childDynamic.IsDynamic() {
					childNames = childDynamic.GetDynamicNames(origLine)
					newLine = append(newLine, childNames...)
				}
			}
		}
	}

	if len(newLine) != 1 {
		//println("LENGTH NEWLINE: ", len(newLine))
		return
	}

	tmpLine := make([]rune, 0, len(line))
	for i := offset; i < len(line); i++ {
		if line[i] == ' ' {
			continue
		}

		tmpLine = append(tmpLine, line[i:]...)
		return doInternal(lineCompleter, tmpLine, len(tmpLine), origLine)
	}

	if goNext {
		return doInternal(lineCompleter, nil, 0, origLine)
	}
	return
}
