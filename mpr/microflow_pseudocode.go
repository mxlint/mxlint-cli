package mpr

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type microflowFlow struct {
	ID          string
	Origin      string
	Destination string
	CaseValues  []string
}

func generateMicroflowPseudocode(microflowName string, attributes map[string]interface{}) (string, error) {
	graph, err := buildMicroflowGraph(attributes)
	if err != nil {
		return "", err
	}

	orderedNodeIDs := traverseMicroflowNodes(graph.startID, graph.outgoing)
	if len(orderedNodeIDs) == 0 {
		return "", fmt.Errorf("no traversable nodes in microflow %s", microflowName)
	}

	labelByID, needsLabel := computeLabels(graph, orderedNodeIDs)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("MICROFLOW: %s\n", microflowName))
	b.WriteString(fmt.Sprintf("RETURN TYPE: %s\n", extractReturnType(attributes)))
	params := extractMicroflowParameters(graph.objectsByID)
	if len(params) > 0 {
		b.WriteString("PARAMETERS:\n")
		for _, p := range params {
			b.WriteString(fmt.Sprintf("  - %s\n", p))
		}
	}
	b.WriteString("\nPSEUDOCODE\n----------\n")
	b.WriteString("BEGIN\n")

	for idx, nodeID := range orderedNodeIDs {
		node := graph.objectsByID[nodeID]
		nodeType, _ := node["$Type"].(string)
		out := graph.outgoing[nodeID]
		nextID := ""
		if idx+1 < len(orderedNodeIDs) {
			nextID = orderedNodeIDs[idx+1]
		}

		if needsLabel[nodeID] {
			b.WriteString(fmt.Sprintf("  LABEL %s\n", labelByID[nodeID]))
		}
		for _, line := range renderNodeInstruction(nodeID, node, nodeType, graph) {
			b.WriteString("  " + line + "\n")
		}
		for _, line := range renderControlFlow(nodeType, node, out, labelByID, nextID) {
			b.WriteString("  " + line + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("END")
	return b.String(), nil
}

func renderNodeInstruction(nodeID string, node map[string]interface{}, nodeType string, graph microflowGraph) []string {
	switch nodeType {
	case "Microflows$ActionActivity":
		actionRaw, ok := node["Action"].(map[string]interface{})
		if !ok {
			return []string{"// ActionActivity (unparsed action)"}
		}
		actionType, _ := actionRaw["$Type"].(string)
		switch actionType {
		case "Microflows$LogMessageAction":
			msg := extractNestedString(actionRaw, "MessageTemplate", "Text")
			return []string{fmt.Sprintf("log info: %q", msg)}
		case "Microflows$CreateVariableAction":
			name, _ := actionRaw["VariableName"].(string)
			initial, _ := actionRaw["InitialValue"].(string)
			return []string{fmt.Sprintf("%s = %s", name, initial)}
		case "Microflows$ChangeVariableAction":
			name, _ := actionRaw["ChangeVariableName"].(string)
			value, _ := actionRaw["Value"].(string)
			return []string{fmt.Sprintf("%s = %s", name, value)}
		case "Microflows$RetrieveAction":
			result, _ := actionRaw["ResultVariableName"].(string)
			entity := extractNestedString(actionRaw, "RetrieveSource", "Entity")
			if result == "" {
				result = "<result>"
			}
			if entity == "" {
				entity = "<entity>"
			}
			return []string{fmt.Sprintf("%s = retrieve from database %s", result, entity)}
		case "Microflows$ChangeAction":
			return renderChangeAction(actionRaw)
		case "Microflows$MicroflowCallAction":
			return []string{renderMicroflowCallAction(actionRaw)}
		default:
			return []string{fmt.Sprintf("// action: %s", actionType)}
		}
	case "Microflows$ExclusiveSplit":
		return nil
	case "Microflows$ExclusiveMerge":
		return nil
	case "Microflows$LoopedActivity":
		return renderLoopedActivity(nodeID, node, graph)
	case "Microflows$StartEvent":
		return nil
	case "Microflows$EndEvent":
		return nil
	default:
		return []string{fmt.Sprintf("// node: %s", nodeType)}
	}
}

func renderChangeAction(actionRaw map[string]interface{}) []string {
	variableName, _ := actionRaw["ChangeVariableName"].(string)
	if variableName == "" {
		variableName = "<variable>"
	}
	lines := make([]string, 0)

	for _, item := range asObjectSlice(actionRaw["Items"]) {
		itemType, _ := item["$Type"].(string)
		if itemType != "Microflows$ChangeActionItem" {
			continue
		}
		attribute, _ := item["Attribute"].(string)
		value, _ := item["Value"].(string)
		if value == "" {
			value = "<value>"
		}
		fieldName := attribute
		if idx := strings.LastIndex(attribute, "."); idx >= 0 && idx+1 < len(attribute) {
			fieldName = attribute[idx+1:]
		}
		if fieldName == "" {
			fieldName = "<field>"
		}
		lines = append(lines, fmt.Sprintf("%s.%s = %s", variableName, fieldName, value))
	}
	if commit, _ := actionRaw["Commit"].(string); strings.EqualFold(commit, "yes") {
		lines = append(lines, fmt.Sprintf("commit %s", variableName))
	}
	if len(lines) == 0 {
		lines = append(lines, "// change object")
	}
	return lines
}

func renderMicroflowCallAction(actionRaw map[string]interface{}) string {
	call, ok := actionRaw["MicroflowCall"].(map[string]interface{})
	if !ok {
		return "call <microflow>"
	}
	microflowName, _ := call["Microflow"].(string)
	if microflowName == "" {
		microflowName = "<microflow>"
	}
	args := make([]string, 0)
	for _, mapping := range asObjectSlice(call["ParameterMappings"]) {
		param, _ := mapping["Parameter"].(string)
		arg, _ := mapping["Argument"].(string)
		if param == "" || arg == "" {
			continue
		}
		shortParam := param
		if idx := strings.LastIndex(param, "."); idx >= 0 && idx+1 < len(param) {
			shortParam = param[idx+1:]
		}
		args = append(args, fmt.Sprintf("%s = %s", shortParam, arg))
	}
	if len(args) == 0 {
		return fmt.Sprintf("call %s()", microflowName)
	}
	return fmt.Sprintf("call %s(%s)", microflowName, strings.Join(args, ", "))
}

func renderLoopedActivity(nodeID string, node map[string]interface{}, graph microflowGraph) []string {
	loopSource, _ := node["LoopSource"].(map[string]interface{})
	listName, _ := loopSource["ListVariableName"].(string)
	iteratorName, _ := loopSource["VariableName"].(string)
	if listName == "" {
		listName = "<list>"
	}
	if iteratorName == "" {
		iteratorName = "<item>"
	}

	innerObjects := getLoopObjectCollectionObjects(node)
	if len(innerObjects) == 0 {
		return []string{fmt.Sprintf("FOR EACH %s IN %s", iteratorName, listName), "END FOR"}
	}

	innerByID := make(map[string]map[string]interface{}, len(innerObjects))
	for _, obj := range innerObjects {
		id := readMicroflowID(obj["$ID"])
		if id == "" {
			continue
		}
		innerByID[id] = obj
	}
	if len(innerByID) == 0 {
		return []string{fmt.Sprintf("FOR EACH %s IN %s", iteratorName, listName), "END FOR"}
	}

	splitID, splitObj := findFirstExclusiveSplit(innerByID)
	if splitID != "" {
		branchFlows := make([]microflowFlow, 0, 2)
		for _, f := range graph.outgoing[splitID] {
			if _, ok := innerByID[f.Destination]; ok {
				branchFlows = append(branchFlows, f)
			}
		}
		if len(branchFlows) == 2 {
			var trueDest, falseDest string
			for _, f := range branchFlows {
				switch firstCaseValue(f.CaseValues) {
				case "true":
					trueDest = f.Destination
				case "false":
					falseDest = f.Destination
				}
			}
			if trueDest != "" && falseDest != "" {
				expr := extractNestedString(splitObj, "SplitCondition", "Expression")
				if expr == "" {
					expr = "<condition>"
				}
				trueLines := collectLinearBranchLines(trueDest, innerByID, graph.outgoing, graph)
				falseLines := collectLinearBranchLines(falseDest, innerByID, graph.outgoing, graph)
				lines := []string{fmt.Sprintf("FOR EACH %s IN %s", iteratorName, listName)}
				lines = append(lines, fmt.Sprintf("  IF %s THEN", expr))
				for _, l := range trueLines {
					lines = append(lines, "    "+l)
				}
				lines = append(lines, "  ELSE")
				for _, l := range falseLines {
					lines = append(lines, "    "+l)
				}
				lines = append(lines, "  END IF")
				lines = append(lines, "END FOR")
				return lines
			}
		}
	}

	lines := []string{fmt.Sprintf("FOR EACH %s IN %s", iteratorName, listName)}
	entries := getInnerEntryNodes(innerByID, graph.outgoing)
	for _, entry := range entries {
		for _, l := range collectLinearBranchLines(entry, innerByID, graph.outgoing, graph) {
			lines = append(lines, "  "+l)
		}
	}
	lines = append(lines, "END FOR")
	return lines
}

func traverseLoopBody(entryIDs []string, outgoing map[string][]microflowFlow, scope map[string]map[string]interface{}) []string {
	visited := make(map[string]bool, len(scope))
	order := make([]string, 0, len(scope))
	var visit func(string)
	visit = func(id string) {
		if id == "" || visited[id] {
			return
		}
		if _, ok := scope[id]; !ok {
			return
		}
		visited[id] = true
		order = append(order, id)
		for _, edge := range outgoing[id] {
			if _, ok := scope[edge.Destination]; ok {
				visit(edge.Destination)
			}
		}
	}

	for _, entry := range entryIDs {
		visit(entry)
	}
	for id := range scope {
		visit(id)
	}
	return order
}

func getLoopObjectCollectionObjects(node map[string]interface{}) []map[string]interface{} {
	collection, ok := node["ObjectCollection"].(map[string]interface{})
	if !ok {
		return nil
	}
	return asObjectSlice(collection["Objects"])
}

func asObjectSlice(raw interface{}) []map[string]interface{} {
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		result = append(result, m)
	}
	return result
}

func firstIDSegment(id string) string {
	if id == "" {
		return "X"
	}
	parts := strings.Split(id, "-")
	if len(parts) == 0 || parts[0] == "" {
		return "X"
	}
	raw := parts[0]
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	token := strings.Trim(b.String(), "_")
	if token == "" {
		return "X"
	}
	if len(token) > 12 {
		token = token[:12]
	}
	return token
}

func renderControlFlow(nodeType string, node map[string]interface{}, out []microflowFlow, labelByID map[string]string, nextID string) []string {
	if nodeType == "Microflows$EndEvent" || nodeType == "Microflows$LoopedActivity" {
		return nil
	}
	if len(out) == 0 {
		return nil
	}

	switch nodeType {
	case "Microflows$ExclusiveSplit":
		expr := extractNestedString(node, "SplitCondition", "Expression")
		if expr == "" {
			expr = "<condition>"
		}

		// Fast path for two-way boolean split.
		if len(out) == 2 {
			var trueFlow *microflowFlow
			var falseFlow *microflowFlow
			for i := range out {
				switch firstCaseValue(out[i].CaseValues) {
				case "true":
					trueFlow = &out[i]
				case "false":
					falseFlow = &out[i]
				}
			}

			if trueFlow != nil && falseFlow != nil {
				trueLabel := labelByID[trueFlow.Destination]
				falseLabel := labelByID[falseFlow.Destination]
				if trueLabel == "" {
					trueLabel = "<next>"
				}
				if falseLabel == "" {
					falseLabel = "<next>"
				}
				return []string{
					fmt.Sprintf("IF %s THEN", expr),
					fmt.Sprintf("  GOTO %s", trueLabel),
					"ELSE",
					fmt.Sprintf("  GOTO %s", falseLabel),
					"END IF",
				}
			}
		}

		// Generic split fallback, preserving all outgoing branches.
		lines := []string{fmt.Sprintf("IF %s THEN", expr)}
		for idx, flow := range out {
			label := labelByID[flow.Destination]
			if label == "" {
				label = "<next>"
			}
			caseLabel := strings.Join(flow.CaseValues, ", ")
			if caseLabel == "" {
				caseLabel = "default"
			}
			prefix := "  GOTO"
			if idx > 0 {
				prefix = "  // OR GOTO"
			}
			lines = append(lines, fmt.Sprintf("%s %s  // case: %s", prefix, label, caseLabel))
		}
		lines = append(lines, "END IF")
		return lines
	default:
		dest := out[0].Destination
		if dest == nextID {
			return nil
		}
		return []string{fmt.Sprintf("GOTO %s", labelByID[dest])}
	}
}

func computeLabels(graph microflowGraph, orderedNodeIDs []string) (map[string]string, map[string]bool) {
	nextByID := make(map[string]string, len(orderedNodeIDs))
	for i, id := range orderedNodeIDs {
		if i+1 < len(orderedNodeIDs) {
			nextByID[id] = orderedNodeIDs[i+1]
		}
	}

	incoming := make(map[string][]string)
	for origin, outs := range graph.outgoing {
		for _, edge := range outs {
			incoming[edge.Destination] = append(incoming[edge.Destination], origin)
		}
	}

	needsLabel := make(map[string]bool, len(orderedNodeIDs))
	for _, id := range orderedNodeIDs {
		preds := incoming[id]
		if len(preds) == 0 {
			continue
		}
		if len(preds) > 1 {
			needsLabel[id] = true
			continue
		}
		pred := preds[0]
		predOut := graph.outgoing[pred]
		if len(predOut) != 1 || nextByID[pred] != id {
			needsLabel[id] = true
		}
	}

	labelByID := make(map[string]string)
	labelCounter := 1
	for _, id := range orderedNodeIDs {
		if needsLabel[id] {
			labelByID[id] = fmt.Sprintf("L%03d", labelCounter)
			labelCounter++
		}
	}
	return labelByID, needsLabel
}

func findFirstExclusiveSplit(innerByID map[string]map[string]interface{}) (string, map[string]interface{}) {
	ids := make([]string, 0, len(innerByID))
	for id := range innerByID {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		obj := innerByID[id]
		objType, _ := obj["$Type"].(string)
		if objType == "Microflows$ExclusiveSplit" {
			return id, obj
		}
	}
	return "", nil
}

func collectLinearBranchLines(startID string, scope map[string]map[string]interface{}, outgoing map[string][]microflowFlow, graph microflowGraph) []string {
	lines := make([]string, 0)
	visited := make(map[string]bool)
	currentID := startID
	for currentID != "" {
		if visited[currentID] {
			break
		}
		visited[currentID] = true
		node, ok := scope[currentID]
		if !ok {
			break
		}
		nodeType, _ := node["$Type"].(string)
		if nodeType == "Microflows$ExclusiveSplit" {
			break
		}
		lines = append(lines, renderNodeInstruction(currentID, node, nodeType, graph)...)
		next := ""
		for _, flow := range outgoing[currentID] {
			if _, ok := scope[flow.Destination]; ok {
				next = flow.Destination
				break
			}
		}
		if next == "" {
			break
		}
		currentID = next
	}
	if len(lines) == 0 {
		lines = append(lines, "// no-op")
	}
	return lines
}

func getInnerEntryNodes(scope map[string]map[string]interface{}, outgoing map[string][]microflowFlow) []string {
	incoming := make(map[string]int, len(scope))
	for id := range scope {
		incoming[id] = 0
	}
	for origin, outs := range outgoing {
		if _, ok := scope[origin]; !ok {
			continue
		}
		for _, f := range outs {
			if _, ok := scope[f.Destination]; ok {
				incoming[f.Destination]++
			}
		}
	}
	entries := make([]string, 0)
	for id, count := range incoming {
		if count == 0 {
			entries = append(entries, id)
		}
	}
	slices.Sort(entries)
	if len(entries) == 0 {
		for id := range scope {
			entries = append(entries, id)
		}
		slices.Sort(entries)
	}
	return entries
}

func firstCaseValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(values[0]))
}

