package rtree

import "math"

type float float32

const d3roundValues = true // only set to true when using 32-bit floats

func d3fmin(a, b float) float {
	if a < b {
		return a
	}
	return b
}
func d3fmax(a, b float) float {
	if a > b {
		return a
	}
	return b
}

const (
	d3numDims            = 3
	d3maxNodes           = 13
	d3minNodes           = d3maxNodes / 2
	d3useSphericalVolume = true // Better split classification, may be slower on some systems
)

var d3unitSphereVolume = []float64{
	0.000000, 2.000000, 3.141593, // Dimension  0,1,2
	4.188790, 4.934802, 5.263789, // Dimension  3,4,5
	5.167713, 4.724766, 4.058712, // Dimension  6,7,8
	3.298509, 2.550164, 1.884104, // Dimension  9,10,11
	1.335263, 0.910629, 0.599265, // Dimension  12,13,14
	0.381443, 0.235331, 0.140981, // Dimension  15,16,17
	0.082146, 0.046622, 0.025807, // Dimension  18,19,20
}[d3numDims]

type d3RTree struct {
	root *d3nodeT ///< Root of tree
}

/// Minimal bounding rectangle (n-dimensional)
type d3rectT struct {
	min [d3numDims]float ///< Min dimensions of bounding box
	max [d3numDims]float ///< Max dimensions of bounding box
}

/// May be data or may be another subtree
/// The parents level determines this.
/// If the parents level is 0, then this is data
type d3branchT struct {
	rect  d3rectT     ///< Bounds
	child *d3nodeT    ///< Child node
	data  interface{} ///< Data Id or Ptr
}

/// d3nodeT for each branch level
type d3nodeT struct {
	count  int                   ///< Count
	level  int                   ///< Leaf is zero, others positive
	branch [d3maxNodes]d3branchT ///< Branch
}

func (node *d3nodeT) isInternalNode() bool {
	return (node.level > 0) // Not a leaf, but a internal node
}
func (node *d3nodeT) isLeaf() bool {
	return (node.level == 0) // A leaf, contains data
}

// Rounding constants for float32 -> float64 conversion.
const d3RNDTOWARDS = (1.0 - 1.0/8388608.0) // Round towards zero
const d3RNDAWAY = (1.0 + 1.0/8388608.0)    // Round away from zero

// Convert an sqlite3_value into an RtreeValue (presumably a float)
// while taking care to round toward negative or positive, respectively.
func d3rtreeValueDown(d float64) float {
	if !d3roundValues {
		return float(d)
	}
	f := float(d)
	if float64(f) > d {
		if d < 0 {
			f = float(d * d3RNDAWAY)
		} else {
			f = float(d * d3RNDTOWARDS)
		}
	}
	return f
}
func d3rtreeValueUp(d float64) float {
	if !d3roundValues {
		return float(d)
	}
	f := float(d)
	if float64(f) < d {
		if d < 0 {
			f = float(d * d3RNDTOWARDS)
		} else {
			f = float(d * d3RNDAWAY)
		}
	}
	return f
}

/// A link list of nodes for reinsertion after a delete operation
type d3listNodeT struct {
	next *d3listNodeT ///< Next in list
	node *d3nodeT     ///< Node
}

const d3notTaken = -1 // indicates that position

/// Variables for finding a split partition
type d3partitionVarsT struct {
	partition [d3maxNodes + 1]int
	total     int
	minFill   int
	count     [2]int
	cover     [2]d3rectT
	area      [2]float

	branchBuf      [d3maxNodes + 1]d3branchT
	branchCount    int
	coverSplit     d3rectT
	coverSplitArea float
}

func d3New() *d3RTree {
	// We only support machine word size simple data type eg. integer index or object pointer.
	// Since we are storing as union with non data branch
	return &d3RTree{
		root: &d3nodeT{},
	}
}

