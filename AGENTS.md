# ManboTv Codex Rules

本仓库中的 Codex 代理必须遵守以下规则。这些规则由 `.kimi` 与 `.opencode` 中的同名 skill 迁移而来，作为 Codex 的项目级约束。

## 沟通规则

- 默认使用中文沟通与汇报
- 仅在用户明确指定时使用其他语言

## Skill 使用规则

- 当任务涉及 ManboTV 或 MoonTV 的重构、前后端分离、Go 后端开发、接口分层、相关代码审查时，优先读取并遵守 [/.codex/skills/refactor-moontv/SKILL.md](/Users/Zhuanz1/Desktop/project/ManboTv/.codex/skills/refactor-moontv/SKILL.md)
- 如果该 skill 与更高优先级系统规则冲突，以系统规则为准；其余情况下必须执行

## 不可违反的工程红线

- 任何 Go、TS、JS、CSS 代码文件不得超过 800 行，HTML 除外
- 禁止直接写魔法数值；可复用或有语义的数值必须定义为具名常量
- 每完成一个重构模块，必须立即进行该模块的验证后再继续下一个模块

## 模块验证规则

- 默认验证流程为 `docker compose up -d --build`
- 使用 `curl` 覆盖该模块至少一组实际相关接口请求；按接口情况执行 `GET`、`POST`、`DELETE`
- 需要记录验证结果；若失败，先修复，再继续后续模块
- 若当前环境缺少 Docker、服务依赖或必要权限，必须明确说明阻塞原因，不能假装已验证
- 任何 bug 修复完成后，必须在 Docker 环境中重建并重启相关服务；默认执行 `docker compose up -d --build`
- bug 修复后的验证不能只停留在本地代码层，必须基于容器内实际运行的版本进行页面或接口验证
