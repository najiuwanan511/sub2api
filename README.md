# Sub2API 定制版

基于 [Wei-Shaw/Sub2API](https://github.com/Wei-Shaw/sub2api) 的 AI API 网关。定制版换用了自定义后台界面，仪表盘显示平均 Token 速度，也可以在「账号管理 → 接入上级 Sub2API」填写上级站点和 API Key、同步模型并测试连通性。

安装和更新均使用[本仓库的发行版本](https://github.com/najiuwanan511/sub2api/releases)。下面两种部署方式任选其一，不要混用。

## Docker Compose（推荐）

需先安装 Docker 和 Compose v2，并确保有 `curl` 和 OpenSSL。安装会自动配置 PostgreSQL 和 Redis。

**一键安装**（在尚未创建 `sub2api-deploy` 的目录执行）：

```bash
mkdir sub2api-deploy && cd sub2api-deploy &&
  curl -fsSLo docker-deploy.sh https://raw.githubusercontent.com/najiuwanan511/sub2api/main/deploy/docker-deploy.sh &&
  bash docker-deploy.sh && docker compose up -d
```

**更新**（在部署目录执行，只更新 Sub2API）：

```bash
docker compose pull sub2api && docker compose up -d --no-deps sub2api
```

**卸载**（在部署目录执行，移除容器，保留数据和 `.env`）：

```bash
docker compose down
```

## Linux 二进制安装

需使用 systemd 的 Linux（amd64 或 arm64），并提前启动 PostgreSQL 15+ 和 Redis 7+。已用 root 登录且没有 `sudo` 时，去掉命令里的 `sudo`。

**一键安装**：

```bash
curl -fsSL https://raw.githubusercontent.com/najiuwanan511/sub2api/main/deploy/install.sh | sudo bash
```

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

默认访问 `http://服务器IP:8080`。二进制版首次访问时按向导配置；Docker 版可在部署目录运行 `docker compose logs sub2api` 查看首次登录信息。

**绑定域名后不想带端口？** DNS 解析不会把 HTTPS 的 `443` 自动转到 `8080`。在同一台服务器上用 Nginx 或 Caddy 配置 HTTPS 反向代理，将域名的 `443` 请求转发到 `http://127.0.0.1:8080`（或实际的宿主机端口）。确认 `https://你的域名` 能正常打开后，再关闭公网对 `8080` 的访问。

[更多部署说明](deploy/README.md) · [开源许可证（LGPL v3 或更高版本）](LICENSE)

版权所有 © 2026 Wesley Liddick。使用前请确认符合上游服务条款和适用法律。