/// Insert entry
/// \param a_min Min of bounding rect
/// \param a_max Max of bounding rect
/// \param a_dataId Positive Id of data.  Maybe zero, but negative numbers not allowed.
func (tr *d3RTree) Insert(min, max [d3numDims]float64, dataId interface{}) {
	var branch d3branchT
	branch.data = dataId
	for axis := 0; axis < d3numDims; axis++ {
		branch.rect.min[axis] = d3rtreeValueDown(min[axis])
		branch.rect.max[axis] = d3rtreeValueUp(max[axis])
	}
	d3insertRect(&branch, &tr.root, 0)
}

/// Remove entry
/// \param a_min Min of bounding rect
/// \param a_max Max of bounding rect
/// \param a_dataId Positive Id of data.  Maybe zero, but negative numbers not allowed.
func (tr *d3RTree) Remove(min, max [d3numDims]float64, dataId interface{}) {
	var rect d3rectT
	for axis := 0; axis < d3numDims; axis++ {
		rect.min[axis] = d3rtreeValueDown(min[axis])
		rect.max[axis] = d3rtreeValueUp(max[axis])
	}
	d3removeRect(&rect, dataId, &tr.root)
}

/// Find all within d3search rectangle
/// \param a_min Min of d3search bounding rect
/// \param a_max Max of d3search bounding rect
/// \param a_searchResult d3search result array.  Caller should set grow size. Function will reset, not append to array.
/// \param a_resultCallback Callback function to return result.  Callback should return 'true' to continue searching
/// \param a_context User context to pass as parameter to a_resultCallback
/// \return Returns the number of entries found
func (tr *d3RTree) Search(min, max [d3numDims]float64, resultCallback func(data interface{}) bool) int {
	var rect d3rectT
	for axis := 0; axis < d3numDims; axis++ {
		rect.min[axis] = d3rtreeValueDown(min[axis])
		rect.max[axis] = d3rtreeValueUp(max[axis])
	}
	foundCount, _ := d3search(tr.root, rect, 0, resultCallback)
	return foundCount
}

/// Count the data elements in this container.  This is slow as no internal counter is maintained.
func (tr *d3RTree) Count() int {
	var count int
	d3countRec(tr.root, &count)
	return count
}

/// Remove all entries from tree
func (tr *d3RTree) RemoveAll() {
	// Delete all existing nodes
	tr.root = &d3nodeT{}
}

func d3countRec(node *d3nodeT, count *int) {
	if node.isInternalNode() { // not a leaf node
		for index := 0; index < node.count; index++ {
			d3countRec(node.branch[index].child, count)
		}
	} else { // A leaf node
		*count += node.count
	}
}

// Inserts a new data rectangle into the index structure.
// Recursively descends tree, propagates splits back up.
// Returns 0 if node was not split.  Old node updated.
// If node was split, returns 1 and sets the pointer pointed to by
// new_node to point to the new node.  Old node updated to become one of two.
// The level argument specifies the number of steps up from the leaf
// level to insert; e.g. a data rectangle goes in at level = 0.
func d3insertRectRec(branch *d3branchT, node *d3nodeT, newNode **d3nodeT, level int) bool {
	// recurse until we reach the correct level for the new record. data records
	// will always be called with a_level == 0 (leaf)
	if node.level > level {
		// Still above level for insertion, go down tree recursively
		var otherNode *d3nodeT
		//var newBranch d3branchT

		// find the optimal branch for this record
		index := d3pickBranch(&branch.rect, node)

		// recursively insert this record into the picked branch
		childWasSplit := d3insertRectRec(branch, node.branch[index].child, &otherNode, level)

		if !childWasSplit {
			// Child was not split. Merge the bounding box of the new record with the
			// existing bounding box
			node.branch[index].rect = d3combineRect(&branch.rect, &(node.branch[index].rect))
			return false
		} else {
			// Child was split. The old branches are now re-partitioned to two nodes
			// so we have to re-calculate the bounding boxes of each node
			node.branch[index].rect = d3nodeCover(node.branch[index].child)
			var newBranch d3branchT
			newBranch.child = otherNode
			newBranch.rect = d3nodeCover(otherNode)

			// The old node is already a child of a_node. Now add the newly-created
			// node to a_node as well. a_node might be split because of that.
			return d3addBranch(&newBranch, node, newNode)
		}
	} else if node.level == level {
		// We have reached level for insertion. Add rect, split if necessary
		return d3addBranch(branch, node, newNode)
	} else {
		// Should never occur
		return false
	}
}

