/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
 */
package module
import (
	"github.com/liangdas/mqant/conf"
	"net/http"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module/modules/master"
	"sync"
	"net"
	"encoding/json"
	"io/ioutil"
	"strings"
	"strconv"
	"io"
)
var MasterModule = func() (Module){
	master := new(Master)
	return master
}


type HttpResponse struct {
	Error	string
	Code	string
	Result	interface{}
}
func (h *HttpResponse)String()(string){
	b,err:=json.Marshal(h)
	if err==nil{
		return string(b)
	}else{
		resp:=&HttpResponse{
			Code:"fail",
			Result:err.Error(),
		}
		return resp.String()
	}
}


func NewHttpResponse(Code string,Result interface{}) (*HttpResponse) {
	resp:=&HttpResponse{
		Code:Code,
		Result:Result,
	}
	return resp
}

func NewErrorResponse(Code string,Error string) (*HttpResponse) {
	resp:=&HttpResponse{
		Code:Code,
		Error:Error,
	}
	return resp
}

/**
每一个模块的最新汇报信息
 */
type ModuleReport struct {
	ModuleType	string
	Id		string
	Version		string
	ProcessID	string
	Executing	int64	//当前正在执行的函数数量,暂态的,下一次上报时刷新
	ReportForm	map[string]*StatisticalMethod //运行状态报表
}

type Master struct {
	BaseModule
	app 		App
	listener 	net.Listener
	ProcessMap	map[string]*master.Process
	ModuleReports   map[string]*ModuleReport	//moduleID -- ModuleReport
	rwmutex 	sync.RWMutex
}
func (m *Master) GetType()(string){
	//很关键,需要与配置文件中的Module配置对应
	return "Master"
}
func (m *Master) Version()(string){
	return "1.0.0"
}

func (m *Master) OnInit(app App,settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m,app,settings)
	m.app=app
	m.ModuleReports=map[string]*ModuleReport{}
	for Type,modSettings:=range conf.Conf.Module {
		for _,setting:=range modSettings{
			if _,ok:=m.ModuleReports[setting.Id];ok{
				//
			}else{
				reportForm:=&ModuleReport{
					Id:setting.Id,
					ProcessID:setting.ProcessID,
					ModuleType:Type,
					ReportForm:nil,
				}
				m.ModuleReports[setting.Id]=reportForm
			}
		}
	}

	m.ProcessMap=map[string]*master.Process{}
	for _,psetting :=range app.GetSettings().Master.Process{
		ps:=new(master.Process)
		ps.Init(app.GetSettings().Master,psetting)
		m.ProcessMap[psetting.ProcessID]=ps
	}
	m.GetServer().RegisterGO("HD_Start_Process",m.startProcess)
	m.GetServer().RegisterGO("HD_Stop_Process",m.stopProcess)
	m.GetServer().RegisterGO("ReportForm",m.ReportForm)
}

func (m *Master) Run(closeSig chan bool) {
	if m.app.GetSettings().Master.WebHost!=""{
		//app := golf.New()
		//app.Static("/", m.app.GetSettings().Master.WebRoot)
		//app.Run(m.app.GetSettings().Master.WebHost)
		l, _ := net.Listen("tcp", m.app.GetSettings().Master.WebHost)
		m.listener = l
		go func() {
			log.Info("Master web server Listen : %s",m.app.GetSettings().Master.WebHost)
			http.Handle("/",http.StripPrefix("/", http.FileServer(http.Dir(m.app.GetSettings().Master.WebRoot))))
			http.HandleFunc("/api/process/list.json",m.ProcessList)
			http.HandleFunc("/api/process/state/update.json",m.UpdateProcessState)
			http.HandleFunc("/api/process/start.json",m.StartProcess)
			http.HandleFunc("/api/process/stop.json",m.StopProcess)
			http.HandleFunc("/api/module/list.json",m.ModuleList)
			http.Serve(m.listener, nil)
		}()
		<-closeSig
		log.Info("Master web server Shutting down...")
		m.listener.Close()
	}

}

func (m *Master) GetArgs(req *http.Request)(map[string]string){
	req.ParseForm() //解析参数，默认是不会解析的
	args:=map[string]string{}
	if req.Method == "GET" {
		for k, v := range req.Form {
			args[k]=strings.Join(v, "")
		}
	} else if req.Method == "POST" {
		result, _ := ioutil.ReadAll(req.Body)
		req.Body.Close()
		//未知类型的推荐处理方法
		var f interface{}
		json.Unmarshal(result, &f)
		m := f.(map[string]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				args[k]=vv
			case int:

			case float64:

			case []interface{}:

			default:

			}
		}
	}
	return args
}

/**
获取进程状态
 */
func (m *Master) ProcessList(w http.ResponseWriter, req *http.Request) {
	req.BasicAuth()
	args:=m.GetArgs(req)
	Host:=args["host"]
	ProcessID:=args["pid"]
	State:=args["state"]
	list:=[]map[string]interface{}{}
	for _,process:=range m.ProcessMap{
		if Host!=""&&Host!=process.Process.Host{
			continue
		}
		if ProcessID!=""&&ProcessID!=process.Process.ProcessID{
			continue
		}
		if State!=""{
			s,err := strconv.Atoi(State)
			if err == nil{
				if s!=process.State{
					continue
				}
			}
		}
		list=append(list,map[string]interface{}{
			"State":process.State,
			"ProcessID":process.Process.ProcessID,
			"Host":process.Process.Host,
		})
	}
	response:=NewHttpResponse("success",list)
	io.WriteString(w, response.String())
}
/**
获取模块状态
 */
