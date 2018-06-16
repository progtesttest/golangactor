package main

import (
	"runtime"
	"fmt"
	"log"
	"time"
	"math/rand"
	"github.com/progtesttest/myqueue"
	"sync/atomic"
)

type Acc struct {
	Accname string
	Token	string
}

type Player struct {
	Userid uint32
	Accname string
	baseinfo []byte
}

type Roledata struct {
	acc  *Acc
	player * Player
	msgque  *myqueue.MyQueue
}

type RoledataExpend struct {
	acc  *Acc
	player * Player
	msgque  *myqueue.MyQueue
	index   int
}

type RoleChanInfo struct {
	Accname string
	Userid uint32
	msgque  *myqueue.MyQueue
}

type AccAppendChan struct {
	Accname string
	Token	string
	msgque  *myqueue.MyQueue
}

type TrandPlay struct {
	accname  string
	msgque  *myqueue.MyQueue
}


type AccMgr struct {
	msgque  *myqueue.MyQueue
	rolemap  map[string] * RoleChanInfo
	autoUserid  uint32
}

var (
	accmsgr *AccMgr  = nil
)

func SetupLogicMgr()  {
	accmsgr  = new(AccMgr)
	accmsgr.msgque  = myqueue.UnboundedMyQueue(myqueue.NewDefaultDispatcher(300),accmsgr)
	accmsgr.rolemap  = make(map[string] * RoleChanInfo,128)
}

func SendMsgToAccMgr(v interface{}) {
	accmsgr.msgque.PostUserMessage(v)
}

func (self* AccMgr)MessageReceived(r interface{})  {

	switch r.(type) {
	case *AccAppendChan :
		accpendchan := r.(*AccAppendChan)
		if accpendchan.msgque != nil {
			if v,ok := self.rolemap[accpendchan.Accname];ok {
				if v.msgque != nil {
					msgQue :=  v.msgque
					v.msgque = accpendchan.msgque
					//self.rolemap[accpendchan.Accname] = &RoleChanInfo{v.Accname,v.Userid,accpendchan.msgque}
					//v.msgque
					//msgQue.Push(&TrandPlay{accpendchan.Accname,accpendchan.msgque})
					msgQue.PostUserMessage(&TrandPlay{accpendchan.Accname,accpendchan.msgque})
				}
				//accpendchan.msgque.Push(&Roledata{v.acc,v.player})
				//accpendchan.msgque.Push(&TrandPlay{accpendchan.Accname,accpendchan.msgque})
				//delete(self.rolemap,accpendchan.Accname)

				//self.rolemap[accpendchan.Accname] = &RoleChanInfo{roledata.acc.Accname,roledata.player.Userid,accpendchan.msgque}
			}else {
				//从DB get
				roledata := &Roledata{}
				//	if rand.Intn(100) >= 0 {  //DB 里有数据
				roledata.acc= &Acc{accpendchan.Accname,accpendchan.Token}
				roledata.player = &Player{self.autoUserid,accpendchan.Accname,[]byte{}}
				self.autoUserid++
				//	}else {  //没角色
				//		roledata.acc= &Acc{accpendchan.Accname,accpendchan.Token}
				//	}
				self.rolemap[accpendchan.Accname] = &RoleChanInfo{roledata.acc.Accname,roledata.player.Userid,accpendchan.msgque}
				//accpendchan.msgque.Push(roledata)
				accpendchan.msgque.PostUserMessage(roledata)

			}
		}
	case *myqueue.MySynMsgChan:
		syndata := r.(*myqueue.MySynMsgChan)
		log.Printf("recv syn request data=%v \n ",syndata)
		switch syndata.MsgData.(type) {
		case string:
			if atomic.LoadInt32(&syndata.ChanState)  > 0  {
				strvalue := syndata.MsgData.(string)
				strvalue  += "world"
				syndata.MsgChan <- strvalue
				log.Printf("send syn msg %v \n",strvalue)
			}
		}
	}
}

