"""
用于项目打包
"""
import os
import zipfile
import re
import shutil
import sys
import getopt

"""
定义应用环境变量
"""
# 排除的文件（目录也看成是文件），文件路径必须是相对pack.py文件的路径，比如main.go的路径就是写main.go，可以用正则表达式
FILE_EXCLUDES = [
    {
        # 这里写成一个字典而非字符串是为了以后复杂写法的兼容
        "path": ".+\\.go",
        "regex": True
    },
    {
        "path": ".+\\.py",
        # 默认是False
        "regex": True
    },
    {
        "path": "go.mod",
        "regex": False
    },
    {
        "path": "go.sum",
        "regex": False
    },
    {
        "path": ".gitignore",
        "regex": False
    },
    {
        "path": ".*\\.vscode.*",
        "regex": True
    },
    {
        "path": ".*\\.idea.*",
        "regex": True
    },
    {
        "path": ".*\\.git.*",
        "regex": True
    },
    {
        "path": ".*test",
        "regex": True
    },
    {
        "path": ".*output.*",
        "regex": True
    }
]

APP_ROOT_PATH = os.path.dirname(os.path.join(os.getcwd(), __file__))

APP_ZIP_FILE_NAME = os.path.split(APP_ROOT_PATH)[1] + ".zip"

# 这里实际上是要r:rebuild=，在后面使用时会加上
# 参数可以是-r True或--rebuild=True（如果这两个都写了，且不一致那属于用的人傻叉，只取一个即可）
REBUILD_OPTION = ["r", "rebuild"]

REBUILD_COMMAND = "go build service_mocker.go"


# 全局定义之间需要分隔两行，而局部定义则一行即可
def is_exclude(path: str) -> bool:
    flag: bool = False
    for val in FILE_EXCLUDES:
        if val['regex']:
            if re.fullmatch(val['path'], path) is not None:
                flag = True
        else:
            if path.endswith(val['path']):
                flag = True
    return flag


if __name__ == "__main__":
    # 看是否需要重新build项目（这个是针对go的，如果是其他语言的可以改下这部分代码）
    opts, args = getopt.getopt(sys.argv[1:], REBUILD_OPTION[0] + ":", REBUILD_OPTION[1] + "=")
    recompile_flag = str(False).lower()
    for val in opts:
        if val[0] == "-" + REBUILD_OPTION[0] or val[0] == "--" + REBUILD_OPTION[1]:
            if val[1] == str(True).lower() or val[1] == str(False).lower():
                recompile_flag = val[1]
    # print(recompile_flag)
    if recompile_flag == str(True).lower():
        print("开始重新编译项目")
        result = os.system(REBUILD_COMMAND)
        if result == 0:
            print("重新编译成功")
        else:
            print("重新编译失败")
            sys.exit(-1)

    # 创建/清空输出目录
    if not os.path.isdir(os.path.join(APP_ROOT_PATH, "output")):
        os.mkdir(os.path.join(APP_ROOT_PATH, "output"))
    else:
        # 先清空
        shutil.rmtree(os.path.join(APP_ROOT_PATH, "output"))
        os.mkdir(os.path.join(APP_ROOT_PATH, "output"))

    # 创建压缩文件
    zip_file = zipfile.ZipFile(os.path.join(APP_ROOT_PATH, "output", APP_ZIP_FILE_NAME), 'w', zipfile.ZIP_DEFLATED)
    # 类似返回了元组
    print("以下文件将被压缩到:output/" + APP_ZIP_FILE_NAME)
    for root, dirs, files in os.walk(APP_ROOT_PATH):
        # 打印文件
        for f in files:
            if not is_exclude(os.path.join(root, f)):
                tmp = os.path.join(root, f)
                # [2:]是获取子字符串，即从下标是2开始到最后
                tmp2 = tmp.replace(APP_ROOT_PATH, ".", 1)[2:]
                print(tmp2)
                # tmp是待压缩文件的绝对路径，因此需要替换掉前面的路径，否则压缩进去的目录是从绝对路径开始算，类似/home/silentdoer/...
                zip_file.write(tmp, tmp2)

    zip_file.close()