func extractNestedString(m map[string]interface{}, keys ...string) string {
	current := m
	for i, key := range keys {
		value, ok := current[key]
		if !ok {
			return ""
		}
		if i == len(keys)-1 {
			s, _ := value.(string)
			return s
		}
		next, ok := value.(map[string]interface{})
		if !ok {
			return ""
		}
		current = next
	}
	return ""
}

func extractReturnType(attributes map[string]interface{}) string {
	rt, ok := attributes["MicroflowReturnType"].(map[string]interface{})
	if !ok {
		return "unknown"
	}
	return formatDataType(rt)
}

func extractMicroflowParameters(objectsByID map[string]map[string]interface{}) []string {
	params := make([]string, 0)
	for _, obj := range objectsByID {
		objType, _ := obj["$Type"].(string)
		if objType != "Microflows$MicroflowParameter" {
			continue
		}
		name, _ := obj["Name"].(string)
		varType, _ := obj["VariableType"].(map[string]interface{})
		required, _ := obj["IsRequired"].(bool)
		requiredText := "optional"
		if required {
			requiredText = "required"
		}
		params = append(params, fmt.Sprintf("%s: %s (%s)", name, formatDataType(varType), requiredText))
	}
	slices.Sort(params)
	return params
}

func formatDataType(raw map[string]interface{}) string {
	typeName, _ := raw["$Type"].(string)
	switch typeName {
	case "DataTypes$VoidType":
		return "void"
	case "DataTypes$IntegerType":
		return "integer"
	case "DataTypes$BooleanType":
		return "boolean"
	case "DataTypes$StringType":
		return "string"
	case "DataTypes$ObjectType":
		entity, _ := raw["Entity"].(string)
		if entity == "" {
			return "object"
		}
		return entity
	default:
		if typeName == "" {
			return "unknown"
		}
		return typeName
	}
}

