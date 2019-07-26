package redactif

import (
	"reflect"
	"strings"
	"unsafe"
)

type redactor struct {
	tags          map[string]struct{}
	seenAddresses map[uintptr]struct{}
}

// Redact iterates over a datastructure and zeroes any struct fields with the "redactif"
// struct tag, matching the following rules:
//
// Struct Tag is like
//   redactif:"user"
// and tags passed to Redact includes "example"
//
// Struct Tag is like
//   redactif:"!admin"
// and tags passed to Redact does not include "admin"
func Redact(i interface{}, tags ...string) interface{} {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		panic("cannot redact non-pointer value")
	}

	tagMap := map[string]struct{}{}
	for _, tag := range tags {
		tagMap[tag] = struct{}{}
	}

	r := &redactor{
		tags:          tagMap,
		seenAddresses: map[uintptr]struct{}{},
	}
	r.redactValue(reflect.ValueOf(i))
	return i
}

func (r *redactor) redactValue(v reflect.Value) {
	if !v.IsValid() {
		// zero value => probably result of nil pointer
		return
	}

	// Avoid infinite loops caused by data structure cycles.
	// This assumes that you cannot make a cycle consisting of entirely unaddressable values
	if v.CanAddr() {
		addr := v.UnsafeAddr()
		if _, ok := r.seenAddresses[addr]; ok {
			// already redacted the value at this address
			return
		}
		r.seenAddresses[addr] = struct{}{}
	}

	switch v.Kind() {
	// Indirections
	case reflect.Ptr, reflect.Interface:
		r.redactValue(v.Elem())

	// Collections
	case reflect.Struct:
		// The only actually interesting case :)
		r.redactStruct(v)

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			r.redactValue(v.Index(i))
		}
	case reflect.Map:
		// TODO

	default:
		// type not interesting
	}
}

func (r *redactor) redactStruct(v reflect.Value) {
	for index := 0; index < v.Type().NumField(); index++ {
		field := v.Field(index)
		if !field.CanAddr() {
			// TODO: when does this happen? Can we work around it?
			continue
		}
		fieldTag, ok := v.Type().Field(index).Tag.Lookup("redactif")
		if ok {
			fieldTags := strings.Split(fieldTag, ",")
			for _, tag := range fieldTags {
				_, haveTag := r.tags[strings.TrimPrefix(tag, "!")]
				notConstraint := strings.HasPrefix(tag, "!")
				if (notConstraint && !haveTag) ||
					(!notConstraint && haveTag) {
					// Should redact this field
					field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
					field.Set(reflect.Zero(field.Type()))
				}
			}
		}

		// recurse
		r.redactValue(field)
	}
}

//func (m *mapper) mapMap(mapVal reflect.Value, parentID nodeID, inlineable bool) (nodeID, string) {
//	// create a string type while escaping graphviz special characters
//	mapType := escapeString(mapVal.Type().String())
//
//	nodeKey := getNodeKey(mapVal)
//
//	if mapVal.Len() == 0 {
//		m.nodeSummaries[nodeKey] = mapType + "\\{\\}"
//
//		if inlineable {
//			return 0, m.nodeSummaries[nodeKey]
//		}
//
//		return m.newBasicNode(mapVal, m.nodeSummaries[nodeKey]), mapType
//	}
//
//	mapID := m.getNodeID(mapVal)
//	var id nodeID
//	if inlineable && mapVal.Len() <= m.inlineableItemLimit {
//		m.nodeSummaries[nodeKey] = mapType
//		id = parentID
//	} else {
//		id = mapID
//	}
//
//	var links []string
//	var fields string
//	for index, mapKey := range mapVal.MapKeys() {
//		keyID, keySummary := m.mapValue(mapKey, id, true)
//		valueID, valueSummary := m.mapValue(mapVal.MapIndex(mapKey), id, true)
//		fields += fmt.Sprintf("|{<%dkey%d> %s| <%dvalue%d> %s}", mapID, index, keySummary, mapID, index, valueSummary)
//		if keyID != 0 {
//			links = append(links, fmt.Sprintf("  %d:<%dkey%d> -> %d:name;\n", id, mapID, index, keyID))
//		}
//		if valueID != 0 {
//			links = append(links, fmt.Sprintf("  %d:<%dvalue%d> -> %d:name;\n", id, mapID, index, valueID))
//		}
//	}
//
//	for _, link := range links {
//		fmt.Fprint(m.writer, link)
//	}
//
//	if inlineable && mapVal.Len() <= m.inlineableItemLimit {
//		// inline map
//		// remove stored summary so this gets regenerated every time
//		// we need to do this so that we get a chance to print out the new links
//		delete(m.nodeSummaries, nodeKey)
//
//		// have to remove invalid leading |
//		return 0, "{" + fields[1:] + "}"
//	}
//
//	// else create a new node
//	node := fmt.Sprintf("  %d [label=\"<name> %s %s \"];\n", id, mapType, fields)
//	fmt.Fprint(m.writer, node)
//
//	return id, m.nodeSummaries[nodeKey]
//}
