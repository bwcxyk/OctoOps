# 版本发布流程

## 发布流程

- 从 `develop` 新建 `release/x.y.z` 分支，并修改 `package.json` 中的版本号，推送分支至远程仓库，并提交一个合入`develop`的 Pull Request 到仓库
- 确认无误后，合并分支入`develop`

合入 `develop` 后，仓库会触发 Github Action 合入 `main` 分支，并将版本号作为 `tag` 打在仓库上