// Insert a data rectangle into an index structure.
// d3insertRect provides for splitting the root;
// returns 1 if root was split, 0 if it was not.
// The level argument specifies the number of steps up from the leaf
// level to insert; e.g. a data rectangle goes in at level = 0.
// InsertRect2 does the recursion.
//
func d3insertRect(branch *d3branchT, root **d3nodeT, level int) bool {
	var newNode *d3nodeT

	if d3insertRectRec(branch, *root, &newNode, level) { // Root split

		// Grow tree taller and new root
		newRoot := &d3nodeT{}
		newRoot.level = (*root).level + 1

		var newBranch d3branchT

		// add old root node as a child of the new root
		newBranch.rect = d3nodeCover(*root)
		newBranch.child = *root
		d3addBranch(&newBranch, newRoot, nil)

		// add the split node as a child of the new root
		newBranch.rect = d3nodeCover(newNode)
		newBranch.child = newNode
		d3addBranch(&newBranch, newRoot, nil)

		// set the new root as the root node
		*root = newRoot

		return true
	}
	return false
}

// Find the smallest rectangle that includes all rectangles in branches of a node.
func d3nodeCover(node *d3nodeT) d3rectT {
	rect := node.branch[0].rect
	for index := 1; index < node.count; index++ {
		rect = d3combineRect(&rect, &(node.branch[index].rect))
	}
	return rect
}

// Add a branch to a node.  Split the node if necessary.
// Returns 0 if node not split.  Old node updated.
// Returns 1 if node split, sets *new_node to address of new node.
// Old node updated, becomes one of two.
func d3addBranch(branch *d3branchT, node *d3nodeT, newNode **d3nodeT) bool {
	if node.count < d3maxNodes { // Split won't be necessary
		node.branch[node.count] = *branch
		node.count++
		return false
	} else {
		d3splitNode(node, branch, newNode)
		return true
	}
}

// Disconnect a dependent node.
// Caller must return (or stop using iteration index) after this as count has changed
func d3disconnectBranch(node *d3nodeT, index int) {
	// Remove element by swapping with the last element to prevent gaps in array
	node.branch[index] = node.branch[node.count-1]
	node.branch[node.count-1].data = nil
	node.branch[node.count-1].child = nil
	node.count--
}

// Pick a branch.  Pick the one that will need the smallest increase
// in area to accommodate the new rectangle.  This will result in the
// least total area for the covering rectangles in the current node.
// In case of a tie, pick the one which was smaller before, to get
// the best resolution when searching.
func d3pickBranch(rect *d3rectT, node *d3nodeT) int {
	var firstTime bool = true
	var increase float
	var bestIncr float = -1
	var area float
	var bestArea float
	var best int
	var tempRect d3rectT

	for index := 0; index < node.count; index++ {
		curRect := &node.branch[index].rect
		area = d3calcRectVolume(curRect)
		tempRect = d3combineRect(rect, curRect)
		increase = d3calcRectVolume(&tempRect) - area
		if (increase < bestIncr) || firstTime {
			best = index
			bestArea = area
			bestIncr = increase
			firstTime = false
		} else if (increase == bestIncr) && (area < bestArea) {
			best = index
			bestArea = area
			bestIncr = increase
		}
	}
	return best
}

// Combine two rectangles into larger one containing both
func d3combineRect(rectA, rectB *d3rectT) d3rectT {
	var newRect d3rectT

	for index := 0; index < d3numDims; index++ {
		newRect.min[index] = d3fmin(rectA.min[index], rectB.min[index])
		newRect.max[index] = d3fmax(rectA.max[index], rectB.max[index])
	}

	return newRect
}

