package j2g

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type jiration struct {
	title string
	re    *regexp.Regexp
	repl  interface{}
}

func JiraToMD(str string, attachments AttachmentMap, userMap UserMap) (string, error) {
	//* TODO
	// - Citations (buggy)
	// - Emoji

	jirations := []jiration{
		// 태그로 묶인 속성을 먼저 처리해야 한다.
		{
			title: "Remove color: unsupported in md",
			re:    regexp.MustCompile(`(?m)\{color:[^}]+\}(.*?)\{color\}`),
			repl:  "$1",
		},
		{
			title: "Remove unsupported line breaks",
			re:    regexp.MustCompile(`(\r\n|\n\r)`),
			repl:  "\n",
		},
		{
			title: "Pre-formatted text",
			re:    regexp.MustCompile(`{noformat}`),
			repl:  "```",
		},

		//! 반드시 Code Block End가 먼저 나와야 한다.
		{
			title: "Code Block End",
			re:    regexp.MustCompile(`{code}`),
			repl:  "\n```",
		},
		{
			title: "Code Block",
			re:    regexp.MustCompile(`\{code(:([a-z]+))?([:|]?(title|borderStyle|borderColor|borderWidth|bgColor|titleBGColor)=.+?)*\}`),
			repl:  "```$2",
		},

		{
			title: "Monospaced text",
			re:    regexp.MustCompile(`\{\{([^}]+)\}\}`),
			repl:  "`$1`",
		},
		{
			title: "panel into table",
			re:    regexp.MustCompile(`(?m)\{panel:title=([^}]*)\}\n?(.*?)\n?\{panel\}`),
			repl:  "\n| $1 |\n| --- |\n| $2 |",
		},

		// 이후
		{
			title: "image",
			re:    regexp.MustCompile(`(?m)!([^!|]+)(?:\|([^!|]+))?!`),
			repl: func(groups []string) (string, error) {
				_, name, _ := groups[0], groups[1], groups[2]
				if attachment, ok := attachments[name]; ok {
					return attachment.Markdown, nil
				} else {
					return "", errors.Errorf("attachment not found: %s", name)
				}
			},
		},
		{ //* Mention
			title: "Mention",
			re:    regexp.MustCompile(`(?m)\[~([^]]+)\]`),
			repl: func(groups []string) (string, error) {
				_, username := groups[0], groups[1]
				if user, ok := userMap[username]; ok {
					return "@" + user.Username, nil
				} else {
					return "", errors.Errorf("user not found: %s", username)
				}
			},
		},
		{
			title: "UnOrdered Lists",
			re:    regexp.MustCompile(`(?m)^[ \t]*(\*+)\s+`),
			repl: func(groups []string) (string, error) {
				_, stars := groups[0], groups[1]
				return strings.Repeat("  ", len(stars)-1) + "* ", nil
			},
		},
		{
			title: "Ordered Lists",
			re:    regexp.MustCompile(`(?m)^[ \t]*(#+)\s+`),
			repl: func(groups []string) (string, error) {
				_, nums := groups[0], groups[1]
				return strings.Repeat("  ", len(nums)-1) + "1. ", nil
			},
		},
		{
			title: "Headers 1-6",
			re:    regexp.MustCompile(`(?m)^h([0-6])\.(.*)$`),
			repl: func(groups []string) (string, error) {
				_, level, content := groups[0], groups[1], groups[2]
				i, _ := strconv.Atoi(level)
				return strings.Repeat("#", i) + content, nil
			},
		},
		{
			title: "Bold",
			re:    regexp.MustCompile(`\{\*\}(\S[^*]*)\{\*\}`),
			repl:  "**$1**",
		},
		{
			title: "Italic",
			re:    regexp.MustCompile(`\{\_\}(\S[^_]*)\{\_\}`),
			repl:  "*$1*",
		},
		// /* Citations (buggy)",
		// {
		// 	re:   regexp.MustCompile(`\?\?((?:.[^?]|[^?].)+)\?\?`),
		// 	repl: "<cite>$1</cite>",
		// },
		{
			title: "Inserts",
			re:    regexp.MustCompile(`\{\+\}([^+]*)\{\+\}`),
			repl:  "<ins>$1</ins>",
		},
		{
			title: "Superscript",
			re:    regexp.MustCompile(`\^([^^]*)\^`),
			repl:  "<sup>$1</sup>",
		},
		{
			title: "Subscript",
			re:    regexp.MustCompile(`~([^~]*)~`),
			repl:  "<sub>$1</sub>",
		},

		//! Rule은 Strikethrough보다 먼저 나와야 한다.
		{
			title: "Rule",
			re:    regexp.MustCompile(`-{4,}`),
			repl:  "---",
		},
		{
			title: "Strikethrough",
			re:    regexp.MustCompile(`(\s+)-(\S+.*?\S)-(\s+)`),
			repl:  "$1~~$2~~$3",
		},
		// { //* n-named Links
		// 	re:   regexp.MustCompile(`(?U)\[([^|]+?)\]`),
		// 	repl: "<$1>",
		// },
		{
			title: "Named Links",
			re:    regexp.MustCompile(`\[(.+?)\|(.+?)\]`),
			repl:  "[$1]($2)",
		},
		{
			title: "Single Paragraph Blockquote",
			re:    regexp.MustCompile(`(?m)^bq\.\s+`),
			repl:  "> ",
		},
		{
			title: "table",
			re:    regexp.MustCompile(`(?m)\|\|(([^|\n\r]+)\|\|)+?(\r?\n\|(([^|\n\r]+?)\|)+)+`),
			repl: func(groups []string) (string, error) {
				reHeader := regexp.MustCompile(`(?m)^(\|\|(?:[^|\n\r]+?\|\|)+)`)
				reRows := regexp.MustCompile(`(?m)\r?\n(\|(?:[^|\n\r]+?\|)+)`)

				headerMatches := reHeader.FindAllStringSubmatch(groups[0], -1)
				rowMatches := reRows.FindAllStringSubmatch(groups[0], -1)

				if len(headerMatches) == 0 || len(rowMatches) == 0 {
					return "", errors.New("table header or rows not found")
				}

				headerstr := headerMatches[0][1]
				rowStrs := []string{}
				for _, rowMatch := range rowMatches {
					rowStrs = append(rowStrs, rowMatch[1])
				}

				// Trim | on header and split into columns

				headerColumns := strings.Split(strings.Trim(headerstr, "|"), "||")
				rows := [][]string{}
				for _, rowStr := range rowStrs {
					rowColumns := strings.Split(strings.Trim(rowStr, "|"), "|")
					if len(rowColumns) != len(headerColumns) {
						return "", errors.Errorf("row column count not match: %d != %d", len(rowColumns), len(headerColumns))
					}

					rows = append(rows, rowColumns)
				}

				result := fmt.Sprintf("| %s |\n", strings.Join(headerColumns, " | "))
				result += "|" + strings.Repeat(" --- |", len(headerColumns)) + "\n"
				for _, row := range rows {
					result += fmt.Sprintf("| %s |\n", strings.Join(row, " | "))
				}

				// trim last \n
				result = result[:len(result)-1]

				return result, nil
			},
		},
	}

	for _, jiration := range jirations {
		// log.Debugf("Substituting '%s'", jiration.title)
		switch v := jiration.repl.(type) {
		case string:
			str = jiration.re.ReplaceAllString(str, v)
		case func([]string) (string, error):
			newStr, err := replaceAllStringSubmatchFunc(jiration.re, str, v)
			if err != nil {
				return "", errors.Wrap(err, "JiraToMD")
			} else {
				str = newStr
			}
		default:
			return "", errors.Errorf("unknown type: %v", v)
		}
	}
	return str, nil
}

func replaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) (string, error)) (string, error) {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			if v[i] == -1 {
				groups = append(groups, "")
				continue
			}
			groups = append(groups, str[v[i]:v[i+1]])
		}

		r, err := repl(groups)
		if err != nil {
			return "", errors.Wrap(err, "replaceAllStringSubmatchFunc")
		}

		result += str[lastIndex:v[0]] + r
		lastIndex = v[1]
	}

	return result + str[lastIndex:], nil
}
