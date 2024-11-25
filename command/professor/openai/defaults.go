package openai

import "github.com/openai/openai-go"

var (
	defaultModel   = openai.ChatModelGPT4o
	defaultContext = "Explain the given command and give me your answer using markdown; this explanation should contain the following sections, summary, breakdown, example of use and cautions; these sections encode them as markdown headings. The command can contain parameters of the form {{.name}} where name is the name of the parameter, which are meant to be replaced. When formatting the code in the explanation, use fish as the format. Don't mention how to replace the parameters. Here is the command:%s"
)