type microflowGraph struct {
	objectsByID map[string]map[string]interface{}
	outgoing    map[string][]microflowFlow
	startID     string
}

func buildMicroflowGraph(attributes map[string]interface{}) (microflowGraph, error) {
	objectsByID := make(map[string]map[string]interface{})
	outgoing := make(map[string][]microflowFlow)
	startID := ""

	objectCollectionRaw, ok := attributes["ObjectCollection"].(map[string]interface{})
	if !ok {
		return microflowGraph{}, fmt.Errorf("ObjectCollection not found")
	}
	objectsRaw, ok := objectCollectionRaw["Objects"].([]interface{})
	if !ok {
		return microflowGraph{}, fmt.Errorf("ObjectCollection.Objects not found")
	}

	for _, item := range objectsRaw {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id := readMicroflowID(obj["$ID"])
		if id == "" {
			continue
		}
		objectsByID[id] = obj
		if objType, _ := obj["$Type"].(string); objType == "Microflows$StartEvent" {
			startID = id
		}
	}
	if startID == "" {
		return microflowGraph{}, fmt.Errorf("start event not found")
	}

	flowsRaw, ok := attributes["Flows"].([]interface{})
	if !ok {
		return microflowGraph{}, fmt.Errorf("Flows not found")
	}

	for _, item := range flowsRaw {
		flowMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		origin := readMicroflowID(flowMap["OriginPointer"])
		dest := readMicroflowID(flowMap["DestinationPointer"])
		if origin == "" || dest == "" {
			continue
		}
		flow := microflowFlow{
			ID:          readMicroflowID(flowMap["$ID"]),
			Origin:      origin,
			Destination: dest,
			CaseValues:  extractCaseValues(flowMap["CaseValues"]),
		}
		outgoing[origin] = append(outgoing[origin], flow)
	}

	return microflowGraph{
		objectsByID: objectsByID,
		outgoing:    outgoing,
		startID:     startID,
	}, nil
}

