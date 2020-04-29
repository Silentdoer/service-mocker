package app

import (
	"errors"
	"flag"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/json-iterator/go"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	ph "path"
	"service-mocker/app/common"
	"service-mocker/app/constant"
	"service-mocker/app/setting"
	"strings"
	"time"
)

/*
先不实现监听app.json，因为动态增加或删除项目需要做的工作有点大，暂时只支持动态修改响应数据
*/
var autoRefresh bool

var projectSettings map[string]setting.ProjectSettings

var address string

var server *http.Server

var serveMuxHandler *http.ServeMux

func init() {
	file, err := os.Open(constant.APP_CONFIG_PATH)
	if err != nil {
		// 类似Java的throw ex
		panic("应用配置没有找到")
	}
	// _是一个默认变量，所以不需要_ := 来声明
	defer func() { _ = file.Close() }()
	configBytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic("读取应用配置失败")
	}
	var appSettings setting.AppSettings
	err = jsoniter.Unmarshal(configBytes, &appSettings)
	if err != nil {
		panic("应用配置不符合格式要求")
	}

	address = appSettings.Address

	tmp := linq.From(os.Args).WhereT(func(a string) bool {
		return strings.HasPrefix(a, fmt.Sprintf("-%s", constant.APP_AUTO_REFRESH_ARG))
	}).First()
	// 说明没有设置该参数（默认是不自动刷新的）
	if tmp == nil {
		autoRefresh = appSettings.AutoRefresh
	} else {
		// 真垃圾，连判断是否有某个option的方法都没有
		flag.BoolVar(&autoRefresh, constant.APP_AUTO_REFRESH_ARG, false, "是否自动刷新")
		// 真傻叉，居然对同一个option不能解析两次
		//tmp := flag.Bool(constant.APP_AUTO_REFRESH_ARG, true, "是否自动刷新")
		flag.Parse()
	}

	// 处理projects
	processProject(appSettings.MockProjects)
	// 判断是否有重复的项目（至于接口什么的太细了就不判断了）
	count := linq.From(appSettings.MockProjects).SelectT(func(p setting.ProjectSettings) string {
		return p.Name
	}).Distinct().Count()
	if count != len(appSettings.MockProjects) {
		panic(fmt.Errorf("%s里存在相同的项目配置", constant.APP_CONFIG_NAME))
	}

	projectSettings = make(map[string]setting.ProjectSettings)
	// 原来泛型参数的实现其实就是令整个接口都是空接口（所有接口也实现了空接口）
	linq.From(appSettings.MockProjects).ForEachT(func(p setting.ProjectSettings) {
		projectSettings[p.Name] = p
	})

	// 开启或关闭自动刷新
	toggleRefresh(autoRefresh)
}

func Start() {
	server = &http.Server{
		Addr:         address,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	serveMuxHandler = http.NewServeMux()
	initMockHandlers(serveMuxHandler)
	server.Handler = serveMuxHandler

	var errChan = make(chan error)
	go func() {
		err := <- errChan
		if err == nil {
			log.Println("Mocker服务已启动->", "http://" + address)
		} else {
			log.Fatal("Mocker服务启动失败:", err)
		}
	}()
	// 想法很好，可惜自己忘了一件事，，就是这里必须产生了错误才会返回。。。，所以上面的协程只能处理失败情况，这里再加个协程来发启动成功的信息吧
	go func() {
		// 1秒用于监听启动绝大多数都够了。。
		time.Sleep(1 * time.Second)
		errChan <- nil
	}()
	errChan <- server.ListenAndServe()
}

func Stop() {
	err := server.Close()
	if err != nil {
		panic(errors.New(fmt.Sprintf("关闭服务失败:%s\n", err.Error())))
	}
}

func initMockHandlers(mux *http.ServeMux) {
	//mux.Handle("/", loggingHandler(generateMockHandler("我是响应")))
	for _, val := range projectSettings {
		for _, api := range val.APIs {
			mux.Handle(ph.Join(val.Prefix, api.API), loggingHandler(generateMockHandler(fmt.Sprint(api.ResponseValue))))
		}
	}
}

func generateMockHandler(respStr string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(respStr))
		if err != nil {
			log.Println("响应数据", respStr, "失败")
		}
	})
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		starts := time.Now()
		log.Printf("Started->%s, %s\n", request.URL, request.Method)
		next.ServeHTTP(writer, request)
		log.Printf("Completed->%s in %v\n\n", request.URL, time.Since(starts))
	})
}

/*
单元测试其实就是单个的功能方法测试
*/
func toggleRefresh(flag bool) {
	if flag {

	} else {

	}
}

// 切片是指针类型
func processProject(settings []setting.ProjectSettings) {
	// recover()的defer最好是方法/函数里第一个defer，否则可能出现问题
	defer func() {
		// catch
		if r := recover(); r != nil {
			log.Fatal("解析project settings时产生错误:", r)
		}
	}()
	if settings == nil {
		panic(fmt.Errorf("project settings 是nil"))
	}
	for i := 0; i < len(settings); i++ {
		set := &settings[i]
		name := set.Name

		if len(strings.TrimSpace(name)) <= 0 {
			panic(errors.New("项目名不能为空"))
		}
		path := set.Path
		// 这里允许把project的配置写到app.json里（切片nil的len判断也是0）
		if len(set.APIs) > 0 && strings.HasPrefix(set.Prefix, constant.URI_START_CHAR) {
			continue
		}
		// prefix和APIs配置在project.json里
		projectConfigFile, _ := os.Open(ph.Join(path, constant.PROJECT_CONFIG_NAME))
		projectConfigBytes, _ := ioutil.ReadAll(projectConfigFile)
		_ = jsoniter.Unmarshal(projectConfigBytes, set)
		prefix := set.Prefix
		if !strings.HasPrefix(prefix, constant.URI_START_CHAR) {
			panic(fmt.Sprintf("%s里没有配置正确的prefix", constant.PROJECT_CONFIG_NAME))
		}

		// 这个apiMap又属于某个Project
		var apiMap = make(map[string]string, 10)
		// 解析判断API里的response是文件还是JSON对象
		for j := 0; j < len(set.APIs); j++ {
			api := &set.APIs[j]
			if api.ResponseValue != nil {
				valueBytes, _ := jsoniter.Marshal(api.ResponseValue)
				apiMap[api.API] = string(valueBytes)
			} else if len(api.ResponseRef) > 0 {
				respFile, _ := os.Open(ph.Join(set.Path, api.ResponseRef))
				// 这里通过一个方法来实现关闭文件就可以避免因为respFile不是defer方法参数而不会实时计算导致都是最后一个文件句柄
				// 不过循环里其实最好还是不要用defer，因为循环结束就可以释放了，不需要等那么久，万一循环很大不及时释放句柄
				// 可能导致程序获取不到句柄出错（这里是练手就不考虑这个问题了，知道就行），主要是这个是小程序也不会出现这个情况
				defer common.DeferClose(respFile)
				respBytes, _ := ioutil.ReadAll(respFile)
				apiMap[api.API] = string(respBytes)
			} else {
				panic(errors.New("接口没有配置响应数据"))
			}
			// 将responseValue赋值为可以直接返回的字符串
			api.ResponseValue = apiMap[api.API]
		}
		// 备注：此时appSettings对象已经解析了完整的所有配置
	}
}
