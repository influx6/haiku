package views

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/influx6/haiku/jsutils"
)

// CreateFragment returns a DocumentFragment with the given html dom
func CreateFragment(html string) *js.Object {
	//if we are not in a browser,panic
	panicBrowserDetect()

	//we need to use innerhtml but DocumentFragments dont have that so we use
	//a discardable div

	// div := doc.CreateElement("div")
	div := jsutils.CreateElement("div")

	//build up the html right in the div
	jsutils.SetInnerHTML(div, html)

	//unwrap all the special Text UnknownELement we are using
	// jsutils.UnWrapSpecialTextElements(div)

	//create the document fragment
	fragment := jsutils.CreateDocumentFragment()

	//add the nodes from the div into the fragment
	nodes := jsutils.ChildNodeList(div)

	jsutils.AppendChild(fragment, nodes...)

	//unwrap all the special Text UnknownELement we are using
	// jsutils.UnWrapSpecialTextElements(fragment)

	return fragment
}

// AddNodeIfNone uses the dom.Node.IsEqualNode method to check if not already exist and if so swap them else just add
// NOTE: bad way of doing it use it as last option
func AddNodeIfNone(dest, src *js.Object) {
	AddNodeIfNoneInList(dest, jsutils.ChildNodeList(dest), src)
}

// AddNodeIfNoneInList checks a node in a node list if it finds an equal it replaces only else does nothing
func AddNodeIfNoneInList(dest *js.Object, against []*js.Object, with *js.Object) bool {
	for _, no := range against {
		if jsutils.IsEqualNode(no, with) {
			// if no.IsEqualNode(with) {
			jsutils.ReplaceNode(dest, with, no)
			return true
		}
	}
	//not matching, add it
	jsutils.AppendChild(dest, with)
	// dest.AppendChild(with)
	return false
}