// Split a node.
// Divides the nodes branches and the extra one between two nodes.
// Old node is one of the new ones, and one really new one is created.
// Tries more than one method for choosing a partition, uses best result.
func d3splitNode(node *d3nodeT, branch *d3branchT, newNode **d3nodeT) {
	// Could just use local here, but member or external is faster since it is reused
	var localVars d3partitionVarsT
	parVars := &localVars

	// Load all the branches into a buffer, initialize old node
	d3getBranches(node, branch, parVars)

	// Find partition
	d3choosePartition(parVars, d3minNodes)

	// Create a new node to hold (about) half of the branches
	*newNode = &d3nodeT{}
	(*newNode).level = node.level

	// Put branches from buffer into 2 nodes according to the chosen partition
	node.count = 0
	d3loadNodes(node, *newNode, parVars)
}

// Calculate the n-dimensional volume of a rectangle
func d3rectVolume(rect *d3rectT) float {
	var volume float = 1
	for index := 0; index < d3numDims; index++ {
		volume *= rect.max[index] - rect.min[index]
	}
	return volume
}

// The exact volume of the bounding sphere for the given d3rectT
func d3rectSphericalVolume(rect *d3rectT) float64 {
	var sumOfSquares float64 = 0
	var radius float64

	for index := 0; index < d3numDims; index++ {
		halfExtent := float64(rect.max[index]-rect.min[index]) * 0.5
		sumOfSquares += halfExtent * halfExtent
	}

	radius = math.Sqrt(sumOfSquares)

	// Pow maybe slow, so test for common dims just use x*x, x*x*x.
	switch d3numDims {
	default:
		return (math.Pow(radius, d3numDims) * d3unitSphereVolume)
	case 2:
		return (radius * radius * d3unitSphereVolume)
	case 3:
		return (radius * radius * radius * d3unitSphereVolume)
	case 4:
		return (radius * radius * radius * radius * d3unitSphereVolume)
	case 5:
		return (radius * radius * radius * radius * radius * d3unitSphereVolume)
	}
}

// Use one of the methods to calculate retangle volume
func d3calcRectVolume(rect *d3rectT) float {
	if d3useSphericalVolume {
		return float(d3rectSphericalVolume(rect)) // Slower but helps certain merge cases
	} else { // RTREE_USE_SPHERICAL_VOLUME
		return d3rectVolume(rect) // Faster but can cause poor merges
	} // RTREE_USE_SPHERICAL_VOLUME
}

// Load branch buffer with branches from full node plus the extra branch.
func d3getBranches(node *d3nodeT, branch *d3branchT, parVars *d3partitionVarsT) {
	// Load the branch buffer
	for index := 0; index < d3maxNodes; index++ {
		parVars.branchBuf[index] = node.branch[index]
	}
	parVars.branchBuf[d3maxNodes] = *branch
	parVars.branchCount = d3maxNodes + 1

	// Calculate rect containing all in the set
	parVars.coverSplit = parVars.branchBuf[0].rect
	for index := 1; index < d3maxNodes+1; index++ {
		parVars.coverSplit = d3combineRect(&parVars.coverSplit, &parVars.branchBuf[index].rect)
	}
	parVars.coverSplitArea = d3calcRectVolume(&parVars.coverSplit)
}

