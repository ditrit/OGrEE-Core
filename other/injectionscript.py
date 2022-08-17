#!/usr/bin/env python3
import os
import re

# Load y.go into string var 'm'
m = ""
with open(os.getcwd()+"/y.go", "r") as f:
    contents = f.readlines()
    for i in contents:
        m += i

# Inject error message in yyParse() func
x = m.find("yylex.Error(msg)")

goCode='\t\t\tif cmd.State.DebugLvl > 0 {\n\t\t\t\tif cmd.State.ScriptCalled == true {\n\t\t\t\t\tprintln("File:",filepath.Base(cmd.GetScriptPath()) )\n\t\t\t\t\tprintln("Line:", cmd.GetLineNumber())\n\t\t\t\t}\n\t\t\t}'

m = m.replace("yylex.Error(msg)",
              "println(\"OGREE: Unrecognised command!\")\n"+goCode+"\n\t\t\tl.GetWarningLogger().Println(\"Unknown Command\")\t\t\t/*yylex.Error(msg)*/")


# Inject a static Analyser func, yyAnalyse()
# This function is the same as yyParse
# except that the parse actions are replaced
# with 'return 0'. We use this to indicate successful
# parse analysis
yyParseIdx = m.find("func yyParse(yylex yyLexer) int {")
yyAnalyse = str(m[yyParseIdx:])
matcher = re.compile("(yyVAL\.\w+ = )")
yyAnalyse = matcher.sub("return 0\n\t\t\t\\1", yyAnalyse)
yyAnalyse = yyAnalyse.replace(
    "func yyParse(yylex yyLexer) int {", "\n\nfunc yyAnalyse(yylex yyLexer) int {")
m += yyAnalyse


# Finally write our changes to file
with open(os.getcwd()+"/y.go", "w") as f:
    f.write(m)