// Patch takes a dom string and creates a documentfragment from it and patches a existing dom element that is supplied. This algorithim only ever goes one-level deep, its not performant
// WARNING: this method is specifically geared to dealing with the haiku.Tree dom generation
func Patch(fragment, live *js.Object, onlyReplace bool) {
	if !live.Call("hasChildNodes").Bool() {
		// if the live element is actually empty, then just append the fragment which
		// actually appends the nodes within it efficiently

		jsutils.AppendChild(live, fragment)
		return
	}

	// log.Printf("doing patching")
	// shadowNodes := fragment.ChildNodes()

	shadowNodes := jsutils.ChildNodeList(fragment)
	liveNodes := jsutils.ChildNodeList(live)

	// FIXED: instead of going through the children which may be many,
	// liveNodes := fragment.ChildNodes()
	// log.Printf("patchtree will now add: \n%+s", shadowNodes)

patchloop:
	for n, node := range shadowNodes {
		if node == nil || node == js.Undefined {
			continue
		}
		// switch node.Get("constructor") {
		// case js.Global.Get("Node"):
		// elem := node

		if node.Get("constructor") == js.Global.Get("Text") {
			// log.Printf("text %+s %s %s %d", node, node.Get("nodeName"), node.Get("innerText"), node.Get("nodeType").Int())

			if _, empty := jsutils.EmptyTextNode(node); empty {
				jsutils.AppendChild(live, node)
				continue patchloop
			}

			var liveNodeAt *js.Object

			if n < len(liveNodes) {
				liveNodeAt = liveNodes[n]
			}

			if liveNodeAt == nil || liveNodeAt == js.Undefined {
				jsutils.AppendChild(live, node)
			} else {
				jsutils.InsertBefore(live, liveNodeAt, node)
			}

			continue patchloop
		}

		//get the tagname
		tagname := jsutils.GetTag(node)
		// log.Printf("Working with tag %s -> %+s", tagname, nchildren)

		// get the basic attrs
		var id, hash, class, uid string

		// do we have 'id' attribute? if so its a awesome chance to simplify
		if jsutils.HasAttribute(node, "id") {
			id = jsutils.GetAttribute(node, "id")
		}

		if jsutils.HasAttribute(node, "class") {
			id = jsutils.GetAttribute(node, "class")
		}

		// lets check for the hash and uid, incase its a pure template based script
		if jsutils.HasAttribute(node, "hash") {
			hash = jsutils.GetAttribute(node, "hash")
		}

		if jsutils.HasAttribute(node, "uid") {
			uid = jsutils.GetAttribute(node, "uid")
		}

		// if tagname == "tmlview" {
		// 	frstchild := node.()
		// 	frstchild.
		// }

		// if we have no id,class, uid or hash, we digress to bad approach of using Node.IsEqualNode
		if allEmpty(id, hash, uid) {
			// log.Printf("adding since hash,id,uid are empty")
			AddNodeIfNone(live, node)
			continue patchloop
		}

		// eliminate which ones are empty and try to use the non empty to get our target
		if allEmpty(hash, uid) {
			// is the id empty also then we know class is not or vise-versa
			if allEmpty(id) {
				// log.Printf("adding since class")
				// class is it and we only want those that match narrowing our set
				no := jsutils.QuerySelectorAll(live, class)

				// if none found we add else we replace
				if len(no) <= 0 {
					jsutils.AppendChild(live, node)
				} else {
					// check the available sets and replace else just add it
					AddNodeIfNoneInList(live, no, node)
				}

			} else {
				// id is it and we only want one
				// log.Printf("adding since id")
				no := jsutils.QuerySelector(live, fmt.Sprintf("#%s", id))

				// if none found we add else we replace
				if no == nil || no != js.Undefined {
					jsutils.AppendChild(live, node)
				} else {
					jsutils.ReplaceNode(live, node, no)
				}
			}

			continue patchloop
		}

		// lets use our unique id to check for the element if it exists
		sel := fmt.Sprintf(`%s[uid='%s']`, strings.ToLower(tagname), uid)

		// we know hash and uid are not empty so we kick ass the easy way
		target := jsutils.QuerySelector(live, sel)

		// log.Printf("textElem %s -> %s -> %s : target -> %+s", node, node.Get("tagName"), sel, target)

		// if we are nil then its a new node add it and return
		if target == nil || target == js.Undefined {
			jsutils.AppendChild(live, node)
			continue patchloop
		}

		if onlyReplace {
			jsutils.ReplaceNode(live, node, target)
			continue patchloop
		}

		//if we are to be removed then remove the target
		if jsutils.HasAttribute(node, "haikuRemoved") {
			// log.Printf("removed node: %+s", node)
			// target.ParentNode().RemoveChild(target)
			jsutils.RemoveChild(target, target)
			continue patchloop
		}

		// if the target hash is exactly the same with ours skip it
		if jsutils.GetAttribute(target, "hash") == hash {
			continue patchloop
		}

		nchildren := jsutils.ChildNodeList(node)
		// log.Printf("Checking size of children %s %d", sel, len(nchildren))
		//if the new node has no children, then just replace it
		// if len(elem.ChildNodes()) <= 0 {
		if len(nchildren) <= 0 {
			// live.ReplaceChild(node, target)
			jsutils.ReplaceNode(live, node, target)
			continue patchloop
		}

		//here we are not be removed and we do have kids

		//cleanout all the targets text-nodes
		jsutils.CleanAllTextNode(target)

		//so we got this dude, are we already one level deep ? if so swap else
		// run through the children with Patch
		// if level >= 1 {
		// live.ReplaceChild(node, target)
		attrs := jsutils.Attributes(node)

		for key, value := range attrs {
			jsutils.SetAttribute(target, key, value)
		}

		children := jsutils.ChildNodeList(target)

		// log.Printf("checking targets children %+s %d", target, len(children))
		if len(children) <= 1 {
			jsutils.SetInnerHTML(target, "")

			jsutils.AppendChild(target, nchildren...)

			// for _, enode := range nchildren {
			// 	target.AppendChild(enode)
			// }

			continue patchloop
		}

		Patch(node, target, onlyReplace)
		// }

		// default:
		// 	// add it if its not an element
		// 	jsutils.AppendChild(live, node)
		// live.AppendChild(node)
		// }
	}
}

// allEmpty checks if all strings supplied are empty
func allEmpty(s ...string) bool {
	var state = true

	for _, so := range s {
		if so != "" {
			state = false
			return state
		}
	}

	return state
}