func extractCaseValues(raw interface{}) []string {
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		c, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		v, _ := c["Value"].(string)
		if v != "" {
			values = append(values, v)
		}
	}
	return values
}

func traverseMicroflowNodes(startID string, outgoing map[string][]microflowFlow) []string {
	visited := make(map[string]bool)
	order := make([]string, 0)
	var visit func(string)
	visit = func(id string) {
		if id == "" || visited[id] {
			return
		}
		visited[id] = true
		order = append(order, id)
		for _, edge := range outgoing[id] {
			visit(edge.Destination)
		}
	}
	visit(startID)
	return order
}

func readMicroflowID(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return value
	case map[string]interface{}:
		// Raw YAML form:
		// $ID:
		//   subtype: 0
		//   data: [118, 185, ...]
		dataRaw, ok := value["data"]
		if !ok {
			return ""
		}
		parts, ok := dataRaw.([]interface{})
		if !ok {
			return ""
		}
		builder := strings.Builder{}
		for idx, p := range parts {
			if idx > 0 {
				builder.WriteByte('-')
			}
			builder.WriteString(intLikeToString(p))
		}
		return builder.String()
	default:
		return ""
	}
}

func intLikeToString(v interface{}) string {
	switch n := v.(type) {
	case int:
		return strconv.Itoa(n)
	case int32:
		return strconv.Itoa(int(n))
	case int64:
		return strconv.FormatInt(n, 10)
	case uint64:
		return strconv.FormatUint(n, 10)
	case float64:
		return strconv.Itoa(int(n))
	case float32:
		return strconv.Itoa(int(n))
	default:
		return ""
	}
}
