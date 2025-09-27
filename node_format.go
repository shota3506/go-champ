package champ

import (
	"fmt"
	"strings"
)

func indent(depth int) string {
	return strings.Repeat("  ", depth)
}

// nodeString returns an indented string representation of the node.
func nodeString[K Key, V any](n node[K, V], depth int, shift uint) string {
	if n == nil {
		return fmt.Sprintf("%s<nil>", indent(depth))
	}

	switch node := n.(type) {
	case *bitmapIndexedNode[K, V]:
		return bitmapIndexedNodeString(node, depth, shift)
	case *collisionNode[K, V]:
		return collisionNodeString(node, depth)
	default:
		return fmt.Sprintf("%s<unknown node type>", indent(depth))
	}
}

func bitmapIndexedNodeString[K Key, V any](n *bitmapIndexedNode[K, V], depth int, shift uint) string {
	itob := func(datamap uint32, index int) int {
		count := -1
		for bit := range 32 {
			if datamap&(1<<uint(bit)) != 0 {
				count++
				if count == index {
					return bit
				}
			}
		}
		return -1
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%sBitmapNode[shift=%d]{\n", indent(depth), shift))
	sb.WriteString(fmt.Sprintf("%s  datamap: 0b%032b (0x%08x)\n", indent(depth), n.datamap, n.datamap))
	sb.WriteString(fmt.Sprintf("%s  nodemap: 0b%032b (0x%08x)\n", indent(depth), n.nodemap, n.nodemap))

	if len(n.keys) > 0 {
		sb.WriteString(fmt.Sprintf("%s  data[%d]:\n", indent(depth), len(n.keys)))
		for i := 0; i < len(n.keys); i++ {
			bit := itob(n.datamap, i)
			sb.WriteString(fmt.Sprintf("%s    [bit %2d] %v => %v\n",
				indent(depth), bit, n.keys[i], n.values[i]))
		}
	}
	if len(n.nodes) > 0 {
		sb.WriteString(fmt.Sprintf("%s  nodes[%d]:\n", indent(depth), len(n.nodes)))
		for i := 0; i < len(n.nodes); i++ {
			bit := itob(n.nodemap, i)
			sb.WriteString(fmt.Sprintf("%s    [bit %2d]:\n", indent(depth), bit))
			sb.WriteString(nodeString(n.nodes[i], depth+2, shift+bitsPerLevel))
		}
	}

	sb.WriteString(fmt.Sprintf("%s}", indent(depth)))
	return sb.String()
}

func collisionNodeString[K Key, V any](n *collisionNode[K, V], depth int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%sCollisionNode{\n", indent(depth)))
	sb.WriteString(fmt.Sprintf("%s  entries[%d]:\n", indent(depth), len(n.keys)))

	for i := 0; i < len(n.keys); i++ {
		sb.WriteString(fmt.Sprintf("%s    %v => %v\n",
			indent(depth), n.keys[i], n.values[i]))
	}

	sb.WriteString(fmt.Sprintf("%s}", indent(depth)))
	return sb.String()
}
