package main

var defaultConfig = []byte(`
app {
  # 生产环境下的可执行文件。支持用“*”代替文件名的一部分，例如："tower-app-*.exe"
  exec : "tower-app-*.exe"

  # 开发环境下用“go run”命令运行的源文件，一般为“main.go”
  main : ""

  # 你的项目在本机运行的端口列表,可以用半角逗号分隔也可以用减号指定范围，也可以两种结合起来用，例如： "5001,5003,5050-5060"。如果为空，则代表不支持访问端口。
  port : "5001-5050"

  # 指定app端口的参数名，例如：webx.exe -p 8080 其中的“-p”就是。如果为空，则代表不支持访问端口。
  portParamName : "-p"

  # go build -o 命令生成的二进制文件保存位置
  buildDir : ""

  # 运行app所需的其它参数，例如：webx.exe -p 8080 -e 90 -d 100 其中的“-e 90 -d 100”就是(注意：内部用[单个]半角空格隔开)。
  params : ""
}

proxy {
  # 你的项目对外公开访问的端口
  port : "8080"

  # 代理引擎。支持fast和standard
  engine : "fast"
}

admin {
  password : ""
  ips : "127.0.0.1,::1"
}

watch {
  # 要监控更改的文件扩展名。多个扩展名时使用"|"隔开，例如：go|html
  fileExtension : "go"

  # 默认会自动监控上面main参数所指定的文件所在之文件夹，如果你还要监控其它文件夹，请在这里指定。如要指定多个文件夹路径，请用“|”分隔。
  otherDir : ""
  
  # 忽略的路径(正则表达式)，不填则不限制(排除某个完整的文件夹名请用“/文件夹名/”的格式)
  ignoredPath : ""
}

# 是否显示细节信息。如果设置为true，会自动将下面的logLevel设置为Debug
verbose : false

# 日志等级。支持的值有Debug/Info/Warn/Error/Fatal
logLevel : "Debug"

# 是否在控制台显示request日志
logRequest : true

# 是否自动删除以前的可执行文件
autoClear : true

# 是否离线模式(即开发模式)
offline : true

`)
