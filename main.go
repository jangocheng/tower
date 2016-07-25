package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/admpub/confl"
	"github.com/admpub/log"
)

func init() {
	log.DefaultLog.Category = `tower`
	log.DefaultLog.SyncMode = true
	log.DefaultLog.Formatter = func(_ *log.Logger, e *log.Entry) string {
		return e.Message
	}
}

const ConfigName = ".tower.yml"

var (
	_appMainFile   *string
	_appPort       *string
	_pxyPort       *string
	_appBuildDir   *string
	_portParamName *string
	_runParams     *string
	_verbose       *bool
	_configFile    *string
	_adminPwd      *string
	_adminIPs      *string
	_pxyEngine     *string
	_autoClear     *bool
	_logLevel      *string

	app   App
	build string = "1"
)

func main() {
	_appMainFile = flag.String("m", "", "path to your app's main file.")
	_appPort = flag.String("p", "5001-5050", "port range of your app.")
	_pxyPort = flag.String("r", "8080", "proxy port of your app.")
	_pxyEngine = flag.String("e", "fast", "fast/standard")
	_appBuildDir = flag.String("o", "", "save the executable file the folder.")
	_portParamName = flag.String("n", "", "app's port param name.")
	_runParams = flag.String("s", "", "app's run params.")
	_verbose = flag.Bool("v", false, "show more stuff.")
	_configFile = flag.String("c", ConfigName, "yaml configuration file location.")
	_adminPwd = flag.String("w", "", "admin password.")
	_adminIPs = flag.String("i", "127.0.0.1,::1", "admin allow IP.")
	_autoClear = flag.Bool("a", true, "automatically deletes previously compiled files when you startup Tower in the compile mode")
	_logLevel = flag.String("log", "Info", "logger level(Debug/Info/Warn/Error/Fatal)")

	flag.Parse()

	args := flag.Args()
	if len(args) == 1 && args[0] == "init" {
		generateExampleConfig()
		return
	}
	if !fileExist(*_configFile) {
		generateExampleConfig()
	}
	startTower()
}

func fileExist(path string) bool {
	fi, err := os.Stat(path)
	return (err == nil || os.IsExist(err)) && !fi.IsDir()
}

func saveFile(filePath string, b []byte) (int, error) {
	os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	fw, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer fw.Close()
	return fw.Write(b)
}

func generateExampleConfig() {
	var configContent []byte
	if atob(build) {
		configContent = defaultDevConfig
	} else {
		configContent = defaultProdConfig
	}
	_, err := saveFile(ConfigName, configContent)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("== Generated config file " + ConfigName)
}

func atob(a string) bool {
	return a == `1` || a == `true` || a == `on` || a == `yes`
}

func findBinFile(f string) string {
	var prefix, suffix string
	tg := strings.Split(filepath.Base(f), `*`)
	switch len(tg) {
	case 2:
		prefix = tg[0]
		suffix = tg[1]
	default:
		panic(`error format.`)
	}
	var file string
	err := filepath.Walk(filepath.Dir(f), func(filePath string, info os.FileInfo, e error) (err error) {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return
		}
		name := info.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
			file = filePath
			return filepath.SkipDir
		}
		return
	})
	if err != nil && err != filepath.SkipDir {
		panic(err)
	}
	return file
}

func checkBinFile(appMainFile string, suffix string, _suffix *string, appBuildDir *string) error {
	_, err := os.Stat(appMainFile)
	if err != nil {
		if *appBuildDir == `` {
			return err
		}
		appMainFile = filepath.Join(*appBuildDir, appMainFile)
		_, err = os.Stat(appMainFile)
		if err != nil {
			return err
		}
	}
	appMainFile, err = filepath.Abs(appMainFile)
	if err != nil {
		return err
	}
	if *appBuildDir == `` {
		*appBuildDir = filepath.Dir(appMainFile)
	}
	fileName := filepath.Base(appMainFile)
	AppBin = fileName
	if strings.HasSuffix(AppBin, suffix) {
		AppBin = strings.TrimSuffix(AppBin, suffix)
		*_suffix = suffix
	}
	nameOk := strings.HasPrefix(AppBin, BinPrefix)
	if nameOk {
		fileName := strings.TrimPrefix(AppBin, BinPrefix)
		_, err := strconv.ParseInt(fileName, 10, 64)
		if err != nil {
			nameOk = false
		}
	}
	if !nameOk {
		return fmt.Errorf("exec参数指定的可执行文件名称格式应该为：%v0%v(当前为：%v)。\n其中的“0”是代表版本号的整数，请修改为此格式。", BinPrefix, *_suffix, fileName)
	}
	return nil
}

