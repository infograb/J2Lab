package j2g

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type jiration struct {
	title string
	re    *regexp.Regexp
	repl  interface{}
}

func JiraToMD(str string, attachments AttachmentMap, userMap UserMap) (string, []string, error) {
	usedAttachments := []string{}

	jirations := []jiration{

		{
			title: "Remove unsupported line breaks",
			re:    regexp.MustCompile(`(\r\n|\n\r)+`),
			repl:  "\n\n",
		},

		//* Code Blocks
		{
			title: "Code Block",
			re:    regexp.MustCompile(`(?m)\{code(?::(.+))?\}\n?((?:.|\s)*?)\n?\{code\}`),
			// repl:  "```$1\n$2\n```",
			repl: func(groups []string) (string, error) {
				_, metaStr, content := groups[0], groups[1], groups[2]

				lang := ""
				metadata := strings.Split(metaStr, "|")
				for _, v := range metadata {
					match := regexp.MustCompile(`(?m)^ *([^=]+)(?:=(.*?))? *$`).FindStringSubmatch(v)
					if len(match) == 3 {
						if match[2] == "" {
							lang = match[1]
						} else {
							key, value := match[1], match[2]
							switch key {
							case "title":
								arr := strings.Split(value, ".")
								lang = arr[len(arr)-1]
							}
						}
					}
				}

				return fmt.Sprintf("```%s\n%s\n```", lang, content), nil
			},
		}, {
			title: "Noformat Block",
			re:    regexp.MustCompile(`(?m)\{noformat\}\n?((?:.|\s)*?)\n?\{noformat\}`),
			repl:  "```\n$1\n```",
		},

		//* Text Effects
		{
			title: "Strong to Bold around brackets",
			re:    regexp.MustCompile(`(?m)\{\*\}([^\s](?:[^\n\r*]*?[^\s])?)\{\*\}`),
			repl:  "**$1**",
		}, {
			title: "Italic to Italic around brackets",
			re:    regexp.MustCompile(`(?m)\{_\}([^\s](?:[^\n\r_]*?[^\s])?)\{_\}`),
			repl:  "<i>$1</i>",
		}, {
			title: "Deleted to Strikethrough around brackets",
			re:    regexp.MustCompile(`(?m)\{-\}([^\s](?:[^\n\r*]*?[^\s])?)\{-\}`),
			repl:  "<del>$1</del>",
		}, {
			title: "Inserted to Underline around brackets",
			re:    regexp.MustCompile(`(?m)\{\+\}([^\s](?:[^\n\r*]*?[^\s])?)\{\+\}`),
			repl:  "<ins>$1</ins>",
		}, {
			title: "Superscript around brackets",
			re:    regexp.MustCompile(`(?m)\{\^\}([^\s](?:[^\n\r*]*?[^\s])?)\{\^\}`),
			repl:  "<sup>$1</sup>",
		}, {
			title: "Subscript around brackets",
			re:    regexp.MustCompile(`(?m)\{~\}([^\s](?:[^\n\r*]*?[^\s])?)\{~\}`),
			repl:  "<sub>$1</sub>",
		},

		{
			title: "Strong to Bold",
			re:    regexp.MustCompile(`(?m)(^| |\W)\*([^\s](?:[^\n\r*]*?[^\s])?)\*($| |\W)`),
			repl:  "$1**$2**$3",
		}, {
			title: "Italic to Italic",
			re:    regexp.MustCompile(`(?m)(^| |\W)_([^\s](?:[^\n\r_]*?[^\s])?)_($| |\W)`),
			repl:  "$1<i>$2</i>$3",
		}, {
			title: "Citation to Italic",
			re:    regexp.MustCompile(`(?m)(^| |\W)\?\?([^\s](?:[^\n\r?]*?[^\s])?)\?\?($| |\W)`),
			repl:  "$1<i>$2</i>$3",
		}, {
			title: "Deleted to Strikethrough",
			re:    regexp.MustCompile(`(?m)($| |\W)-([^\s](?:[^\n\r*]*?[^\s])?)-($| |\W)`),
			repl: func(groups []string) (string, error) {
				all, before, content, after := groups[0], groups[1], groups[2], groups[3]
				if len(content) == strings.Count(content, "-") {
					return all, nil
				}
				return before + "<del>" + content + "</del>" + after, nil
			},
		}, {
			title: "Inserted to Underline",
			re:    regexp.MustCompile(`(?m)(^| |\W)\+([^\s](?:[^\n\r+]*?[^\s])?)\+($| |\W)`),
			repl:  "$1<ins>$2</ins>$3",
		}, {
			title: "Superscript",
			re:    regexp.MustCompile(`(?m)(^| |\W)\^([^\s](?:[^\n\r\^]*?[^\s])?)\^($| |\W)`),
			repl:  "$1<sup>$2</sup>$3",
		}, {
			title: "Subscript",
			re:    regexp.MustCompile(`(?m)(^| |\W)~([^\s](?:[^\n\r~]*?[^\s])?)~($| |\W)`),
			repl:  "$1<sub>$2</sub>$3",
		}, {
			title: "Monospaced text Inline Code",
			re:    regexp.MustCompile(`(?m)(^| |\W)\{\{([^\s](?:[^\n\r]*?[^\s])?)\}\}($| |\W)`),
			repl:  "$1`$2`$3",
		}, {
			title: "Blockquote to Blockquote",
			re:    regexp.MustCompile(`(?m)(^| |\W)bq\.(.*)$`),
			repl:  "> $1",
		}, {
			title: "Quote to Blockquote",
			re:    regexp.MustCompile(`\{quote\}((?:.|\s)*?)\{quote\}`),
			repl: func(groups []string) (string, error) {
				content, _ := groups[0], groups[1]
				content = strings.ReplaceAll(content, "{quote}", "")
				content = strings.ReplaceAll(content, "\n", "\n> ")
				return "> " + content, nil
			},
		}, {
			title: "Color to None",
			re:    regexp.MustCompile(`(?m)\{color:.+?\}((?:.|\s)*?)\{color\}`),
			repl:  "$1",
		},

		//* Text Breaks
		{
			title: "// to \n\n",
			re:    regexp.MustCompile(`(?m)^([^\\]*)\\\\([^\\]*)$`),
			repl:  "$1\n\n$2",
		}, { //! Dash 변환은 반드시 길이가 짧은 순서로 진행되어야 한다.
			title: "-- to –(en dash)",
			re:    regexp.MustCompile(`(?m)(^| )(?:--)($| )`),
			repl:  "$1–$2",
		}, {
			title: "--- to —(em dash)",
			re:    regexp.MustCompile(`(?m)(^| )(?:---)($| )`),
			repl:  "$1—$2",
		}, {
			title: "---- to Ruler",
			re:    regexp.MustCompile(`(?m)^( *)?(?:-{4,})( *)?$`),
			repl:  "$1---$2",
		},

		//* Links
		{
			title: "Achor to Anchor Link", // TODO but removed currently
			re:    regexp.MustCompile(`(?m)\[(?:(.+)\|)?#([^|\n\r]+)\]`),
			repl: func(groups []string) (string, error) {
				_, name, anchor := groups[0], groups[1], groups[2]
				if name == "" {
					name = anchor
				}
				// return "[" + name + "](#" + anchor + ")", nil
				return name, nil
			},
		}, {
			title: "Link to Link",
			re:    regexp.MustCompile(`(?m)\[(?:(.+)\|)?([^#][^|\n\r]+)\]`),
			repl: func(groups []string) (string, error) {
				all, name, link := groups[0], groups[1], groups[2]
				if name == "" {
					name = link
				}

				if strings.HasPrefix(link, "http") {
					return "[" + name + "](" + link + ")", nil
				}

				return all, nil
			},
		}, {
			title: "Mailto to Mailto Link",
			re:    regexp.MustCompile(`(?m)\[(?:([^\n\r\|]+)\|)?mailto:([^\s]+)\]`),
			repl: func(groups []string) (string, error) {
				_, name, mail := groups[0], groups[1], groups[2]
				if name == "" {
					name = mail
				}
				return fmt.Sprintf("[%s✉️](mailto:%s)", name, mail), nil
			},
		}, {
			title: "Anchor name to Anchor Link", // TODO but removed currently
			re:    regexp.MustCompile(`(?m)(^| |\W)\{anchor:.+\}($| |\W)`),
			repl:  "$1$3",
		}, {
			title: "Mention to Mention",
			re:    regexp.MustCompile(`(?m)^(.*)\[~([^]]+)\](.*)$`),
			repl: func(groups []string) (string, error) {
				_, before, username, after := groups[0], groups[1], groups[2], groups[3]
				if user, ok := userMap[username]; ok {
					if before != "" && before[len(before)-1] != ' ' {
						before += " "
					}
					if after != "" && after[0] != ' ' {
						after = " " + after
					}
					return before + "@" + user.Username + after, nil
				} else {
					return "", errors.Errorf("user not found: %s", username)
				}
			},
		},

		//* Lists
		{
			title: "All List to All List",
			re:    regexp.MustCompile(`(?m)(?:^| +)(\*+|\-+|#+)(?: )+([^\s].*)(\n+|$)`),
			repl: func(groups []string) (string, error) {
				_, bullets, content, breaks := groups[0], groups[1], groups[2], groups[3]
				depth := len(bullets) - 1
				if breaks != "" {
					breaks = strings.Repeat("\n", len(breaks)/2)
				}

				var bulletType string
				switch bullets[len(bullets)-1] {
				case '*', '-':
					bulletType = "*"
				case '#':
					bulletType = "1."
				}

				return strings.Repeat("  ", depth) + bulletType + " " + content + breaks, nil
			},
		},

		//* Attachments
		{
			title: "Image Attachment",
			re:    regexp.MustCompile(`(?m)!([^!|\s]+)(?:\|((?:[^!|\s]| )+))?!`),
			repl: func(groups []string) (string, error) {
				_, name, metadata := groups[0], groups[1], groups[2]
				if attachment, ok := attachments[name]; ok {
					usedAttachments = append(usedAttachments, name)

					widthMatch := regexp.MustCompile(`width=(\d+)`).FindStringSubmatch(metadata)
					heightMatch := regexp.MustCompile(`height=(\d+)`).FindStringSubmatch(metadata)

					metadataStr := ""
					width, height := 0, 0
					if len(widthMatch) > 0 {
						width, _ = strconv.Atoi(widthMatch[1])
						metadataStr += fmt.Sprintf(" width=\"%s\"", widthMatch[1])
					}
					if len(heightMatch) > 0 {
						height, _ = strconv.Atoi(heightMatch[1])
						metadataStr += fmt.Sprintf(" height=\"%s\"", heightMatch[1])
					}

					if width > 0 || height > 0 {
						return fmt.Sprintf(`<img src="%s" alt="%s"%s>`, attachment.URL, attachment.Alt, metadataStr), nil
					}

					return attachment.Markdown, nil
				} else {
					log.Debugf("attachment not found: %s", name)
					return fmt.Sprintf("![%s](%s)", name, name), nil
				}
			},
		},
		{
			title: "File Attachment",
			re:    regexp.MustCompile(`(?m)\[\^(.+)\]`),
			repl: func(groups []string) (string, error) {
				_, name := groups[0], groups[1]
				if attachment, ok := attachments[name]; ok {
					usedAttachments = append(usedAttachments, name)
					return fmt.Sprintf("[%s](%s)", attachment.Alt, attachment.URL), nil
				} else {
					log.Debugf("attachment not found: %s", name)
					return fmt.Sprintf("[^%s]", name), nil
				}
			},
		},

		{ //! Heading은 반드시 List 이후에 나와야 한다.
			title: "Headers 1-6",
			re:    regexp.MustCompile(`(?m)^ *h([0-6])\. (.*)$`),
			repl: func(groups []string) (string, error) {
				_, level, content := groups[0], groups[1], groups[2]
				i, _ := strconv.Atoi(level)
				return strings.Repeat("#", i) + " " + content, nil
			},
		},

		//* Table

		{
			title: "table",
			re:    regexp.MustCompile(`(?m)\|\|(([^|\n\r]+)\|\|)+?(\n\|(([^|\n\r]+?)\|)+)+`),
			repl: func(groups []string) (string, error) {
				reHeader := regexp.MustCompile(`(?m)^(\|\|(?:[^|\n\r]+?\|\|)+)`)
				reRows := regexp.MustCompile(`(?m)\n(\|(?:[^|\n\r]+?\|)+)`)

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

				return result, nil
			},
		}, {
			title: "panel into table",
			re:    regexp.MustCompile(`(?m)\{panel(?::(.+))?\}\n?((?:.|\s)*?)\n?\{panel\}`),
			repl: func(groups []string) (string, error) {
				_, metaStr, content := groups[0], groups[1], groups[2]

				title := ""
				metadata := strings.Split(metaStr, "|")
				for _, v := range metadata {
					match := regexp.MustCompile(`(?m)^ *([^=]+)(?:=(.*?))? *$`).FindStringSubmatch(v)
					if len(match) == 3 && match[2] != "" {
						key, value := match[1], match[2]
						switch key {
						case "title":
							title = value
						}
					}
				}

				content = strings.Trim(content, "\n")
				content = regexp.MustCompile(`(?m)\n+`).ReplaceAllString(content, "<br/>")

				return fmt.Sprintf("\n| %s |\n| --- |\n| %s |\n", title, content), nil
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
				return "", nil, errors.Wrap(err, "JiraToMD")
			} else {
				str = newStr
			}
		default:
			return "", nil, errors.Errorf("unknown type: %v", v)
		}
	}
	return str, usedAttachments, nil
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
