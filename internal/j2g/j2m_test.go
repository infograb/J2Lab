/*
 * This file is part of the InfoGrab project.
 *
 * Copyright (C) 2023 InfoGrab
 *
 * This program is free software: you can redistribute it and/or modify it
 * it is available under the terms of the GNU Lesser General Public License
 * by the Free Software Foundation, either version 3 of the License or by the Free Software Foundation
 * (at your option) any later version.
 */
package j2g

import (
	"testing"

	"github.com/xanzy/go-gitlab"
)

var testCases = []struct {
	input       string
	expected    string
	description string
}{
	{
		input:       "Hello World",
		expected:    "Hello World",
		description: "No change",
	}, {
		input:       "{*}bold{*}",
		expected:    "**bold**",
		description: "Bold",
	}, {
		input:       "{_}italic{_}",
		expected:    "*italic*",
		description: "Italic",
	}, {
		input:       "{+}underline{+}",
		expected:    "<ins>underline</ins>",
		description: "Underline",
	}, {
		input:       "{color:#ff0000}red{color}",
		expected:    "red",
		description: "Color",
	}, {
		input:       "What [~jeff] said!",
		expected:    "What @infograb-jeff said!",
		description: "User mention",
	}, {
		input:       "{*}bold{*} [~admin] said!",
		expected:    "**bold** @dexter.shin said!",
		description: "User mention + bold",
	}, {
		input:       "!SCR-20230906-ofnz.png!",
		expected:    "![SCR-20230906-ofnz.png](https://jira.infograb.net/secure/attachment/10000/SCR-20230906-ofnz.png)",
		description: "Image",
	}, {
		input:       "!SCR-20230906-oflk.png|thumbnail!",
		expected:    "![SCR-20230906-oflk.png](https://jira.infograb.net/secure/attachment/10000/SCR-20230906-oflk.png)",
		description: "Image with thumbnail",
	}, {
		input:       "{color:#ff0000}h1. Header 1{color}",
		expected:    "# Header 1",
		description: "Header 1 with Color",
	}, {
		input:       "{color:#ff0000}h1. Header 1{color}",
		expected:    "# Header 1",
		description: "Header 1",
	},
	{
		input:       "asdff\n||표머리일||표머리2||\r\n|표내용일|표내용2|",
		expected:    "asdff\n| 표머리일 | 표머리2 |\n| --- | --- |\n| 표내용일 | 표내용2 |",
		description: "Table",
	},
	{
		input:       "||표머리일||표머리1||\r\n|표내용이|표내용2|\r\n|표내용삼|표내용3|",
		expected:    "| 표머리일 | 표머리1 |\n| --- | --- |\n| 표내용이 | 표내용2 |\n| 표내용삼 | 표내용3 |",
		description: "Table",
	},
	{
		input: "\n\r||표머리일||표머리1||\r\n|표내용이|표내용2|\r\n|표내용삼|표내용3|\r\n",
		expected: `
| 표머리일 | 표머리1 |
| --- | --- |
| 표내용이 | 표내용2 |
| 표내용삼 | 표내용3 |
`,
		description: "Table 3",
	},
	{
		input:       "\n\r----\n\r",
		expected:    "\n---\n",
		description: "Rule",
	},
	{
		input:       "\n\r-asdf-\n\r",
		expected:    "\n~~asdf~~\n",
		description: "Strikethrough",
	},
	{
		input:       "\n\r{quote}asdf{quote}\n\r",
		expected:    "\n\n> asdf\n",
		description: "One line blockquote",
	},
	{
		input:       "\n\r{quote}asdf\n\rasdf\n\r{quote}\n\r",
		expected:    "\n\n> asdf\n> asdf\n> \n",
		description: "Multi line blockquote",
	},
	{
		description: "Headings 1",
		input:       "h1. Headin{*}g 1{*}",
		expected:    "# Headin**g 1**",
	},

	// * 여기서부터 공식 문서대로 하나씩 (https://jira.atlassian.com/secure/WikiRendererHelpAction.jspa?section=all)
	{
		description: "Headings 1",
		input:       "h1. Header 1",
		expected:    "# Header 1",
	}, {
		description: "Headings 2",
		input:       "h2. Header 2",
		expected:    "## Header 2",
	}, {
		description: "Headings 3",
		input:       "h3. Header 3",
		expected:    "### Header 3",
	}, {
		description: "Headings 4",
		input:       "h4. Header 4",
		expected:    "#### Header 4",
	}, {
		description: "Headings 5",
		input:       "h5. Header 5",
		expected:    "##### Header 5",
	}, {
		description: "Headings 6",
		input:       "h6. Header 6",
		expected:    "###### Header 6",
	},
}

var attachments = AttachmentMap{
	"SCR-20230906-oflk.png": &Attachment{
		Markdown:  "![SCR-20230906-oflk.png](https://jira.infograb.net/secure/attachment/10000/SCR-20230906-oflk.png)",
		Filename:  "SCR-20230906-oflk.png",
		CreatedAt: "2019-09-06T09:00:00+09:00",
	},
	"SCR-20230906-ofnz.png": &Attachment{
		Markdown:  "![SCR-20230906-ofnz.png](https://jira.infograb.net/secure/attachment/10000/SCR-20230906-ofnz.png)",
		Filename:  "SCR-20230906-ofnz.png",
		CreatedAt: "2019-09-06T09:00:00+09:00",
	},
}

var userMap = UserMap{
	"admin": &gitlab.User{
		ID:       13871121,
		Username: "dexter.shin",
	},
	"jeff": &gitlab.User{
		ID:       12709793,
		Username: "infograb-jeff",
	},
}

func TestJiraToMD(t *testing.T) {
	for _, tc := range testCases {
		actual, _, err := JiraToMD(tc.input, attachments, userMap)
		if err != nil {
			t.Errorf("Error: %s", err)
		}

		if actual != tc.expected {
			t.Errorf("JiraToMD('%s'): expected '%s', actual '%s'", tc.input, tc.expected, actual)
		}
	}
}
