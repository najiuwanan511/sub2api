# Sub2API 定制版

基于 [Wei-Shaw/Sub2API](https://github.com/Wei-Shaw/sub2api) 的 AI API 网关。定制版换用了自定义后台界面，仪表盘显示平均 Token 速度，也可以在「账号管理 → 接入上级 Sub2API」填写上级站点和 API Key、同步模型并测试连通性。

安装和更新均使用[本仓库的发行版本](https://github.com/najiuwanan511/sub2api/releases)。**普通版无需 Docker。**

## Linux 二进制安装

需使用 Ubuntu / Debian（amd64 或 arm64，systemd），并提前启动 PostgreSQL 15+ 和 Redis 7+。已用 root 登录时，命令里的 `sudo` 可以去掉。

**一键安装（HTTPS，默认 443）**：把示例域名换成你的域名。DNS 指向服务器公网 IP，Cloudflare 先设为「仅 DNS」（灰云），并放行 TCP `80` 和 `443`。

```bash
curl -fsSL https://raw.githubusercontent.com/najiuwanan511/sub2api/main/deploy/install.sh | sudo bash -s -- install --domain api.example.com
```

脚本会安装 Caddy 自动签发证书，Sub2API 的 `8080` 只在本机监听。有旧实例或端口冲突时会退出，不会覆盖现有部署。没有域名时，可以省略 `--domain api.example.com`，保留原来的 HTTP 安装方式。

**更新**（也可在管理后台检查更新）：

```bash
curl -fsSL https://raw.githubusercontent.com/najiuwanan511/sub2api/main/deploy/install.sh | sudo bash -s -- upgrade
```

**卸载**（在交互式终端执行，按提示确认）：

```bash
curl -fsSL https://raw.githubusercontent.com/najiuwanan511/sub2api/main/deploy/install.sh | sudo bash -s -- uninstall
```

二进制卸载会删除 `/opt/sub2api`（含本地数据），并另行询问是否删除 `/etc/sub2api` 配置。请先备份。

## 安装后

首次配置不要暴露到公网。在自己的电脑运行下面的命令（把 `服务器IP` 换成服务器公网 IP），保持 SSH 窗口打开：

```bash
ssh -N -L 127.0.0.1:18080:127.0.0.1:8080 root@服务器IP
```

浏览器打开 `http://127.0.0.1:18080/setup` 配置数据库、Redis 和管理员账号。完成后用 `https://你的域名` 登录，不用加端口。若再开启 Cloudflare 橙云，SSL/TLS 模式选「完全（严格）」，不要选「灵活」。

## 重置卡与支付宝

- **重置卡**：管理员在「兑换码」选择「重置卡」，填写余额计费分组、返还比例、返还上限和有效天数。用户在「重置卡」页面预览后兑换。比如某分组本轮实扣 `$10`，返还 `30%`、上限 `$2`，就会收到 `$2` 限时额度。升级前的消费不计入；额度仅在指定分组使用，不可退款，不适用于需要预冻结常规余额的批量图片订单。
- **支付宝在线收款**：在管理后台支付设置启用在线支付并添加「支付宝官方」渠道，填写已签约应用的 App ID、应用私钥和支付宝公钥。通知与返回地址填写公网 HTTPS 域名。异步通知为 `https://你的域名/api/v1/payment/webhook/alipay`；使用 Cloudflare 时，不要让该路径要求登录或人机验证。启用后请先用小额订单验证支付和充值入账。

[更多部署说明](deploy/README.md) · [开源许可证（LGPL v3 或更高版本）](LICENSE)

版权所有 © 2026 Wesley Liddick。使用前请确认符合上游服务条款和适用法律。
