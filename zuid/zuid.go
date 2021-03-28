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
	Node int64 // 机器位数
	Step int64 // 序列位数
}

type Node struct {
	mutex sync.Mutex
	rand  *rand.Rand

	bits      Bits
	node      int64
	step      int64
	time      int64
	ms        int64
	epoch     time.Time
	nodeMax   int64
	nodeMask  int64
	stepMask  int64
	randShift int64
	timeShift int64
	nodeShift int64
}

type ID int64

func NewBits(rand, time, node, step int64) Bits {
	return Bits{
		Rand: rand,
		Time: time,
		Node: node,
		Step: step,
	}
}

// TODO 支持不同位排序
func NewNode(epoch, ms, node int64, bits Bits) *Node {
	var n Node
	var now = time.Now()

	n.rand = rand.New(rand.NewSource(now.UnixNano()))

	n.bits = bits
	n.node = node
	n.epoch = now.Add(time.Unix(epoch/1000, epoch%1000*1000000).Sub(now))
	n.ms = ms
	n.nodeMax = -1 ^ (-1 << bits.Node)
	n.nodeMask = n.nodeMax << bits.Step
	n.stepMask = -1 ^ (-1 << bits.Step)
	n.randShift = bits.Time + bits.Node + bits.Step
	n.timeShift = bits.Node + bits.Step
	n.nodeShift = bits.Step

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
	return NewNode(epoch, 1000, node, bits)
}

// 标准雪花算法 - Twitter
func Snowflake(epoch, node int64) *Node {
	var bits = Bits{
		Rand: 0,
		Time: 41,
		Node: 10,
		Step: 12,
	}
	return NewNode(epoch, 1, node, bits)
}

// 索尼雪花算法
func Sonyflake(epoch, node int64) *Node {
	var bits = Bits{
		Rand: 0,
		Time: 39,
		Node: 8,
		Step: 16,
	}
	return NewNode(epoch, 10, node, bits)
}

func (n *Node) now() int64 {
	return time.Since(n.epoch).Milliseconds() / n.ms
}

func (n *Node) bit() int64 {
	return n.bits.Time + n.bits.Node + n.bits.Step
}

func (n *Node) mask() int64 {
	return -1 ^ -1<<n.bit()
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
	return []byte(id.String()), nil
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

// 方案一: Shotflake 共:34bit 单机:4个/秒 | ---- 30bit(时间:秒,使用时长34年) ---- | ---- 2bit(机器号,最多4台) ---- | ---- 2bit(序列号,最多4个/秒) ---- |
// 方案二: Yuanflake 共:40bit 单机:16个/秒 | ---- 32bit(时间:秒,使用时长136年) ---- | ---- 4bit(机器号,最多16台) ---- | ---- 4bit(序列号,最多16个/秒) ---- |
// 方案三: Sonyflake 共:63bit 单机:65536个/10毫秒 | ---- 39bit(时间:10毫秒,使用时长174年) ---- | ---- 8bit(机器号,最多256台) ---- | ---- 16bit(序列号,最多65536个/10毫秒) ---- |
// 方案四: Snowflake 共:63bit 单机:4096个/毫秒 | ---- 41bit(时间:毫秒,使用时长69年) ---- | ---- 10bit(机器号,最多1024台) ---- | ---- 12bit(序列号,最多4096个/毫秒) ---- |
