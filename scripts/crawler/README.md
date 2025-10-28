# 知乎回答爬虫工具

一个基于 Puppeteer 的知乎回答爬虫工具，可以将知乎回答内容转换为 Markdown 格式并保存到本地。

## 功能特性

- ✅ 使用 Puppeteer + Stealth 插件绕过反爬虫检测
- ✅ 自动展开折叠内容，获取完整回答
- ✅ 智能转换为 Markdown 格式
- ✅ 支持多种内容类型：
  - 图片（保留知乎原始链接）
  - 代码块（保留语法高亮标记）
  - 链接和链接卡片
  - 加粗、斜体、引用
  - 有序/无序列表
  - 标题（h2/h3）
- ✅ 自动关闭登录弹窗
- ✅ 文件名自动处理（替换 Windows 不支持的字符）
- ✅ 支持调试模式

## 环境要求

### 1. 安装 Node.js 依赖

```bash
cd scripts/crawler
pnpm install puppeteer puppeteer-extra puppeteer-extra-plugin-stealth
```

### 2. Linux 系统需要安装 Chromium

```bash
apt install -y chromium-browser
```

Windows/Mac 系统会自动下载 Chromium，无需手动安装。

## 使用方法

### 基本用法

```bash
node index.js <知乎回答URL>
```

### 示例

```bash
node index.js https://www.zhihu.com/question/123456/answer/789012
```

### 调试模式

调试模式会显示浏览器窗口，并在完成后保持打开状态，方便查看爬取过程：

```bash
node index.js <URL> --debug
```

## 输出说明

- 爬取的内容会保存在 `output/` 目录下
- 文件名格式：`标题 - 作者.md`
- Markdown 文件包含：
  - 完整的回答内容
  - 底部附带发布时间和地点
  - 原文链接

## 技术实现

### 内容提取流程

1. 启动无头浏览器（调试模式下显示窗口）
2. 加载目标页面并等待内容渲染
3. 自动关闭登录/模态框弹窗
4. 点击"展开"按钮显示完整内容
5. 提取标题、作者、时间等元数据
6. 将 HTML 内容转换为 Markdown
7. 处理图片、代码块等特殊元素
8. 保存到本地文件

### Markdown 转换规则

| HTML 元素 | Markdown 语法 |
|-----------|--------------|
| `<h2>`    | `## 标题`    |
| `<h3>`    | `### 标题`   |
| `<strong>` / `<b>` | `**粗体**` |
| `<em>` / `<i>` | `*斜体*` |
| `<code>`  | `` `代码` `` |
| `<blockquote>` | `> 引用` |
| `<li>`    | `- 列表项`   |
| `<img>`   | `![](url)`   |
| `<a>`     | `[文本](url)` |

### 图片处理

- 保留知乎原始图片链接（不下载到本地）
- 支持图片说明文字（居中显示）

### 文件名处理

自动替换 Windows 不支持的文件名字符：

```
/ \ ? : * " < > |  →  全角字符
```
