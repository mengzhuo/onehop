package onehop

type Tree struct {
	Min   RecordID
	Max   RecordID
	Limit int

	NodeList []*Node
}

func (t *Tree) SplitBy(newNode *Node) (left *Tree) {

	left = &Tree{Limit: t.Limit}

	mid_point := t.Max
	mid_point.RShift()

	left.Min = t.Min
	left.Max = mid_point

	t.Min = mid_point

	right_list := make([]*Node, 0)
	length := len(t.NodeList)

	for i := 0; i < length; i++ {

		n := t.NodeList[0]
		t.NodeList = t.NodeList[1:]

		if n.ID.Less(newNode.ID) {
			left.NodeList = append(left.NodeList, n)
		} else {
			right_list = append(right_list, n)
		}
	}

	if newNode.ID.Less(mid_point) {
		left.NodeList = append(left.NodeList, newNode)
	} else {
		t.NodeList = append(t.NodeList, newNode)
	}

	t.NodeList = right_list

	return
}
