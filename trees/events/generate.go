// +build ignore

// The generation of this package was inspired by Neelance work on DOM (https://github.com/neelance/dom)

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type event struct {
	Name string
	Link string
	Desc string
}

func main() {
	nameMap := map[string]string{
		"afterprint":              "AfterPrint",
		"animationend":            "AnimationEnd",
		"animationiteration":      "AnimationIteration",
		"animationstart":          "AnimationStart",
		"audioprocess":            "AudioProcess",
		"beforeprint":             "BeforePrint",
		"beforeunload":            "BeforeUnload",
		"canplay":                 "CanPlay",
		"canplaythrough":          "CanPlayThrough",
		"chargingchange":          "ChargingChange",
		"chargingtimechange":      "ChargingTimeChange",
		"compassneedscalibration": "CompassNeedsCalibration",
		"compositionend":          "CompositionEnd",
		"compositionstart":        "CompositionStart",
		"compositionupdate":       "CompositionUpdate",
		"contextmenu":             "ContextMenu",
		"dblclick":                "DblClick",
		"devicelight":             "DeviceLight",
		"devicemotion":            "DeviceMotion",
		"deviceorientation":       "DeviceOrientation",
		"deviceproximity":         "DeviceProximity",
		"dischargingtimechange":   "DischargingTimeChange",
		"dragend":                 "DragEnd",
		"dragenter":               "DragEnter",
		"dragleave":               "DragLeave",
		"dragover":                "DragOver",
		"dragstart":               "DragStart",
		"durationchange":          "DurationChange",
		"focusin":                 "FocusIn",
		"focusout":                "FocusOut",
		"fullscreenchange":        "FullScreenChange",
		"fullscreenerror":         "FullScreenError",
		"gamepadconnected":        "GamepadConnected",
		"gamepaddisconnected":     "GamepadDisconnected",
		"hashchange":              "HashChange",
		"keydown":                 "KeyDown",
		"keypress":                "KeyPress",
		"keyup":                   "KeyUp",
		"languagechange":          "LanguageChange",
		"levelchange":             "LevelChange",
		"loadeddata":              "LoadedData",
		"loadedmetadata":          "LoadedMetadata",
		"loadend":                 "LoadEnd",
		"loadstart":               "LoadStart",
		"mousedown":               "MouseDown",
		"mouseenter":              "MouseEnter",
		"mouseleave":              "MouseLeave",
		"mousemove":               "MouseMove",
		"mouseout":                "MouseOut",
		"mouseover":               "MouseOver",
		"mouseup":                 "MouseUp",
		"noupdate":                "NoUpdate",
		"orientationchange":       "OrientationChange",
		"pagehide":                "PageHide",
		"pageshow":                "PageShow",
		"pointerlockchange":       "PointerLockChange",
		"pointerlockerror":        "PointerLockError",
		"popstate":                "PopState",
		"ratechange":              "RateChange",
		"readystatechange":        "ReadystateChange",
		"timeupdate":              "TimeUpdate",
		"touchcancel":             "TouchCancel",
		"touchend":                "TouchEnd",
		"touchenter":              "TouchEnter",
		"touchleave":              "TouchLeave",
		"touchmove":               "TouchMove",
		"touchstart":              "TouchStart",
		"transitionend":           "TransitionEnd",
		"updateready":             "UpdateReady",
		"upgradeneeded":           "UpgradeNeeded",
		"userproximity":           "UserProximity",
		"versionchange":           "VersionChange",
		"visibilitychange":        "VisibilityChange",
		"volumechange":            "VolumeChange",
	}

	doc, err := goquery.NewDocument("https://developer.mozilla.org/en-US/docs/Web/Events")
	if err != nil {
		panic(err)
	}

	events := make(map[string]*event)

	doc.Find(".standard-table").Eq(0).Find("tr").Each(func(i int, s *goquery.Selection) {
		cols := s.Find("td")
		if cols.Length() == 0 || cols.Find(".icon-thumbs-down-alt").Length() != 0 {
			return
		}
		link := cols.Eq(0).Find("a").Eq(0)
		var e event
		e.Name = link.Text()
		e.Link, _ = link.Attr("href")
		e.Desc = strings.TrimSpace(cols.Eq(3).Text())
		if e.Desc == "" {
			e.Desc = "(no documentation)"
		}

		funName := nameMap[e.Name]
		if funName == "" {
			funName = capitalize(e.Name)
		}

		events[funName] = &e
	})

	var names []string
	for name := range events {
		names = append(names, name)
	}
	sort.Strings(names)

	file, err := os.Create("event.gen.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprint(file, `// The generation of this package was inspired by Neelance work on DOM (https://github.com/neelance/dom)

//go:generate go run generate.go

// Documentation source: "Event reference" by Mozilla Contributors, https://developer.mozilla.org/en-US/docs/Web/Events, licensed under CC-BY-SA 2.5.

//Package events defines the event binding system for Haiku(https://github.com/influx6/Haiku)
package events

import (
	"github.com/influx6/haiku/trees"
)
`)

	for _, name := range names {
		e := events[name]
		fmt.Fprintf(file, `
// %s Documentation is as below:
// %s
// https://developer.mozilla.org%s
/* This event provides options() to be called when the events is triggered and an optional selector which will override the internal selector mechanism of the trees.Element i.e if the selectorOverride argument is an empty string then trees.Element will create an appropriate selector matching its type and uid value in this format  (ElementType[uid='UID_VALUE']) but if the selector value is not empty then that becomes the default selector used
match the event with. */
func %s(fx trees.EventHandler,selectorOverride string) *trees.Event {
	return trees.NewEvent("%s",selectorOverride,fx)
}
`, name, e.Desc, e.Link[6:], name, e.Name)
	}
}

func capitalize(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}
