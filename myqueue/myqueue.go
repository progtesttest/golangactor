package myqueue

import (
//	"testactor/queue/goring"
	"sync/atomic"
	"runtime"
	"testactor/myqueue/queue/mpsc"
//	"testactor/queue/goring"
	"time"
	"log"
)

const (
	idle int32 = iota
	running
)

////////////////////////
type Dispatcher interface {
	Schedule(fn func())
	Throughput() int
}

type goroutineDispatcher int

func (goroutineDispatcher) Schedule(fn func()) {
	go fn()
}

func (d goroutineDispatcher) Throughput() int {
	return int(d)
}

func NewDefaultDispatcher(throughput int) Dispatcher {
	return goroutineDispatcher(throughput)
}


type synchronizedDispatcher int

func (synchronizedDispatcher) Schedule(fn func()) {
	fn()
}

func (d synchronizedDispatcher) Throughput() int {
	return int(d)
}

func NewSynchronizedDispatcher(throughput int) Dispatcher {
	return synchronizedDispatcher(throughput)
}

/////////////////////////

type Statistics interface {
	MessageReceived(message interface{})
}


type MyQueue struct {
	userMailbox     *mpsc.Queue
	schedulerStatus int32
	userMessages    int32
	suspended       bool
	dispatcher      Dispatcher
	mailboxStats    []Statistics
}

func (m *MyQueue) PostUserMessage(message interface{}) {
//	for _, ms := range m.mailboxStats {
//		ms.MessagePosted(message)
//	}
	m.userMailbox.Push(message)
	atomic.AddInt32(&m.userMessages, 1)
	//m.schedule()
	if atomic.CompareAndSwapInt32(&m.schedulerStatus, idle, running) {
		m.dispatcher.Schedule(m.processMessages)
	}
}


func (m *MyQueue) processMessages() {

process:
	m.run()

	// set mailbox to idle
	atomic.StoreInt32(&m.schedulerStatus, idle)
	user := atomic.LoadInt32(&m.userMessages)
	// check if there are still messages to process (sent after the message loop ended)
	if  user > 0 {
		// try setting the mailbox back to running
		if atomic.CompareAndSwapInt32(&m.schedulerStatus, idle, running) {
			//	fmt.Printf("looping %v %v %v\n", sys, user, m.suspended)
			goto process
		}
	}

}

func (m *MyQueue) run() {
	var msg interface{}
//	var bok bool

	defer func() {
		if r := recover(); r != nil {
		//	plog.Debug("[ACTOR] Recovering", log.Object("actor", m.invoker), log.Object("reason", r), log.Stack())
		//	m.invoker.EscalateFailure(r, msg)
		}
	}()

	i, t := 0, 1024
	for {
		if i > t {
			i = 0
			runtime.Gosched()
		}

		i++

		if msg = m.userMailbox.Pop(); msg != nil {
			atomic.AddInt32(&m.userMessages, -1)
			if m.mailboxStats[0] != nil {
				m.mailboxStats[0].MessageReceived(msg)
			}
		}else {
			return
		}
	}

}

func UnboundedMyQueue(dispatcher Dispatcher,mailboxStats ...Statistics) *MyQueue {

	return &MyQueue{
		userMailbox:   mpsc.New(),
		mailboxStats: mailboxStats,
		dispatcher:	dispatcher,
	}
}

type MySynMsgChan struct {
	MsgData  interface{}
	MsgChan chan interface{}
	ChanState int32   // 0 关闭  1 开启
}

//生成同步消息   return true OK  false timeout
//func  SendMySynMsgAndWait(dtimeout time.Duration,msg interface{},)(interface{},bool){

//}
func GenerateMySynMsg(data interface{})* MySynMsgChan  {

	return &MySynMsgChan{data,make(chan interface{},1),1}
}

//return true OK  false timeout
func (self *MySynMsgChan)WaitSynReply(d time.Duration) (interface{},bool) {
	if d < time.Millisecond  {
		d = time.Millisecond
	}
	timetick := time.NewTimer(d)
	for {
		select {
			case <- timetick.C:
			//	atomic.StoreInt32(&self.ChanState,0)
				return nil,false
			case msg := <- self.MsgChan :
			//	atomic.StoreInt32(&self.ChanState,0)
				timetick.Reset(0  * time.Second)  //设置过期 防止 导致goroutine和内存泄露
				return  msg,true
		}
	}
	return nil,false
}

