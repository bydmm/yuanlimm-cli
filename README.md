# 援力满满 - 二次元虚拟股市 - 高速许愿工具

## linux VPS快速启动指令

```bash
curl https://cdn.yuanlimm.com/release/yuanlimm-cli --output yuanlimm-cli \
&& sudo chmod +x ./yuanlimm-cli \
&& nohup ./yuanlimm-cli -a 钱包地址 --code 股票代码 -c 1 -w 英梨梨我喜欢你 &
```

## 检查上面的运行状态

```bash
tail -f nohup.out
```


## 用法

下载压缩包，解压缩获得以下文件

#### MACOS客户端

yuanlimm-cli_darwin_x64

#### Linux客户端

yuanlimm-cli_linux_x64

#### Windows客户端

yuanlimm-cli_win_x64.exe

## 参数

除了直接启动程序输入之外，还可以直接用flag启动程序

栗子:

```bash
cli_linux_x64 -a example_address -c 3 -code ERIRI
```

-a: 钱包地址, https://www.yuanlimm.com/#/profile

-c: 并发，建议和cpu数量相同

-code: 股票代码，比如英梨梨的地址：https://www.yuanlimm.com/#/stock/ERIRI，的股票代码是ERIRI，

## 开发白皮书

[白皮书](/wite_paper.md)

## 支持

QQ群：774800449
