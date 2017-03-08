/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
 */
package module
import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module/modules/timer"
)
var TimerModule = func() (Module){
	Timer := new(Timer)
	return Timer
}

type Timer struct {
	Module
}
func (m *Timer) GetType()(string){
	//很关键,需要与配置文件中的Module配置对应
	return "Timer"
}

func (m *Timer) OnInit(app App,settings *conf.ModuleSettings) {
	// 定时器2，不传参数
	//SetTimer("callback2", 100, m.callback2, time.Now().UnixNano())
}
//func (m *Timer)callback2(args interface{}) {
//	//每次在当前时间点之后5s插入一个定时器，这样就能形成每隔5秒调用一次callback2回调函数，可以用于周期性事件
//	SetTimer("callback2", 60, m.callback2, time.Now().UnixNano())
//	log.Debug("time",(time.Now().UnixNano()-args.(int64))/1000000)
//	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
//}
func (m *Timer) Run(closeSig chan bool) {
	timer.Run(closeSig)
}

func (m *Timer) OnDestroy() {
}
