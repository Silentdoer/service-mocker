package main

/*import (
	"log"
	"github.com/fsnotify/fsnotify"
	"github.com/json-iterator/go"
	"io/ioutil"
	"service-mocker/app/util"
	"net/http"
)*/
import (
	"service-mocker/app"
	_ "service-mocker/app"
)

/*
思路：
1、默认是自动刷新的，可以在启动参数里关闭
2、分层级，程序同级的config.json里配置有哪些待模拟的项目，对应的路径是什么（一般用相对路径在同级目录里，当然绝对也行）

 */

/*
全局初始化函数
 */
func init() {

}

/*
程序入口
 */
var ARGS string

/*
不要用jsoniter，有bug，至少jsoniter.MarshalToString(appSettings)方法是有bug的，ResponseRef最后少了个"
其他方法貌似是OK的
 */
func main() {
	defer app.Stop()
	//tmp := flag.Bool(constant.APP_AUTO_REFRESH_ARG, false, "是否自动刷新")
	//flag.Parse()
	//println(*tmp)
	app.Start()
}