// Method #0 for choosing a partition:
// As the seeds for the two groups, pick the two rects that would waste the
// most area if covered by a single rectangle, i.e. evidently the worst pair
// to have in the same group.
// Of the remaining, one at a time is chosen to be put in one of the two groups.
// The one chosen is the one with the greatest difference in area expansion
// depending on which group - the rect most strongly attracted to one group
// and repelled from the other.
// If one group gets too full (more would force other group to violate min
// fill requirement) then other group gets the rest.
// These last are the ones that can go in either group most easily.
func d3choosePartition(parVars *d3partitionVarsT, minFill int) {
	var biggestDiff float
	var group, chosen, betterGroup int

	d3initParVars(parVars, parVars.branchCount, minFill)
	d3pickSeeds(parVars)

	for ((parVars.count[0] + parVars.count[1]) < parVars.total) &&
		(parVars.count[0] < (parVars.total - parVars.minFill)) &&
		(parVars.count[1] < (parVars.total - parVars.minFill)) {
		biggestDiff = -1
		for index := 0; index < parVars.total; index++ {
			if d3notTaken == parVars.partition[index] {
				curRect := &parVars.branchBuf[index].rect
				rect0 := d3combineRect(curRect, &parVars.cover[0])
				rect1 := d3combineRect(curRect, &parVars.cover[1])
				growth0 := d3calcRectVolume(&rect0) - parVars.area[0]
				growth1 := d3calcRectVolume(&rect1) - parVars.area[1]
				diff := growth1 - growth0
				if diff >= 0 {
					group = 0
				} else {
					group = 1
					diff = -diff
				}

				if diff > biggestDiff {
					biggestDiff = diff
					chosen = index
					betterGroup = group
				} else if (diff == biggestDiff) && (parVars.count[group] < parVars.count[betterGroup]) {
					chosen = index
					betterGroup = group
				}
			}
		}
		d3classify(chosen, betterGroup, parVars)
	}

	// If one group too full, put remaining rects in the other
	if (parVars.count[0] + parVars.count[1]) < parVars.total {
		if parVars.count[0] >= parVars.total-parVars.minFill {
			group = 1
		} else {
			group = 0
		}
		for index := 0; index < parVars.total; index++ {
			if d3notTaken == parVars.partition[index] {
				d3classify(index, group, parVars)
			}
		}
	}
}

// Copy branches from the buffer into two nodes according to the partition.
func d3loadNodes(nodeA, nodeB *d3nodeT, parVars *d3partitionVarsT) {
	for index := 0; index < parVars.total; index++ {
		targetNodeIndex := parVars.partition[index]
		targetNodes := []*d3nodeT{nodeA, nodeB}

		// It is assured that d3addBranch here will not cause a node split.
		d3addBranch(&parVars.branchBuf[index], targetNodes[targetNodeIndex], nil)
	}
}

// Initialize a d3partitionVarsT structure.
func d3initParVars(parVars *d3partitionVarsT, maxRects, minFill int) {
	parVars.count[0] = 0
	parVars.count[1] = 0
	parVars.area[0] = 0
	parVars.area[1] = 0
	parVars.total = maxRects
	parVars.minFill = minFill
	for index := 0; index < maxRects; index++ {
		parVars.partition[index] = d3notTaken
	}
}

func d3pickSeeds(parVars *d3partitionVarsT) {
	var seed0, seed1 int
	var worst, waste float
	var area [d3maxNodes + 1]float

	for index := 0; index < parVars.total; index++ {
		area[index] = d3calcRectVolume(&parVars.branchBuf[index].rect)
	}

	worst = -parVars.coverSplitArea - 1
	for indexA := 0; indexA < parVars.total-1; indexA++ {
		for indexB := indexA + 1; indexB < parVars.total; indexB++ {
			oneRect := d3combineRect(&parVars.branchBuf[indexA].rect, &parVars.branchBuf[indexB].rect)
			waste = d3calcRectVolume(&oneRect) - area[indexA] - area[indexB]
			if waste > worst {
				worst = waste
				seed0 = indexA
				seed1 = indexB
			}
		}
	}

	d3classify(seed0, 0, parVars)
	d3classify(seed1, 1, parVars)
}

// Put a branch in one of the groups.
func d3classify(index, group int, parVars *d3partitionVarsT) {
	parVars.partition[index] = group

	// Calculate combined rect
	if parVars.count[group] == 0 {
		parVars.cover[group] = parVars.branchBuf[index].rect
	} else {
		parVars.cover[group] = d3combineRect(&parVars.branchBuf[index].rect, &parVars.cover[group])
	}

	// Calculate volume of combined rect
	parVars.area[group] = d3calcRectVolume(&parVars.cover[group])

	parVars.count[group]++
}