func startTower() {
	var (
		appMainFile        = *_appMainFile
		appPort            = *_appPort
		pxyPort            = *_pxyPort
		pxyEngine          = *_pxyEngine
		appBuildDir        = *_appBuildDir
		portParamName      = *_portParamName
		runParams          = *_runParams
		configFile         = *_configFile
		verbose            = *_verbose
		adminPwd           = *_adminPwd
		adminIPs           = *_adminIPs
		autoClear          = *_autoClear
		allowBuild         = atob(build)
		logLevel           = *_logLevel
		suffix             = ".exe"
		_suffix            = ""
		watchedFiles       string
		watchedOtherDir    string
		ignoredPathPattern string
		offlineMode        bool
		disabledLogRequest bool
	)
	if configFile == "" {
		configFile = ConfigName
	}
	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Error(err)
	} else {
		newmap := map[string]string{}
		yamlErr := confl.Unmarshal(contents, &newmap)
		if yamlErr != nil {
			log.Error(yamlErr)
		}
		appPort, _ = newmap["app_port"]
		pxyPort, _ = newmap["pxy_port"]
		if v, ok := newmap["pxy_engine"]; ok {
			pxyEngine = v
		}
		if v, ok := newmap["auto_clear"]; ok {
			autoClear = atob(v)
		}
		if v, ok := newmap["log_level"]; ok {
			logLevel = v
		}
		appBuildDir, _ = newmap["app_buildDir"]
		portParamName, _ = newmap["app_portParamName"]
		runParams, _ = newmap["app_runParams"]
		watchedFiles, _ = newmap["watch"]
		watchedOtherDir, _ = newmap["watch_otherDir"] //编译模式下有效
		ignoredPathPattern, _ = newmap["watch_ignoredPath"]
		offlineModeStr, _ := newmap["offline_mode"]

		if v, ok := newmap["admin_pwd"]; ok {
			adminPwd = v
		}
		if v, ok := newmap["admin_ip"]; ok {
			adminIPs = v
		}
		if atob(offlineModeStr) {
			offlineMode = true
		}
		if logRequestStr, ok := newmap["log_request"]; ok {
			disabledLogRequest = atob(logRequestStr) == false
		}
		if pxyPort == "" {
			pxyPort = ProxyPort
		}
		if allowBuild {
			appMainFile, _ = newmap["main"] //编译模式下有效
		} else {
			appMainFile, _ = newmap["exec"] //非编译模式下有效
			if appMainFile == "" {
				log.Error("请设置exec参数用来指定执行文件位置")
				time.Sleep(time.Second * 10)
				return
			}
		}
	}

	if verbose {
		logLevel = `Debug`
	}

	log.DefaultLog.SetLevel(logLevel)

	err = dialAddress("127.0.0.1:"+pxyPort, 1)
	if err == nil {
		log.Error("Error: port (" + pxyPort + ") already in used.")
		os.Exit(1)
	}
	if !allowBuild {
		if strings.Contains(appMainFile, `*`) {
			orgiMainFile := appMainFile
			appMainFile = findBinFile(appMainFile)
			if appMainFile == `` {
				if appBuildDir != `` {
					appMainFile = filepath.Join(appBuildDir, orgiMainFile)
					appMainFile = findBinFile(appMainFile)
				}
			}
		}
		if err := checkBinFile(appMainFile, suffix, &_suffix, &appBuildDir); err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 300)
			return
		}
	} else {
		if appBuildDir == `` {
			appMainFile, _ = filepath.Abs(appMainFile)
			appBuildDir = filepath.Dir(appMainFile)
		}
		if autoClear {
			err := filepath.Walk(appBuildDir, func(filePath string, info os.FileInfo, e error) (err error) {
				if e != nil {
					return e
				}
				if info.IsDir() {
					return
				}
				name := info.Name()
				if strings.HasPrefix(name, BinPrefix) {
					err = os.Remove(filePath)
					if err != nil {
						return
					}
				}
				return
			})
			if err != nil {
				log.Error(err)
			}
		}
	}
	app = NewApp(appMainFile, appPort, appBuildDir, portParamName)
	app.OfflineMode = offlineMode
	app.DisabledLogRequest = disabledLogRequest
	if runParams != `` {
		app.RunParams = strings.Split(runParams, ` `)
	}
	watchedDir := app.Root
	if !allowBuild {
		if app.BuildDir != `` {
			watchedDir = app.BuildDir
		}
	}
	if watchedOtherDir != "" {
		watchedDir = watchedOtherDir + "|" + watchedDir
	}
	watcher := NewWatcher(watchedDir, watchedFiles, ignoredPathPattern)
	proxy := NewProxy(&app, &watcher)
	proxy.AdminPwd = adminPwd
	proxy.Engine = pxyEngine
	if adminIPs != `` {
		proxy.AdminIPs = strings.Split(adminIPs, `,`)
	}
	if allowBuild {
		watcher.OnChanged = func(file string) {
			log.Debug(`== Build Mode.`)
			watcher.Reset()
			fileName := filepath.Base(file)
			if strings.HasPrefix(fileName, BinPrefix) {
				log.Info(`忽略`, fileName, `更改`)
				return
			}
			if !app.SupportMutiPort() {
				log.Error(`Unspecified switchable other ports.`)
				return
			}
			port := app.UseRandPort()
			for i := 0; i < 3 && port == app.Port; i++ {
				app.Clean()
				time.Sleep(time.Second)
				port = app.UseRandPort()
			}
			if port == app.Port {
				log.Error(`取得的端口与当前端口相同，无法编译切换`)
				return
			}
			err = app.Start(true, port)
			if err != nil {
				log.Error(err)
			}
		}
	} else {
		watcher.OnChanged = func(file string) {
			log.Debug(`== Switch Mode.`)
			watcher.Reset()
			if !app.SupportMutiPort() {
				log.Error(`Unspecified switchable other ports.`)
				return
			}
			port := app.UseRandPort()
			for i := 0; i < 3 && port == app.Port; i++ {
				app.Clean()
				time.Sleep(time.Second)
				port = app.UseRandPort()
			}
			if port == app.Port {
				log.Error(`取得的端口与当前端口相同，无法切换`)
				return
			}

			fileName := filepath.Base(file)
			if !strings.HasPrefix(fileName, BinPrefix) {
				log.Info(`忽略非`, BinPrefix, `前缀文件更改`)
				return
			}
			if _suffix != "" {
				fileName = strings.TrimSuffix(fileName, _suffix)
			}
			newAppBin := fileName
			fileName = strings.TrimPrefix(fileName, BinPrefix)
			newFileTs, err := strconv.ParseInt(fileName, 10, 64)
			if err != nil {
				log.Error(err)
				return
			}
			fileName = strings.TrimPrefix(AppBin, BinPrefix)
			oldFileTs, err := strconv.ParseInt(fileName, 10, 64)
			if err != nil {
				log.Error(err)
				return
			}
			if newFileTs <= oldFileTs {
				log.Info(`新文件时间戳小于旧文件，忽略`)
				return
			}
			AppBin = newAppBin
			err = app.Start(true, port)
			if err != nil {
				log.Error(err)
			}
		}
		watcher.OnlyWatchBin = true
		app.DisabledBuild = true
	}
	proxy.Port = pxyPort
	go func() {
		mustSuccess(watcher.Watch())
	}()
	err = app.Start(true, app.Port)
	if err != nil {
		log.Error(err)
	}
	mustSuccess(proxy.Listen())
}
