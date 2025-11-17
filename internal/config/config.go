package config

import (
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/utils"
)

var strPtr = utils.Ptr[string]

type syntax = proto.Syntax
type highlight = proto.Highlight

type Config struct {
	Syntax    []syntax
	Highlight []highlight
}

func cap(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}

func low(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c - 'A' + 'a'
	}
	return c
}

func lowerCapSCREAM(word string) []string {
	n := len(word)
	upper := make([]byte, n)
	lower := make([]byte, n)
	capitalize := make([]byte, n)

	for i := range n {
		upper[i] = cap(word[i])
		lower[i] = low(word[i])
	}

	copy(capitalize, lower)
	if n > 0 {
		capitalize[0] = cap(word[0])
	}

	return []string{
		string(lower),
		string(capitalize),
		string(upper),
	}
}

var DefaultConfig = Config{
	Syntax: []syntax{
		// symbols
		{
			Group:   "LogSymbol",
			Pattern: proto.MustCompile(`[!@#$%^&*;:?]`),
		},

		// separators
		{
			Group:   "LogSeparatorLine",
			Pattern: proto.MustCompile(`(-{3,}|={3,}|#{3,}|\*{3,}|<{3,}|>{3,})`),
		},

		// " Strings
		// " ------------------------------
		// syn region LogString      start=/"/  end=/"/  end=/$/  skip=/\\./
		// syn region LogString      start=/`/  end=/`/  end=/$/  skip=/\\./
		// " Quoted strings, but no match on quotes like `don't`, possessive `s'` and `'s`
		// syn region LogString      start=/\(s\)\@<!'\(s \|t \)\@!/  end=/'/  end=/$/  skip=/\\./

		// numbers
		{
			Group:   "LogNumber",
			Pattern: proto.MustCompile(`\b\d+\b`),
		},
		{
			Group:   "LogNumberFloat",
			Pattern: proto.MustCompile(`\b\d+\.\d+([eE][+-]?\d+)?\b`),
		},
		{
			Group:   "LogNumberBin",
			Pattern: proto.MustCompile(`\b0[bB][01]+\b`),
		},
		{
			Group:   "LogNumberOctal",
			Pattern: proto.MustCompile(`\b0[oO]?[0-7]+\b`),
		},
		{
			Group:   "LogNumberHex",
			Pattern: proto.MustCompile(`\b0[xX][0-9a-fA-F]+\b`),
		},
		{
			Group:   "LogNumberHex",
			Pattern: proto.MustCompile(`\b[0-9a-fA-F]{4,}\b`),
		},

		// constants
		{
			Group:    "LogBool",
			Keywords: utils.JoinSlices(lowerCapSCREAM("true"), lowerCapSCREAM("false")),
		},
		{
			Group:    "LogNull",
			Keywords: lowerCapSCREAM("null"),
		},

		// date and time
		// MM-DD, DD-MM, MM/DD, DD/MM
		{
			Group:   "LogDate",
			Pattern: proto.MustCompile(`\b\d{2}[-/]\d{2}\b`),
		},
		// YYYY-MM-DD, YYYY/MM/DD, DD-MM-YYYY, DD/MM/YYYY
		{
			Group:   "LogDate",
			Pattern: proto.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`),
		},
		{
			Group:   "LogDate",
			Pattern: proto.MustCompile(`\b\d{4}/\d{2}/\d{2}\b`),
		},
		{
			Group:   "LogDate",
			Pattern: proto.MustCompile(`\b\d{2}-\d{2}-\d{4}\b`),
		},
		{
			Group:   "LogDate",
			Pattern: proto.MustCompile(`\b\d{2}/\d{2}/\d{4}\b`),
		},
		// RFC3339
		{
			Group:   "LogDate",
			Pattern: proto.MustCompile(`(?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?`),
		},
		// 'Dec 31', 'Dec 31, 2023', 'Dec 31 2023'
		{
			Group:   "LogDate",
			Pattern: proto.MustCompile(`\b(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{1,2}(,? [0-9]{4})?\b`),
		},
		// '31-Dec-2023', '31 Dec 2023'
		{
			Group:   "LogDate",
			Pattern: proto.MustCompile(`\b\d{1,2}[- ](Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[- ]\d{4}\b`),
		},
		// weekday string
		{
			Group:    "LogWeekdayStr",
			Keywords: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		},
		// 12:34:56, 12:34:56.700000
		{
			Group:   "LogTime",
			Pattern: proto.MustCompile(`\b\d{2}:\d{2}:\d{2}(,\d{1,6}|\.\d{1,6})?\b`),
		},
		// AM / PM
		{
			Group:    "LogTimeAMPM",
			Keywords: []string{"AM", "am", "PM", "pm"},
		},

		// Duration e.g. 10d20h30m40s, 123.456s, 123ms, 456us, 789ns
		{
			Group:   "LogDuration",
			Pattern: proto.MustCompile(`\b((\d+d)?(\d+h)?(\d+m)?\d+(\.\d+)?[mun]?s)\b`),
		},

		// Objects
		{
			Group:   "LogUrl",
			Pattern: proto.MustCompile(`\bhttps?://\S+`),
		},
		{
			Group:   "LogMacAddr",
			Pattern: proto.MustCompile(`\b[0-9a-fA-F]{2}([:-][0-9a-fA-F]{2}){5}\b`),
		},
		{
			Group:   "LogIPv4",
			Pattern: proto.MustCompile(`\b\d{1,3}(\.\d{1,3}){3}(\/\d+)?\b`),
		},
		{
			Group:   "LogIPv6",
			Pattern: proto.MustCompile(`\b[0-9a-fA-F]{1,4}(:[0-9a-fA-F]{1,4}){7}(\/\d+)?\b`),
		},
		{
			Group:   "LogUUID",
			Pattern: proto.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`),
		},
		{
			Group:   "LogMD5",
			Pattern: proto.MustCompile(`\b[0-9a-fA-F]{32}\b`),
		},
		{
			Group:   "LogSHA",
			Pattern: proto.MustCompile(`\b([0-9a-fA-F]{40}|[0-9a-fA-F]{56}|[0-9a-fA-F]{64}|[0-9a-fA-F]{96}|[0-9a-fA-F]{128})\b`),
		},

		// POSIX-style path    e.g. '/var/log/system.log', './run.sh', '../a/b', '~/c'.
		{
			Group:   "LogPath",
			Pattern: proto.MustCompile(`(?:\/|\.{1,2}\/)?(?:[A-Za-z0-9._-]+\/)*[A-Za-z0-9._-]+\/?`),
		},

		// log levels
		{
			Group:    "LogLvFatal",
			Keywords: lowerCapSCREAM("fatal"),
		},
		{
			Group:    "LogLvEmergency",
			Keywords: utils.JoinSlices(lowerCapSCREAM("emerg"), lowerCapSCREAM("emergency")),
		},
		{
			Group:    "LogLvAlert",
			Keywords: lowerCapSCREAM("alert"),
		},
		{
			Group:    "LogLvCritical",
			Keywords: utils.JoinSlices(lowerCapSCREAM("crit"), lowerCapSCREAM("critical")),
		},
		{
			Group: "LogLvError",
			Keywords: utils.JoinSlices(
				[]string{"E"},
				lowerCapSCREAM("err"),
				lowerCapSCREAM("error"),
				lowerCapSCREAM("errors"),
			),
		},
		{
			Group: "LogLvFail",
			Keywords: utils.JoinSlices(
				[]string{"F"},
				lowerCapSCREAM("fail"),
				lowerCapSCREAM("failed"),
				lowerCapSCREAM("failure"),
			),
		},
		{
			Group:    "LogLvFault",
			Keywords: lowerCapSCREAM("fault"),
		},
		{
			Group:    "LogLvNack",
			Keywords: utils.JoinSlices(lowerCapSCREAM("nack"), lowerCapSCREAM("nak")),
		},
		{
			Group:    "LogLvWarning",
			Keywords: utils.JoinSlices([]string{"W"}, lowerCapSCREAM("warn"), lowerCapSCREAM("warning")),
		},
		{
			Group:    "LogLvBad",
			Keywords: lowerCapSCREAM("bad"),
		},
		{
			Group:    "LogLvNotice",
			Keywords: lowerCapSCREAM("notice"),
		},
		{
			Group:    "LogLvInfo",
			Keywords: utils.JoinSlices([]string{"I"}, lowerCapSCREAM("info")),
		},
		{
			Group:    "LogLvDebug",
			Keywords: utils.JoinSlices([]string{"D"}, lowerCapSCREAM("dbg"), lowerCapSCREAM("debug")),
		},
		{
			Group:    "LogLvTrace",
			Keywords: lowerCapSCREAM("trace"),
		},
		{
			Group:    "LogLvVerbose",
			Keywords: utils.JoinSlices([]string{"V"}, lowerCapSCREAM("verbose")),
		},
		{
			Group:    "LogLvPass",
			Keywords: utils.JoinSlices(lowerCapSCREAM("pass"), lowerCapSCREAM("passed")),
		},
		{
			Group:    "LogLvSuccess",
			Keywords: utils.JoinSlices(lowerCapSCREAM("succeed"), lowerCapSCREAM("succeeded"), lowerCapSCREAM("success")),
		},

		// Composite log levels e.g. *_INFO
		{
			Group:   "LogLvFatal",
			Pattern: proto.MustCompile(`[A-Z_]+_FATAL\b`),
		},
		{
			Group:   "LogLvEmergency",
			Pattern: proto.MustCompile(`[A-Z_]+_EMERG(ENCY)?\b`),
		},
		{
			Group:   "LogLvAlert",
			Pattern: proto.MustCompile(`[A-Z_]+_ALERT\b`),
		},
		{
			Group:   "LogLvCritical",
			Pattern: proto.MustCompile(`[A-Z_]+_CRIT(ICAL)?\b`),
		},
		{
			Group:   "LogLvError",
			Pattern: proto.MustCompile(`[A-Z_]+_ERR(OR)?\b`),
		},
		{
			Group:   "LogLvFail",
			Pattern: proto.MustCompile(`[A-Z_]+_FAIL(URE)?\b`),
		},
		{
			Group:   "LogLvWarning",
			Pattern: proto.MustCompile(`[A-Z_]+_WARN(ING)?\b`),
		},
		{
			Group:   "LogLvNotice",
			Pattern: proto.MustCompile(`[A-Z_]+_NOTICE\b`),
		},
		{
			Group:   "LogLvInfo",
			Pattern: proto.MustCompile(`[A-Z_]+_INFO\b`),
		},
		{
			Group:   "LogLvDebug",
			Pattern: proto.MustCompile(`[A-Z_]+_DEBUG\b`),
		},
		{
			Group:   "LogLvTrace",
			Pattern: proto.MustCompile(`[A-Z_]+_TRACE\b`),
		},
	},
	Highlight: []highlight{
		{Group: "LogNumber", Link: strPtr("Number")},
		{Group: "LogNumberFloat", Link: strPtr("Float")},
		{Group: "LogNumberBin", Link: strPtr("Number")},
		{Group: "LogNumberOctal", Link: strPtr("Number")},
		{Group: "LogNumberHex", Link: strPtr("Number")},
		{Group: "LogSymbol", Link: strPtr("Special")},
		{Group: "LogSeparatorLine", Link: strPtr("Comment")},
		{Group: "LogBool", Link: strPtr("Boolean")},
		{Group: "LogNull", Link: strPtr("Constant")},
		{Group: "LogString", Link: strPtr("String")},
		{Group: "LogDate", Link: strPtr("Type")},
		{Group: "LogWeekdayStr", Link: strPtr("Type")},
		{Group: "LogTime", Link: strPtr("Operator")},
		{Group: "LogTimeAMPM", Link: strPtr("Operator")},
		{Group: "LogTimeZone", Link: strPtr("Operator")},
		{Group: "LogDuration", Link: strPtr("Operator")},
		{Group: "LogSysColumns", Link: strPtr("Statement")},
		{Group: "LogSysProcess", Link: strPtr("Function")},
		{Group: "LogUrl", Link: strPtr("Underlined")},
		{Group: "LogMacAddr", Link: strPtr("Underlined")},
		{Group: "LogIPv4", Link: strPtr("Underlined")},
		{Group: "LogIPv6", Link: strPtr("Underlined")},
		{Group: "LogUUID", Link: strPtr("Label")},
		{Group: "LogMD5", Link: strPtr("Label")},
		{Group: "LogSHA", Link: strPtr("Label")},
		{Group: "LogPath", Link: strPtr("Structure")},
		{Group: "LogLvFatal", Link: strPtr("ErrorMsg")},
		{Group: "LogLvEmergency", Link: strPtr("ErrorMsg")},
		{Group: "LogLvAlert", Link: strPtr("ErrorMsg")},
		{Group: "LogLvCritical", Link: strPtr("ErrorMsg")},
		{Group: "LogLvError", Link: strPtr("ErrorMsg")},
		{Group: "LogLvFail", Link: strPtr("ErrorMsg")},
		{Group: "LogLvFault", Link: strPtr("ErrorMsg")},
		{Group: "LogLvNack", Link: strPtr("ErrorMsg")},
		{Group: "LogLvWarning", Link: strPtr("WarningMsg")},
		{Group: "LogLvBad", Link: strPtr("WarningMsg")},
		{Group: "LogLvNotice", Link: strPtr("Exception")},
		{Group: "LogLvInfo", Link: strPtr("LogBlue")},
		{Group: "LogLvDebug", Link: strPtr("Debug")},
		{Group: "LogLvTrace", Link: strPtr("Special")},
		{Group: "LogLvVerbose", Link: strPtr("Special")},
		{Group: "LogLvPass", Link: strPtr("LogGreen")},
		{Group: "LogLvSuccess", Link: strPtr("LogGreen")},
	},
}

func GetDefaultConfig() Config {
	return DefaultConfig
}
