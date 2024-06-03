# 域名备案信息批量查询工具

这是一个使用 Go 语言编写的通过域名查询 ICP 备案号的工具，它支持**批量查询**。

## 安装

首先，你需要安装 Go 语言环境。然后，你可以通过以下命令来获取和安装这个工具：

```bash
go install github.com/y00k1sec/dtp
```

或者，你可以直接从 GitHub 仓库克隆并安装：

```bash
git clone https://github.com/y00k1sec/dtp.git
cd dtp
go install .
```

这将会在你的 `GOPATH/bin` 目录下生成一个名为 “dtp” 的二进制文件。

请注意，你需要将 `GOPATH/bin` 目录添加到你的 `PATH` 环境变量中，这样你就可以在任何位置运行你的程序了。你可以通过以下命令来添加 `GOPATH/bin` 到你的 `PATH`：

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

## 使用方法

你可以通过命令行参数或者标准输入来提供一个或多个域名。以下是一些使用示例：

```bash
# 查询单个域名
dtp -t example.com

# 从文件中读取域名列表进行查询
dtp -l domains.txt

# 从标准输入读取域名列表进行查询
echo -e "example.com\nexample.net" | dtp
```

## 参数说明

- `-t <target.xyz>`：指定要查询的目标域名。
- `-l <lists.txt>`：指定包含域名列表的文件。
- `-debug`：启用调试模式。
- `-json`：以 JSON 格式输出结果。
- `-r <number>`：设置重试次数。

## 输出结果

工具会输出以下信息：

- Input：输入的域名。
- Query URL：查询的 URL。
- ICP备案/许可证号：ICP 备案/许可证号。
- 审核通过日期：审核通过日期。
- 主办单位名称：主办单位名称。
- 主办单位性质：主办单位性质。
- ICP备案/许可证号：ICP 备案/许可证号。
- 网站域名：网站域名。
