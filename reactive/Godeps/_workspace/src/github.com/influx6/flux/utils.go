package flux

import (
	"regexp"
	"strconv"
	"time"
)

var elapso = regexp.MustCompile(`(\d+)(\w+)`)

//MakeDuration allows you to make create a duration from a string
func MakeDuration(target string, def int) time.Duration {
	if !elapso.MatchString(target) {
		return time.Duration(def)
	}

	matchs := elapso.FindAllStringSubmatch(target, -1)

	if len(matchs) <= 0 {
		return time.Duration(def)
	}

	match := matchs[0]

	if len(match) < 3 {
		return time.Duration(def)
	}

	dur := time.Duration(ConvertToInt(match[1], def))

	mtype := match[2]

	switch mtype {
	case "s":
		log.Info("Setting %d in seconds", dur)
		return dur * time.Second
	case "mcs":
		log.Info("Setting %d in Microseconds", dur)
		return dur * time.Microsecond
	case "ns":
		log.Info("Setting %d in Nanoseconds", dur)
		return dur * time.Nanosecond
	case "ms":
		log.Info("Setting %d in Milliseconds", dur)
		return dur * time.Millisecond
	case "m":
		log.Info("Setting %d in Minutes", dur)
		return dur * time.Minute
	case "h":
		log.Info("Setting %d in Hours", dur)
		return dur * time.Hour
	default:
		log.Info("Defaul %d to Seconds", dur)
		return time.Duration(dur) * time.Second
	}
}

//ConvertToInt wraps the internal int coverter
func ConvertToInt(target string, def int) int {
	fo, err := strconv.Atoi(target)
	if err != nil {
		return def
	}
	return fo
}
