package onehop

type Unit struct {
	Min   RecordID
	Max   RecordID
	Limit int

	NodeList []*Node
}

func (u *Unit) SplitBy(newNode *Node) (left *Unit) {

	left = &Unit{Limit: u.Limit}

	mid_point := u.Max
	mid_point.RShift()

	left.Min = u.Min
	left.Max = mid_point

	u.Min = mid_point

	right_list := make([]*Node, 0)
	length := len(u.NodeList)

	for i := 0; i < length; i++ {

		n := u.NodeList[0]
		u.NodeList = u.NodeList[1:]

		if n.ID.Less(newNode.ID) {
			left.NodeList = append(left.NodeList, n)
		} else {
			right_list = append(right_list, n)
		}
	}

	if newNode.ID.Less(mid_point) {
		left.NodeList = append(left.NodeList, newNode)
	} else {
		u.NodeList = append(u.NodeList, newNode)
	}

	u.NodeList = right_list

	return
}
