package zuid

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Bits struct {
	Rand int64 // 随机位数
	Time int64 // 时间位数
	Node int64 // 节点位数
	Step int64 // 序列位数
}

type Sort struct {
	Rand int64 // 随机位权重
	Time int64 // 时间位权重
	Node int64 // 节点位权重
	Step int64 // 序列位权重
}

type Node struct {
	mutex sync.Mutex
	rand  *rand.Rand

	bits  Bits
	sort  Sort
	node  int64
	step  int64
	time  int64
	ms    int64
	epoch time.Time

	randMax   int64
	timeMax   int64
	nodeMax   int64
	stepMax   int64
	randMask  int64
	timeMask  int64
	nodeMask  int64
	stepMask  int64
	randShift int64
	timeShift int64
	nodeShift int64
	stepShift int64
}

type ID int64

func NewBits(rand, time, node, step int64) Bits {
	var bits = Bits{
		Rand: rand,
		Time: time,
		Node: node,
		Step: step,
	}

	if total := bits.Total(); total < 5 || total > 63 {
		panic(fmt.Errorf("total bits must between 5 and 63"))
	}

	return bits
}

func NewSort(rand, time, node, step int64) Sort {
	var sort = Sort{
		Rand: rand,
		Time: time,
		Node: node,
		Step: step,
	}

	if sort.Count() < 3 {
		panic(fmt.Errorf("sort counts must greater than or equal to 3"))
	}

	return sort
}

func bitShift(bits Bits, sorts Sort, sort int64) (bit int64) {
	if sorts.Rand < sort {
		bit += bits.Rand
	}
	if sorts.Time < sort {
		bit += bits.Time
	}
	if sorts.Node < sort {
		bit += bits.Node
	}
	if sorts.Step < sort {
		bit += bits.Step
	}
	return
}

func NewNode(epoch, ms, node int64, bits Bits, sort Sort) *Node {
	// 校验失败则panic
	bits = NewBits(bits.Rand, bits.Time, bits.Node, bits.Step)
	sort = NewSort(sort.Rand, sort.Time, sort.Node, sort.Step)

	var n Node
	var now = time.Now()

	n.rand = rand.New(rand.NewSource(now.UnixNano()))

	n.bits = bits
	n.sort = sort
	n.node = node
	n.epoch = now.Add(time.Unix(epoch/1000, epoch%1000*1000000).Sub(now))
	n.ms = ms

	n.randMax = -1 ^ (-1 << bits.Rand)
	n.timeMax = -1 ^ (-1 << bits.Time)
	n.nodeMax = -1 ^ (-1 << bits.Node)
	n.stepMax = -1 ^ (-1 << bits.Step)

	n.randShift = bitShift(bits, sort, sort.Rand)
	n.timeShift = bitShift(bits, sort, sort.Time)
	n.nodeShift = bitShift(bits, sort, sort.Node)
	n.stepShift = bitShift(bits, sort, sort.Step)

	n.randMask = n.randMax << n.randShift
	n.nodeMask = n.nodeMax << n.nodeShift
	n.timeMask = n.timeMax << n.timeShift
	n.stepMask = n.stepMax << n.stepShift

	if node < 0 || node > n.nodeMax {
		panic(fmt.Errorf("node number must be between 0 and %v", n.nodeMax))
	}

	return &n
}

// 短ID雪花算法
func Shotflake(epoch, node int64) *Node {
	var bits = Bits{
		Rand: 0,
		Time: 30,
		Node: 2,
		Step: 2,
	}
	var sort = Sort{
		Rand: 4,
		Time: 3,
		Node: 2,
		Step: 1,
	}
	return NewNode(epoch, 1000, node, bits, sort)
}

// 标准雪花算法 - Twitter
func Snowflake(epoch, node int64) *Node {
	var bits = Bits{
		Rand: 0,
		Time: 41,
		Node: 10,
		Step: 12,
	}
	var sort = Sort{
		Rand: 4,
		Time: 3,
		Node: 2,
		Step: 1,
	}
	return NewNode(epoch, 1, node, bits, sort)
}

// 索尼雪花算法
func Sonyflake(epoch, node int64) *Node {
	var bits = Bits{
		Rand: 0,
		Time: 39,
		Node: 8,
		Step: 16,
	}
	var sort = Sort{
		Rand: 4,
		Time: 3,
		Node: 2,
		Step: 1,
	}
	return NewNode(epoch, 10, node, bits, sort)
}

func (b Bits) Total() int64 {
	return b.Rand + b.Time + b.Node + b.Step
}

func (s Sort) Count() int {
	return len(map[int64]struct{}{
		s.Rand: {},
		s.Time: {},
		s.Node: {},
		s.Step: {},
	})
	var set = make(map[int64]struct{})

	set[s.Rand] = struct{}{}
	set[s.Time] = struct{}{}
	set[s.Node] = struct{}{}
	set[s.Step] = struct{}{}

	return len(set)
}

func (n *Node) now() int64 {
	return time.Since(n.epoch).Milliseconds() / n.ms
}

func (n *Node) GenID() ID {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	var now = n.now()

	if now == n.time {
		if n.step = (n.step + 1) & n.stepMask; n.step == 0 {
			for now <= n.time {
				now = n.now()
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	var random int64
	if n.bits.Rand != 0 {
		random = n.rand.Int63() % (-1 ^ (-1 << n.bits.Rand) + 1)
	}

	var id = random<<n.randShift | now<<n.timeShift | n.node<<n.nodeShift | n.step

	return ID(id)
}

func (id ID) Int64() int64 {
	return int64(id)
}

func (id ID) Bytes() []byte {
	return []byte(id.String())
}

func (id ID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *ID) UnmarshalJSON(data []byte) (err error) {
	var (
		i   int64
		str string
	)

	if err = json.Unmarshal(data, &str); err != nil {
		return
	}

	if i, err = strconv.ParseInt(str, 10, 64); err != nil {
		return
	}

	*id = ID(i)

	return
}