// Delete a data rectangle from an index structure.
// Pass in a pointer to a d3rectT, the tid of the record, ptr to ptr to root node.
// Returns 1 if record not found, 0 if success.
// d3removeRect provides for eliminating the root.
func d3removeRect(rect *d3rectT, id interface{}, root **d3nodeT) bool {
	var reInsertList *d3listNodeT

	if !d3removeRectRec(rect, id, *root, &reInsertList) {
		// Found and deleted a data item
		// Reinsert any branches from eliminated nodes
		for reInsertList != nil {
			tempNode := reInsertList.node

			for index := 0; index < tempNode.count; index++ {
				// TODO go over this code. should I use (tempNode->m_level - 1)?
				d3insertRect(&tempNode.branch[index], root, tempNode.level)
			}
			reInsertList = reInsertList.next
		}

		// Check for redundant root (not leaf, 1 child) and eliminate TODO replace
		// if with while? In case there is a whole branch of redundant roots...
		if (*root).count == 1 && (*root).isInternalNode() {
			tempNode := (*root).branch[0].child
			*root = tempNode
		}
		return false
	} else {
		return true
	}
}

// Delete a rectangle from non-root part of an index structure.
// Called by d3removeRect.  Descends tree recursively,
// merges branches on the way back up.
// Returns 1 if record not found, 0 if success.
func d3removeRectRec(rect *d3rectT, id interface{}, node *d3nodeT, listNode **d3listNodeT) bool {
	if node.isInternalNode() { // not a leaf node
		for index := 0; index < node.count; index++ {
			if d3overlap(*rect, node.branch[index].rect) {
				if !d3removeRectRec(rect, id, node.branch[index].child, listNode) {
					if node.branch[index].child.count >= d3minNodes {
						// child removed, just resize parent rect
						node.branch[index].rect = d3nodeCover(node.branch[index].child)
					} else {
						// child removed, not enough entries in node, eliminate node
						d3reInsert(node.branch[index].child, listNode)
						d3disconnectBranch(node, index) // Must return after this call as count has changed
					}
					return false
				}
			}
		}
		return true
	} else { // A leaf node
		for index := 0; index < node.count; index++ {
			if node.branch[index].data == id {
				d3disconnectBranch(node, index) // Must return after this call as count has changed
				return false
			}
		}
		return true
	}
}

// Decide whether two rectangles d3overlap.
func d3overlap(rectA, rectB d3rectT) bool {
	for index := 0; index < d3numDims; index++ {
		if rectA.min[index] > rectB.max[index] ||
			rectB.min[index] > rectA.max[index] {
			return false
		}
	}
	return true
}

// Add a node to the reinsertion list.  All its branches will later
// be reinserted into the index structure.
func d3reInsert(node *d3nodeT, listNode **d3listNodeT) {
	newListNode := &d3listNodeT{}
	newListNode.node = node
	newListNode.next = *listNode
	*listNode = newListNode
}

// d3search in an index tree or subtree for all data retangles that d3overlap the argument rectangle.
func d3search(node *d3nodeT, rect d3rectT, foundCount int, resultCallback func(data interface{}) bool) (int, bool) {
	if node.isInternalNode() {
		// This is an internal node in the tree
		for index := 0; index < node.count; index++ {
			if d3overlap(rect, node.branch[index].rect) {
				var ok bool
				foundCount, ok = d3search(node.branch[index].child, rect, foundCount, resultCallback)
				if !ok {
					// The callback indicated to stop searching
					return foundCount, false
				}
			}
		}
	} else {
		// This is a leaf node
		for index := 0; index < node.count; index++ {
			if d3overlap(rect, node.branch[index].rect) {
				id := node.branch[index].data
				foundCount++
				if !resultCallback(id) {
					return foundCount, false // Don't continue searching
				}

			}
		}
	}
	return foundCount, true // Continue searching
}
