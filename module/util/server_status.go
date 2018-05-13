/**
 * @description:
 * @author: jarekzha@gmail.com
 * @date: 2018/4/26
 */
package util

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/utils"
	"math"
	"time"
)

const ServerStatusExpireTime = time.Second * 30

func statusDataKey(serverId string) string {
	return "status_" + serverId
}

// 保存服务器状态
func SaveServerStatus(app module.App, serverId string, running bool, loadHash int64) {
	//log.Debug("SaveServerStatus serverId(%s) running(%t) loadHash(%d)", serverId, running, loadHash)
	config := app.GetSettings()
	url := config.Rpc.SessionRedisUrl
	if url == "" {
		// 未配置不处理
		return
	}

	key := statusDataKey(serverId)
	status := &module.ServerStatus{
		Running:  running,
		LoadHash: loadHash,
	}
	value, err := json.Marshal(status)
	if err != nil {
		log.Error("SaveServerStatus (%s) status marshal fail: (%s)", key, err)
		return
	}

	pool := utils.GetRedisFactory().GetPool(url).Get()
	defer pool.Close()

	seconds := int(ServerStatusExpireTime.Seconds())

	_, err = pool.Do("SET", key, value, "EX", seconds)
	if err != nil {
		log.Error("SaveServerStatus (%s) expire (%f) fail: (%s)", key, seconds, err)
		return
	}

	return
}

// 读取服务器状态
func ReadServerStatus(app module.App, serverId string) (status *module.ServerStatus) {
	status = &module.ServerStatus{
		Running:  true,
		LoadHash: 0,
	}

	config := app.GetSettings()
	url := config.Rpc.SessionRedisUrl
	if url == "" {
		// 未配置不处理
		return
	}

	pool := utils.GetRedisFactory().GetPool(url).Get()
	defer pool.Close()

	key := statusDataKey(serverId)
	data, err := redis.Bytes(pool.Do("GET", key))
	if err == redis.ErrNil || data == nil {
		status.Running = false
		return
	} else if err != nil {
		log.Error("GetServerStatus (%s) fail: (%s)", key, err)
		return
	}

	err = json.Unmarshal(data, status)
	if err != nil {
		log.Error("GetServerStatus (%s) unmarshal fail: (%s)", key, err)
	}

	return status
}

/**
获取指定类型的运行中的所有模块
*/
func GetRunningServersByType(app module.App, moduleType string) []module.ServerSession {
	servers := app.GetServersByType(moduleType)

	rets := make([]module.ServerSession, 0)
	for _, session := range servers {
		if session.GetServerStatus().Running {
			rets = append(rets, session)
		}
	}
	return rets
}

/**
获取指定类型的运行中的负载最低模块
*/
func GetMinLoadRunningServerByType(app module.App, moduleType string) module.ServerSession {
	servers := app.GetServersByType(moduleType)

	var ret module.ServerSession
	var minLoad = int64(math.MaxInt64)

	for _, session := range servers {
		status := session.GetServerStatus()
		if status.Running {
			if status.LoadHash < minLoad {
				minLoad = status.LoadHash
				ret = session
			}
		}
	}
	return ret
}