func (m *Master) ModuleList(w http.ResponseWriter, req *http.Request) {
	args:=m.GetArgs(req)

	ModuleType:=args["type"]
	ProcessID:=args["pid"]
	ModuleID:=args["mid"]
	list:=[]map[string]interface{}{}
	for _,module:=range m.ModuleReports{
		if ModuleType!=""&&ModuleType!=module.ModuleType{
			continue
		}
		if ProcessID!=""&&ProcessID!=module.ProcessID{
			continue
		}
		if ModuleID!=""&&ModuleID!=module.Id{
			continue
		}
		list=append(list,map[string]interface{}{
			"ProcessID":module.ProcessID,
			"ModuleType":module.ModuleType,
			"ModuleID":module.Id,
			"Version":module.Version,
			"Executing":module.Executing,
			"ReportForm":module.ReportForm,
		})
	}
	response:=NewHttpResponse("success",list)
	io.WriteString(w, response.String())

}

/**
刷新进程状态
 */
func (m *Master) UpdateProcessState(w http.ResponseWriter, req *http.Request){
	args:=m.GetArgs(req)
	Host:=args["host"]
	ProcessID:=args["pid"]
	State:=args["state"]
	for _,process:=range m.ProcessMap{
		if Host!=""&&Host!=process.Process.Host{
			continue
		}
		if ProcessID!=""&&ProcessID!=process.Process.ProcessID{
			continue
		}
		if State!=""{
			s,err := strconv.Atoi(State)
			if err == nil{
				if s!=process.State{
					continue
				}
			}
		}
		process.StateUpdate()
	}
	response:=NewHttpResponse("success","job run")
	io.WriteString(w, response.String())
}

/**
启动进程
 */
func (m *Master) StartProcess(w http.ResponseWriter, req *http.Request){
	args:=m.GetArgs(req)
	Host:=args["host"]
	ProcessID:=args["pid"]
	if Host==""&&ProcessID==""{
		response:=NewErrorResponse("fail","You must specify host or ProcessID")
		io.WriteString(w, response.String())
		return
	}
	for _,process:=range m.ProcessMap{
		if Host!=""&&Host!=process.Process.Host{
			continue
		}
		if ProcessID!=""&&ProcessID!=process.Process.ProcessID{
			continue
		}
		process.Start()
	}
	response:=NewHttpResponse("success","job run")
	io.WriteString(w, response.String())
}

/**
停止进程
 */
func (m *Master) StopProcess(w http.ResponseWriter, req *http.Request){
	args:=m.GetArgs(req)
	Host:=args["host"]
	ProcessID:=args["pid"]
	if Host==""&&ProcessID==""{
		response:=NewErrorResponse("fail","You must specify host or ProcessID")
		io.WriteString(w, response.String())
		return
	}
	for _,process:=range m.ProcessMap{
		if Host!=""&&Host!=process.Process.Host{
			continue
		}
		if ProcessID!=""&&ProcessID!=process.Process.ProcessID{
			continue
		}
		process.Stop()
	}
	response:=NewHttpResponse("success","job run")
	io.WriteString(w, response.String())
}


func (m *Master) OnDestroy() {
	//一定别忘了关闭RPC
	m.GetServer().OnDestroy()
}


/**
根据ProcessID 启动一个远程进程
 */
func (m *Master) startProcess(s map[string]interface{},msg map[string]interface{})(result string,err string){
	ProcessID:=msg["ProcessID"].(string)
	if Process,ok:=m.ProcessMap[ProcessID];ok{
		_,err=Process.Start()
		result="执行了启动命令"
	}else{
		err="配置文件中没有这个进程"
	}

	return
}

/**
根据ProcessID 启动一个远程进程
 */
func (m *Master) stopProcess(s map[string]interface{},msg map[string]interface{})(result string,err string){
	ProcessID:=msg["ProcessID"].(string)
	if Process,ok:=m.ProcessMap[ProcessID];ok{
		_,err=Process.Stop()
		result="执行了停止命令"
	}else{
		err="配置文件中没有这个进程"
	}

	return
}
/**
模块汇报
 */
func (m *Master) ReportForm(moduleType string,ProcessID string,Id string,Version string,statistics string,Executing int64)(result string,err string){
	sm:=LoadStatisticalMethod(statistics)
	if sm==nil{
		err="JSON format is not correct"
	}
	m.rwmutex.RLock()
	if reportForm,ok:=m.ModuleReports[Id];ok{
		reportForm.ProcessID=ProcessID
		reportForm.ModuleType=moduleType
		reportForm.Executing=Executing
		reportForm.Version=Version
		reportForm.ReportForm=sm
	}else{
		reportForm:=&ModuleReport{
			Id:Id,
			Version:Version,
			ProcessID:ProcessID,
			ModuleType:moduleType,
			Executing:Executing,
			ReportForm:sm,
		}
		m.ModuleReports[Id]=reportForm
	}
	m.rwmutex.RUnlock()
	result="success"
	return
}