#watcher
##作用
监控文件夹是否修改,并重启命令,可以自动发布,测试自动重启实例


```bash
##使用方式
-c 命令 支持绝对路径,支持path
-a 命令的参数,可以写多个
-d 要监控的文件夹 支持相对路径
-m 要过滤的扩展名单 不设置不过滤 php|js
 /home/wwwroot/clive/bin/watcher -c php -a /home/wwwroot/clive/wechat/cli.php -a task/http -a task -a worker001 -d ../include -d ../wechat -d ../reginx -d ../schconfig 
```