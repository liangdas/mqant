package timer

import (
	"sync"
	"time"
	"github.com/liangdas/mqant/log"
)

const wheel_cnt uint8 = 5                                                                   //时间轮数量5个
var element_cnt_per_wheel = [wheel_cnt]uint32{256, 64, 64, 64, 64}                          //每个时间轮的槽(元素)数量。在 256+64+64+64+64 = 512 个槽中，表示的范围为 2^32
var right_shift_per_wheel = [wheel_cnt]uint32{8, 6, 6, 6, 6}                                //当指针指向当前时间轮最后一位数，再走一位就需要向上进位。每个时间轮进位的时候，使用右移的方式，最快实现进位。这里是每个轮的进位二进制位数
var base_per_wheel = [wheel_cnt]uint32{1, 256, 256 * 64, 256 * 64 * 64, 256 * 64 * 64 * 64} //记录每个时间轮指针当前指向的位置
var mutex sync.Mutex                                                                        //加锁
var rwmutex sync.RWMutex
var newest [wheel_cnt]uint32                           //每个时间轮当前指针所指向的位置
var timewheels [5][]*Node                              //定义5个时间轮
//var TimerMap map[string]*Node = make(map[string]*Node) //保存待执行的计时器，方便按链表节点指针地址直接删除定时器

type mtimer struct {
	//Name        string            //定时器名称
	Inteval     uint32            //时间间隔，即以插入该定时器的时间为起点，Inteval秒之后执行回调函数DoSomething()。例如进程插入该定时器的时间是2015-04-05 10:23:00，Inteval=5，则执行DoSomething()的时间就是2015-04-05 10:23:05。
	DoSomething func(interface{}) //自定义事件处理函数，需要触发的事件
	Args        interface{}       //上述函数的输入参数
}

func SetTimer(inteval uint32, handler func(interface{}), args interface{}) {
	if inteval <= 0 {
		return
	}
	var bucket_no uint8 = 0
	var offset uint32 = inteval
	var left uint32 = inteval
	for offset >= element_cnt_per_wheel[bucket_no] { //偏移量大于当前时间轮容量，则需要向高位进位
		offset >>= right_shift_per_wheel[bucket_no] //计算高位的值。偏移量除以低位的进制。比如低位当前是256，则右移8个二进制位，就是除以256，得到的结果是高位的值。
		var tmp uint32 = 1
		if bucket_no == 0 {
			tmp = 0
		}
		left -= base_per_wheel[bucket_no] * (element_cnt_per_wheel[bucket_no] - newest[bucket_no] - tmp)
		bucket_no++
	}
	if offset < 1 {
		return
	}
	if inteval < base_per_wheel[bucket_no]*offset {
		return
	}
	left -= base_per_wheel[bucket_no] * (offset - 1)
	pos := (newest[bucket_no] + offset) % element_cnt_per_wheel[bucket_no] //通过类似hash的方式，找到在时间轮上的插入位置

	var node Node
	node.SetData(mtimer{left, handler, args})

	rwmutex.RLock()
	//TimerMap[name] = timewheels[bucket_no][pos].InsertHead(node) //插入定时器
	timewheels[bucket_no][pos].InsertHead(node) //插入定时器
	rwmutex.RUnlock()
	//fmt.Println("pos ", bucket_no, pos, tmp)
}

func step() {
	//var dolist list.List
	{
		rwmutex.RLock()
		//遍历所有桶
		var bucket_no uint8 = 0
		for bucket_no = 0; bucket_no < wheel_cnt; bucket_no++ {
			newest[bucket_no] = (newest[bucket_no] + 1) % element_cnt_per_wheel[bucket_no] //当前指针递增1
			//fmt.Println(newest)
			var head *Node = timewheels[bucket_no][newest[bucket_no]] //返回当前指针指向的槽位置的表头
			var firstElement *Node = head.Next()
			for firstElement != nil { //链表不为空
				if value, ok := firstElement.Data().(mtimer); ok { //如果element里面确实存储了Timer类型的数值，那么ok返回true，否则返回false。
					inteval := value.Inteval
					doSomething := value.DoSomething
					args := value.Args
					if nil != doSomething { //有遇到函数为nil的情况，所以这里判断下非nil
						if 0 == bucket_no || 0 == inteval {
							//dolist.PushBack(value) //执行自定义处理函数
							go func() {
								defer func() {
									if r := recover(); r != nil {
										log.Error("TimerError %v", r)
									}
								}()
								doSomething(args)
							}()
						} else {
							SetTimer(inteval, doSomething, args) //重新插入计时器
						}
					}
					Delete(firstElement) //删除定时器
				}
				firstElement = head.Next() //重新定位到链表第一个元素头
			}
			if 0 != newest[bucket_no] { //指针不是0，还未转回到原点，跳出。如果回到原点，则说明转完了一圈，需要向高位进位1，则继续循环入高位步进一步。
				break
			}
		}
		rwmutex.RUnlock()
	}
}

func Run(closeSig chan bool) {
	tick := time.NewTicker(1 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			step()
		case <-closeSig:
			return
		}
	}
}

func init() { //初始化
	var bucket_no uint8 = 0
	for bucket_no = 0; bucket_no < wheel_cnt; bucket_no++ {
		var i uint32 = 0
		for ; i < element_cnt_per_wheel[bucket_no]; i++ {
			timewheels[bucket_no] = append(timewheels[bucket_no], new(Node))
		}
	}
}
