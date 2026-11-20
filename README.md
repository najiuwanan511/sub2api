# Sub2API 定制版

基于 [Wei-Shaw/Sub2API](https://github.com/Wei-Shaw/sub2api) 的 AI API 网关。定制版换用了自定义后台界面，仪表盘显示平均 Token 速度，也可以在「账号管理 → 接入上级 Sub2API」填写上级站点和 API Key、同步模型并测试连通性。

安装和更新均使用[本仓库的发行版本](https://github.com/najiuwanan511/sub2api/releases)。下面两种部署方式任选其一，不要混用。

## Docker Compose（推荐）

需先安装 Docker 和 Compose v2，并确保有 `curl` 和 OpenSSL。安装会自动配置 PostgreSQL 和 Redis。

**一键安装**（在新目录执行）：有域名时，把下面的 `api.example.com` 换成自己的域名。先将 DNS 指向服务器，并开放 TCP `80` 和 `443`，确保这两个端口没被其他服务占用。

```bash
mkdir sub2api-deploy && cd sub2api-deploy &&
  curl -fsSLo docker-deploy.sh https://raw.githubusercontent.com/najiuwanan511/sub2api/main/deploy/docker-deploy.sh &&
  bash docker-deploy.sh --domain api.example.com && docker compose up -d
```

安装会自带 Caddy，自动申请和续期证书。访问 `https://你的域名` 即可，不用写端口，`8080` 仅在本机监听。没有域名时，把 `--domain api.example.com` 改成 `--http`，使用 `8080`。已有反向代理时，也请用 `--http` 安装，再把生成的 `.env` 中 `BIND_HOST` 改为 `127.0.0.1` 后启动，由现有代理接管 HTTPS（代理须与宿主机共用网络）。

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

Docker 版在部署目录运行 `docker compose logs sub2api` 查看首次登录信息；HTTPS 出现问题可看 `docker compose logs caddy`。二进制版在 `http://服务器IP:8080` 完成首次配置；如果需要 HTTPS，请用自己的反向代理。

[更多部署说明](deploy/README.md) · [开源许可证（LGPL v3 或更高版本）](LICENSE)

版权所有 © 2026 Wesley Liddick。使用前请确认符合上游服务条款和适用法律。