//client
//SendMsgToAccMgr(&AccAppendChan{acc.Accname,acc.Token,que})
func (self* RoledataExpend)MessageReceived(r interface{})  {

	switch r.(type) {
	case  *Roledata :
		roledata := r.(*Roledata)
		self.acc = roledata.acc
		if roledata.player != nil {
			self.player = roledata.player
			//
		}else {
			//告诉客户端需要创建角色
		//	log.Printf("ERROR accname=%v \n",self.acc.Accname)
		//	log.Fatal("ERROR")
		}
	//	log.Printf("get roledata accname=%v  \n",self.acc.Accname)
	case *TrandPlay :
		tranddata := r.(*TrandPlay)
	//	log.Printf("exit role acc=%v user=%v accname=%v \n",self.acc,self.player,tranddata.accname)
		if tranddata.msgque != nil {
			//if self.player  == nil {
			//	log.Printf("accname=%v errormsg=%v \n",tranddata.accname,arraymsg)
			//}
		//	log.Printf("userid=%v accname=%v trans data \n",self.player.Userid,self.player.Accname)
			//tranddata.msgque.Push(&Roledata{self.acc,self.player})
			tranddata.msgque.PostUserMessage(&Roledata{self.acc,self.player,nil})
		}

	}
}

func testfun()  {
  data := make(chan string,1)
	log.Printf("send data=data hello \n")
	data <- "data hello"
  go func() {
  	 time.Sleep(2*time.Second)
	  tick := time.NewTimer(time.Millisecond*500)
	  for {
		  select {
		  case <- tick.C :
		  	log.Printf("timeout  \n")
			  return
		  case msg := <- data :
			  log.Printf("recv data=%v \n",msg)
			  return
		  }
	  }
  }()
  time.Sleep(5*time.Second)
}

func main()  {
	runtime.GOMAXPROCS(runtime.NumCPU())  //runtime.NumCPU()
	rand.Seed(time.Now().UnixNano())
	testfun()
	SetupLogicMgr()
	maxcount := 1
	for  i := 0; i< maxcount ;i++ {
		//go clientlogic(i)
		//roledata := new(Roledata)
		//go roledata.clientlogic(i)
		roledata := new(RoledataExpend)
		roledata.acc = &Acc{fmt.Sprintf("test%v",i),"123456"}
		roledata.index = i
		roledata.msgque  = myqueue.UnboundedMyQueue(myqueue.NewDefaultDispatcher(300),roledata)
		SendMsgToAccMgr(&AccAppendChan{roledata.acc.Accname,roledata.acc.Token,roledata.msgque}) //异步
	}
	log.Printf("curgoruntime1=%v \n",runtime.NumGoroutine())
	time.Sleep(2*time.Second)
	log.Printf("curgoruntime2=%v \n",runtime.NumGoroutine())

	for  i := 0; i< maxcount ;i++ {
		//go clientlogic(i)
		//roledata := new(Roledata)
		//go roledata.clientlogic(i)
		roledata := new(RoledataExpend)
		roledata.acc = &Acc{fmt.Sprintf("test%v",i),"123456"}
		roledata.index = i
		roledata.msgque  = myqueue.UnboundedMyQueue(myqueue.NewDefaultDispatcher(300),roledata)
		SendMsgToAccMgr(&AccAppendChan{roledata.acc.Accname,roledata.acc.Token,roledata.msgque})
	}
	time.Sleep(2*time.Second)
	for  i := maxcount-1; i >= 0 ;i-- {
		//go clientlogic(i)
		//roledata := new(Roledata)
		//go roledata.clientlogic(i)
		roledata := new(RoledataExpend)
		roledata.acc = &Acc{fmt.Sprintf("test%v",i),"123456"}
		roledata.index = i
		roledata.msgque  = myqueue.UnboundedMyQueue(myqueue.NewDefaultDispatcher(300),roledata)
		SendMsgToAccMgr(&AccAppendChan{roledata.acc.Accname,roledata.acc.Token,roledata.msgque})
	}

	time.Sleep(2*time.Second)
	//同步消息
	log.Printf("test syn msg \n")
	synactor := myqueue.GenerateMySynMsg("hello")
	SendMsgToAccMgr(synactor)

	if v,ok:= synactor.WaitSynReply(time.Second); ok {
		log.Printf("reply=%v \n",v.(string))
	}else{
		log.Printf("reply timeout \n")
	}

	synactor1 := myqueue.GenerateMySynMsg("hello[1]")
	SendMsgToAccMgr(synactor1)
	time.Sleep(2*time.Second)
	if v,ok:= synactor1.WaitSynReply(1*time.Second); ok {
		log.Printf("reply1=%v \n",v.(string))
	}else{
		log.Printf("reply1 timeout \n")
	}


	var input string
	fmt.Scanln(&input)
}


/////////////////////////////////////////////////////////
