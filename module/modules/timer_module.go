/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/
package modules

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/modules/timer"
	"time"
)

var TimerModule = func() module.Module {
	Timer := new(Timer)
	return Timer
}

type Timer struct {
	module.Module
}

func (m *Timer) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Timer"
}

func (m *Timer) OnInit(app module.App, settings *conf.ModuleSettings) {
	timewheel.SetTimeWheel(timewheel.New(10*time.Millisecond, 36))
	// 时间轮使用方式
	//import "github.com/liangdas/mqant/module/modules/timer"
	//执行过的定时器会自动被删除
	//timewheel.GetTimeWheel().AddTimer(66 * time.Millisecond , nil,self.Update)
	//
	//timewheel.GetTimeWheel().AddTimerCustom(66 * time.Millisecond ,"baba", nil,self.Update)
	//删除一个为执行的定时器, 参数为添加定时器传递的唯一标识
	//timewheel.GetTimeWheel().RemoveTimer("baba")
}

func (m *Timer) Run(closeSig chan bool) {
	timewheel.GetTimeWheel().Start(closeSig)
}

func (m *Timer) OnDestroy() {
}
