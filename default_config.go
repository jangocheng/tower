package main

var defaultConfig = []byte(`

# file name of the "go run" command
main              : "main.go"

# 你的项目在本机运行的端口列表,可以用半角逗号分隔也可以用减号指定范围，
# 也可以两种结合起来用，例如： "5001,5003,5050-5060"
app_port          : "5001-5010"

# go build -o 命令生成的二进制文件保存位置
app_buildDir      : ""
# 指定app端口的参数名，例如：webx.exe -p 8080 其中的“-p”就是。
app_portParamName : "-p"

# 运行app所需的其它参数，例如：webx.exe -p 8080 -e 90 -d 100 其中的“-e 90 -d 100”就是(注意：内部用[单个]半角空格隔开)。
app_runParams : ""

# 你的项目对外公开访问的端口
pxy_port          : "8080"

# file types to watch for changes in. use "|" to separate multiple types, for example, go|html
watch             : "go"

# 默认会自动监控上面main参数所指定的文件所在之文件夹，如果你还要监控其它文件夹，请在这里指定。
# 如要指定多个文件夹路径，请用“|”分隔。
watch_otherDir    : ""

# 忽略的路径(正则表达式)，不填则不限制(排除某个完整的文件夹名请用“/文件夹名/”的格式)
watch_ignoredPath : ""

# 是否离线模式(即开发模式)
offline_mode : "true"

# 是否在控制台显示request信息
log_request : "true"

`)
